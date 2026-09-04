package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/voxire/onion/internal/modules/echo/internal/application"
	"github.com/voxire/onion/internal/platform/aicache"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/snips"
)

// QuotaService reserves AI tokens against the tenant's monthly allowance and
// reconciles actual consumption afterwards. Implemented by the aiquota module;
// kept as a local interface (primitive signatures) so echo doesn't depend on
// its internals. ReserveQuota atomically debits an estimate before the provider
// call — parallel requests can no longer all pass an exhausted bucket.
type QuotaService interface {
	ReserveQuota(ctx context.Context, businessID, userID int64) (allowed bool, planSlug string, allowanceTokens, usedTokens, remainingTokens, reservedTokens int64)
	ReconcileUsage(ctx context.Context, businessID, userID, actualTokens, reservedTokens int64)
}

// quotaReservation settles exactly once, including early exits where no
// provider call occurs. A reservation must not become a permanent debit just
// because setup, cache lookup, or stream startup returned early.
type quotaReservation struct {
	once           sync.Once
	handler        *GenerateTextStreamHandler
	ctx            context.Context
	businessID     int64
	userID         int64
	reservedTokens int64
}

func (q *quotaReservation) settle(actualTokens int64) {
	if q == nil {
		return
	}
	q.once.Do(func() {
		q.handler.recordUsage(q.ctx, q.businessID, q.userID, actualTokens, q.reservedTokens)
	})
}

const (
	caretSmokeCheckReply = "Caret is working. What would you like help with?"
)

const (
	cachedReplayTinyResponseRunes = 12
	cachedReplayMinChunkRunes     = 8
	cachedReplayMinTailRunes      = 8
	cachedReplaySoftChunkRunes    = 12
	cachedReplayMaxChunkRunes     = 24
	cachedReplayMinDelay          = 95 * time.Millisecond
	cachedReplayMaxDelay          = 650 * time.Millisecond
	cachedReplayTargetDuration    = 8500 * time.Millisecond
	unexpectedStreamClosureError  = "Response stream ended before completion"
	sseHeartbeatInterval          = 15 * time.Second
)

var ErrInvalidRequestBody = errors.New("invalid request body")

func GenerateTextStreamRoutes() []string {
	return []string{
		"/ai/generate-text-stream",
	}
}

type GenerateTextStreamHandler struct {
	AiService      *application.Service
	MemoryService  application.MemoryService
	QuotaService   QuotaService
	AnalystService AnalystService
	ToolExecutors  []ToolExecutor
	summaryMu      sync.Mutex
	summaryRunning map[string]struct{}
}

type GenerateTextStreamRequest struct {
	Prompt        string         `json:"prompt"`
	SessionID     string         `json:"session_id"`
	UserID        int64          `json:"user_id"`
	SystemContext string         `json:"system_context"` // optional — app context injected as system message
	Mode          string         `json:"mode"`           // optional — "analyst" routes through the NL-to-SQL layer
	PageContext   map[string]any `json:"page_context,omitempty"`
}

// UnmarshalJSON keeps older Caret clients working while maintaining a single
// snake_case representation in the request type and in generated responses.
func (r *GenerateTextStreamRequest) UnmarshalJSON(data []byte) error {
	fields := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	decode := func(dst any, names ...string) error {
		for _, name := range names {
			raw, ok := fields[name]
			if !ok {
				continue
			}
			return json.Unmarshal(raw, dst)
		}
		return nil
	}

	var parsed GenerateTextStreamRequest
	if err := decode(&parsed.Prompt, "prompt"); err != nil {
		return err
	}
	if err := decode(&parsed.SessionID, "session_id", "sessionId"); err != nil {
		return err
	}
	if err := decode(&parsed.UserID, "user_id", "userId"); err != nil {
		return err
	}
	if err := decode(&parsed.SystemContext, "system_context"); err != nil {
		return err
	}
	if err := decode(&parsed.Mode, "mode"); err != nil {
		return err
	}
	if err := decode(&parsed.PageContext, "page_context", "pageContext"); err != nil {
		return err
	}
	*r = parsed
	return nil
}

type GenerateTextStreamResponse struct {
	Response   string `json:"response"`
	IsComplete bool   `json:"is_complete"`
	SessionID  string `json:"session_id"`
}

func (h *GenerateTextStreamHandler) Validate(_ context.Context, req *GenerateTextStreamRequest) error {
	if req == nil {
		return onekiterrors.New(http.StatusBadRequest, "Request is required")
	}
	if strings.TrimSpace(req.Prompt) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Prompt is required")
	}

	if req.UserID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "User ID is required")
	}

	return nil
}

func (h *GenerateTextStreamHandler) Serve(_ context.Context, _ *GenerateTextStreamRequest) (GenerateTextStreamResponse, error) {
	if h.AiService == nil || !h.AiService.IsAvailable() {
		return GenerateTextStreamResponse{}, onekiterrors.New(http.StatusServiceUnavailable, "AI service not available")
	}

	return GenerateTextStreamResponse{}, nil
}

func (h *GenerateTextStreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Ctx(ctx)

	log.Info("Processing AI text stream request",
		logger.String("user_agent", r.UserAgent()),
		logger.String("remote_addr", r.RemoteAddr))

	req, err := h.parseAndValidateRequest(w, r)
	if err != nil {
		log.Error("Failed to parse request", logger.Error(err))
		http.Error(w, "Invalid request format", http.StatusBadRequest)

		return
	}

	if err := h.validateRequest(ctx, req); err != nil {
		log.Error("Request validation failed", logger.Error(err))
		serverHttp.WriteOneKitError(ctx, w, err)

		return
	}

	if h.AiService == nil || !h.AiService.IsAvailable() {
		log.Warn("AI service not available")
		http.Error(w, "AI service not available", http.StatusServiceUnavailable)

		return
	}

	if h.MemoryService == nil {
		log.Error("Memory service not available")
		http.Error(w, "Memory service not available", http.StatusInternalServerError)

		return
	}

	// Quota gate. Use the authenticated identity (never the client-supplied
	// userId) so usage is metered against the real tenant.
	device := authPkg.DeviceCtx(ctx)
	var businessID, authUserID int64
	if device.User != nil {
		businessID = device.User.BusinessID
		authUserID = device.User.UserID
	}
	if authUserID <= 0 {
		if authUser, ok := authPkg.GetAuthUser(ctx); ok && authUser != nil {
			businessID = authUser.BusinessID
			authUserID = authUser.UserID
		}
	}
	if businessID <= 0 || authUserID <= 0 {
		log.Warn("AI request rejected without an authenticated business identity")
		http.Error(w, "Authentication is required", http.StatusUnauthorized)
		return
	}
	// Trust the authenticated identity over the client-supplied user_id for the
	// session/memory owner too — prevents one user reading another's history.
	if authUserID > 0 {
		req.UserID = authUserID
	}
	var reservation *quotaReservation
	if h.QuotaService != nil {
		allowed, plan, allowance, used, remaining, reserved := h.QuotaService.ReserveQuota(ctx, businessID, authUserID)
		if !allowed {
			log.Info("AI request blocked by quota",
				logger.Any("owner_business_id", businessID), logger.String("plan", plan))
			h.writeQuotaExceeded(ctx, w, plan, allowance, used, remaining)
			return
		}
		reservation = &quotaReservation{
			handler: h, ctx: ctx, businessID: businessID, userID: authUserID,
			reservedTokens: reserved,
		}
		defer reservation.settle(0)
	}

	requestNow := time.Now()
	prepared, setupErr := application.PrepareTurn(ctx, h.MemoryService, application.TurnRequest{
		Prompt:        req.Prompt,
		SessionID:     req.SessionID,
		UserID:        req.UserID,
		SystemContext: req.SystemContext,
		Mode:          req.Mode,
		PageContext:   req.PageContext,
	}, requestNow)
	if setupErr != nil {
		log.Error("Failed to setup session and store message", logger.Error(setupErr))
		http.Error(w, "Failed to setup session", http.StatusInternalServerError)

		return
	}
	req.Prompt = prepared.Request.Prompt
	req.SessionID = prepared.Request.SessionID
	req.UserID = prepared.Request.UserID
	req.SystemContext = prepared.Request.SystemContext
	req.PageContext = prepared.Request.PageContext

	// Do not spend a model call on connectivity checks. A content-free prompt
	// gives a generative model room to invent a plausible business report even
	// though no business evidence was requested.
	if isCaretSmokeCheck(req.Prompt) {
		h.setupSSEHeaders(w)
		h.streamDirectResponse(ctx, w, req, caretSmokeCheckReply)
		return
	}

	mode := prepared.Mode
	enhancedPrompt := prepared.EnhancedPrompt

	cachePlan := application.BuildCacheWritePlan(mode, application.TurnRequest{
		Prompt:        req.Prompt,
		SystemContext: req.SystemContext,
		PageContext:   req.PageContext,
	}, enhancedPrompt, businessID, authUserID, h.AiService.GetPrimaryProvider(), requestNow)

	h.setupSSEHeaders(w)
	clientGone := h.monitorClientDisconnect(r)

	if cachePlan.Enabled {
		if cached, ok := aicache.GetExact(ctx, cachePlan.Key); ok {
			log.Info("AI exact cache hit",
				logger.String("mode", cachePlan.Mode),
				logger.String("provider", cached.Provider),
				logger.String("cache_key", aicache.KeyPreview(cachePlan.Key)))
			h.streamCachedResponse(ctx, w, clientGone, req, cached, cachePlan, businessID, authUserID)
			return
		}
		aicache.Record(aicache.Event{
			Outcome:         aicache.OutcomeMiss,
			Mode:            cachePlan.Mode,
			Provider:        cachePlan.Provider,
			CacheKeyPreview: aicache.KeyPreview(cachePlan.Key),
			BusinessID:      businessID,
			UserID:          authUserID,
			PromptChars:     len(enhancedPrompt),
		})
	} else {
		aicache.Record(aicache.Event{
			Outcome:         aicache.OutcomeSkip,
			Mode:            cachePlan.Mode,
			Provider:        cachePlan.Provider,
			Reason:          cachePlan.Policy.Reason,
			BusinessID:      businessID,
			UserID:          authUserID,
			PromptChars:     len(enhancedPrompt),
			CacheKeyPreview: "",
		})
	}

	log.Info("Starting AI text generation",
		logger.String("enhanced_prompt_length", strconv.Itoa(len(enhancedPrompt))),
		logger.String("cache_mode", cachePlan.Mode),
		logger.String("cache_policy", cachePlan.Policy.Reason))

	tokenChan, err := h.AiService.GenerateTextStream(ctx, enhancedPrompt)
	if err != nil {
		log.Error("Failed to generate text stream", logger.Error(err))
		aicache.Record(aicache.Event{
			Outcome:     aicache.OutcomeError,
			Mode:        cachePlan.Mode,
			Provider:    cachePlan.Provider,
			Reason:      "provider stream failed to start",
			BusinessID:  businessID,
			UserID:      authUserID,
			PromptChars: len(enhancedPrompt),
		})
		h.writeErrorResponse(w, "Failed to generate response")

		return
	}

	h.streamTokensToClient(ctx, w, tokenChan, clientGone, req, businessID, authUserID, cachePlan, reservation)
}

func isCaretSmokeCheck(prompt string) bool {
	return application.ClassifyCaretRequest(prompt, "").IsSmokeCheck
}

func (h *GenerateTextStreamHandler) streamDirectResponse(ctx context.Context, w http.ResponseWriter, req *GenerateTextStreamRequest, response string) {
	messageID := newAssistantMessageID()
	if err := h.writeStreamPrelude(w, req.SessionID, messageID); err != nil {
		return
	}

	var fullResponse strings.Builder
	renderer := newStreamEventRenderer(w, flushResponse(w), messageID, &fullResponse, withStreamMode(application.ParseMode(req.Mode)))
	if err := renderer.AppendText(response); err != nil {
		return
	}
	if err := h.storeAIResponse(ctx, req.SessionID, int(req.UserID), fullResponse.String()); err != nil {
		h.writeErrorResponse(w, "Failed to persist response")
		return
	}
	if err := renderer.Finish(); err != nil {
		return
	}
}

// writeQuotaExceeded reports that the tenant has spent its monthly AI allowance.
// 402 Payment Required signals "upgrade to continue".
type quotaExceededResponse struct {
	Error           string `json:"error"`
	Message         string `json:"message"`
	Plan            string `json:"plan"`
	AllowanceTokens int64  `json:"allowance_tokens"`
	UsedTokens      int64  `json:"used_tokens"`
	RemainingTokens int64  `json:"remaining_tokens"`
}

func (h *GenerateTextStreamHandler) writeQuotaExceeded(ctx context.Context, w http.ResponseWriter, plan string, allowance, used, remaining int64) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusPaymentRequired)
	if err := json.NewEncoder(w).Encode(quotaExceededResponse{
		Error:           "ai_quota_exceeded",
		Message:         "Your team's monthly AI limit has been reached. Upgrade your plan to keep using Caret.",
		Plan:            plan,
		AllowanceTokens: allowance,
		UsedTokens:      used,
		RemainingTokens: remaining,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to write quota exceeded response", logger.Error(err))
	}
}

// recordUsage meters consumed tokens against the tenant's quota. It runs detached
// from the request context (which is canceled when the SSE handler returns) but
// time-bounded.
func (h *GenerateTextStreamHandler) recordUsage(ctx context.Context, businessID, userID, tokens, reservedTokens int64) {
	if h.QuotaService == nil || (tokens <= 0 && reservedTokens <= 0) {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Ctx(ctx).Error("panic in recordUsage goroutine", logger.Any("panic", r))
			}
		}()
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		h.QuotaService.ReconcileUsage(bg, businessID, userID, tokens, reservedTokens)
	}()
}

// estimateTokens is the fallback when a provider doesn't report usage
// (~4 chars/token).
func estimateTokens(chars int) int {
	if chars <= 0 {
		return 0
	}
	return chars / 4
}

func (h *GenerateTextStreamHandler) parseAndValidateRequest(w http.ResponseWriter, r *http.Request) (*GenerateTextStreamRequest, error) {
	var req GenerateTextStreamRequest
	if err := serverHttp.DecodeJSONRequestBody(w, r, &req); err != nil {
		return nil, ErrInvalidRequestBody
	}

	return &req, nil
}

func (h *GenerateTextStreamHandler) validateRequest(ctx context.Context, req *GenerateTextStreamRequest) error {
	return h.Validate(ctx, req)
}

func (h *GenerateTextStreamHandler) setupSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	if err := http.NewResponseController(w).SetWriteDeadline(time.Time{}); err != nil {
		logger.Global().Warn("could not clear SSE write deadline", logger.Error(err))
	}
}

func (h *GenerateTextStreamHandler) monitorClientDisconnect(r *http.Request) chan bool {
	// Buffer of 1 so the goroutine never blocks if the stream loop
	// already exited via another case (tokenChan close / ctx.Done) —
	// otherwise the send below leaked one goroutine per SSE request.
	clientGone := make(chan bool, 1)

	go func() {
		<-r.Context().Done()
		select {
		case clientGone <- true:
		default:
		}
	}()

	return clientGone
}

func (h *GenerateTextStreamHandler) writeErrorResponse(w http.ResponseWriter, errorMsg string) {
	if writeErr := WriteEvent(w, flushResponse(w), StreamEvent{
		Type:        EventError,
		Error:       errorMsg,
		Recoverable: false,
	}); writeErr != nil {
		logger.Global().Error("Failed to write error to SSE response", logger.Error(writeErr))
	}
}

func (h *GenerateTextStreamHandler) streamCachedResponse(ctx context.Context, w http.ResponseWriter, clientGone chan bool, req *GenerateTextStreamRequest, cached aicache.ExactResponse, plan cacheWritePlan, businessID, userID int64) {
	started := time.Now()
	messageID := newAssistantMessageID()
	if err := h.writeStreamPrelude(w, req.SessionID, messageID); err != nil {
		return
	}

	var fullResponse strings.Builder
	renderer := newStreamEventRenderer(w, flushResponse(w), messageID, &fullResponse, withStreamMode(application.ParseMode(plan.Mode)))

	chunks := cachedReplayChunks(cached.Response)
	replayDelay := cachedReplayDelay(len(chunks))
	for i, chunk := range chunks {
		select {
		case <-clientGone:
			return
		case <-ctx.Done():
			return
		default:
		}

		if err := renderer.AppendText(chunk); err != nil {
			return
		}
		if i < len(chunks)-1 && !waitCachedReplay(ctx, clientGone, replayDelay) {
			return
		}
	}

	if err := renderer.Finish(); err != nil {
		return
	}

	aicache.Record(aicache.Event{
		Outcome:          aicache.OutcomeHit,
		Mode:             plan.Mode,
		Provider:         cached.Provider,
		CacheKeyPreview:  aicache.KeyPreview(plan.Key),
		BusinessID:       businessID,
		UserID:           userID,
		PromptChars:      plan.PromptChars,
		ResponseChars:    len(cached.Response),
		EstimatedSavings: cached.TotalTokens(),
		LatencyMs:        time.Since(started).Milliseconds(),
	})
}

func cachedReplayChunks(text string) []string {
	runes := []rune(text)
	if len(runes) == 0 {
		return nil
	}

	chunks := make([]string, 0, (len(runes)/cachedReplaySoftChunkRunes)+1)
	for start := 0; start < len(runes); {
		end := cachedReplayChunkEnd(runes, start)
		chunks = append(chunks, string(runes[start:end]))
		start = end
	}
	return chunks
}

func cachedReplayChunkEnd(runes []rune, start int) int {
	remaining := len(runes) - start
	if remaining <= cachedReplayTinyResponseRunes {
		return len(runes)
	}

	soft := start + cachedReplaySoftChunkRunes
	maxEnd := start + cachedReplayMaxChunkRunes
	if maxEnd > len(runes) {
		maxEnd = len(runes)
	}
	if soft > maxEnd {
		soft = maxEnd
	}
	if soft >= len(runes) {
		return cachedReplayShortSplitEnd(runes, start)
	}

	for i := soft; i < maxEnd; i++ {
		if cachedReplayBoundary(runes[i]) {
			end := consumeReplayBoundary(runes, i+1)
			if cachedReplayUsableEnd(runes, start, end) {
				return end
			}
		}
	}
	for i := soft; i > start; i-- {
		if cachedReplayBoundary(runes[i]) {
			end := consumeReplayBoundary(runes, i+1)
			if cachedReplayUsableEnd(runes, start, end) {
				return end
			}
		}
	}
	if maxEnd >= len(runes) {
		return cachedReplayShortSplitEnd(runes, start)
	}
	return maxEnd
}

func cachedReplayBoundary(r rune) bool {
	if unicode.IsSpace(r) {
		return true
	}
	return strings.ContainsRune(".,;:!?)]}", r)
}

func consumeReplayBoundary(runes []rune, end int) int {
	for end < len(runes) && unicode.IsSpace(runes[end]) {
		end++
	}
	return end
}

func cachedReplayUsableEnd(runes []rune, start, end int) bool {
	return end < len(runes) &&
		end-start >= cachedReplayMinChunkRunes &&
		len(runes)-end >= cachedReplayMinTailRunes
}

func cachedReplayShortSplitEnd(runes []rune, start int) int {
	remaining := len(runes) - start
	if remaining <= cachedReplayTinyResponseRunes {
		return len(runes)
	}

	end := start + remaining/2
	if end-start < cachedReplayMinChunkRunes {
		end = start + cachedReplayMinChunkRunes
	}
	if len(runes)-end < cachedReplayMinTailRunes {
		return len(runes)
	}

	for i := end; i > start; i-- {
		if cachedReplayBoundary(runes[i]) {
			if candidate := consumeReplayBoundary(runes, i+1); cachedReplayUsableEnd(runes, start, candidate) {
				return candidate
			}
		}
	}
	for i := end; i < len(runes); i++ {
		if cachedReplayBoundary(runes[i]) {
			if candidate := consumeReplayBoundary(runes, i+1); cachedReplayUsableEnd(runes, start, candidate) {
				return candidate
			}
		}
	}
	if cachedReplayUsableEnd(runes, start, end) {
		return end
	}
	return len(runes)
}

func cachedReplayDelay(chunkCount int) time.Duration {
	if chunkCount <= 1 {
		return 0
	}
	delay := cachedReplayTargetDuration / time.Duration(chunkCount)
	if delay < cachedReplayMinDelay {
		return cachedReplayMinDelay
	}
	if delay > cachedReplayMaxDelay {
		return cachedReplayMaxDelay
	}
	return delay
}

func waitCachedReplay(ctx context.Context, clientGone chan bool, delay time.Duration) bool {
	if delay <= 0 {
		return true
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-clientGone:
		return false
	case <-ctx.Done():
		return false
	}
}

func (h *GenerateTextStreamHandler) streamTokensToClient(ctx context.Context, w http.ResponseWriter, tokenChan <-chan snips.StreamToken, clientGone chan bool, req *GenerateTextStreamRequest, businessID, userID int64, cachePlan cacheWritePlan, reservation *quotaReservation) {
	settleUsage := func(actualTokens int64) {
		if reservation != nil {
			reservation.settle(actualTokens)
			return
		}
		h.recordUsage(ctx, businessID, userID, actualTokens, 0)
	}
	var fullResponse strings.Builder
	messageID := newAssistantMessageID()
	if err := h.writeStreamPrelude(w, req.SessionID, messageID); err != nil {
		return
	}
	renderer := newStreamEventRenderer(
		w,
		flushResponse(w),
		messageID,
		&fullResponse,
		withStreamMode(application.ParseMode(cachePlan.Mode)),
		withStreamToolExecution(ctx, h.ToolExecutors),
		withStreamToolBudget(cachePlan.ToolBudget),
		withStreamPlannerOnly(cachePlan.ProviderPlannerOnly),
		withStreamPageContext(req.PageContext),
	)

	promptChars := cachePlan.PromptChars
	token, ok := h.drainProviderStream(ctx, w, tokenChan, clientGone, renderer, cachePlan, businessID, userID, promptChars)
	if !ok {
		partialResponseChars := fullResponse.Len() + len(renderer.pending)
		if token.Error != "" || partialResponseChars > 0 {
			// Providers may consume the full prompt before a transport failure or
			// client interruption. They do not reliably report usage on failed
			// streams, so meter the observed partial attempt conservatively.
			settleUsage(int64(estimateTokens(promptChars) + estimateTokens(partialResponseChars)))
		}
		return
	}

	totalTokens := streamUsageTokens(token, promptChars, fullResponse.Len())
	if renderer.ToolResultCount() > 0 {
		// A live read can change independently of the request/page context, so
		// never replay an answer that was informed by a tool result.
		cachePlan = application.DisableExactCachePlan(cachePlan, "live tool result informed the response")
	}
	if strings.TrimSpace(fullResponse.String()) == "" && renderer.PendingApprovalCount() == 0 {
		if fallback := fallbackCaretResponse(nil, renderer.toolResults); fallback != "" {
			if err := renderer.AppendServerText(fallback); err != nil {
				settleUsage(int64(totalTokens))
				return
			}
		}
	}
	if fullResponse.Len() > 0 {
		if err := h.storeAIResponse(ctx, req.SessionID, int(req.UserID), fullResponse.String()); err != nil {
			h.writeErrorResponse(w, "Failed to persist response")
			settleUsage(int64(totalTokens))
			return
		}
	}
	if err := renderer.Finish(); err != nil {
		return
	}
	settleUsage(int64(totalTokens))
	h.writeExactCacheEntry(ctx, token, fullResponse.String(), totalTokens, cachePlan, businessID, userID)
}

func fallbackCaretResponse(grounding *analystGroundingResult, results []streamToolExecution) string {
	summaries := make([]string, 0, len(results))
	for _, result := range results {
		if summary := strings.TrimSpace(result.Summary); summary != "" && result.Status == "success" {
			summaries = append(summaries, summary)
		}
	}
	if len(summaries) > 0 {
		return "I checked your business data.\n\n" + strings.Join(summaries, "\n")
	}
	if grounding != nil && grounding.ToolName == deferredAnalyticsToolName {
		return "That request is too large to run as a live chat check. I can narrow it to a date range, summary, or specific records."
	}
	if grounding != nil {
		return "I checked the available business data, but it did not contain enough visible evidence for a reliable answer."
	}
	return "I could not produce a reliable answer from the available business context."
}

func (h *GenerateTextStreamHandler) drainProviderStream(ctx context.Context, w http.ResponseWriter, tokenChan <-chan snips.StreamToken, clientGone chan bool, renderer *streamEventRenderer, cachePlan cacheWritePlan, businessID, userID int64, promptChars int) (snips.StreamToken, bool) {
	return h.drainProviderStreamWithHeartbeat(ctx, w, tokenChan, clientGone, renderer, cachePlan, businessID, userID, promptChars, sseHeartbeatInterval)
}

func (h *GenerateTextStreamHandler) drainProviderStreamWithHeartbeat(ctx context.Context, w http.ResponseWriter, tokenChan <-chan snips.StreamToken, clientGone chan bool, renderer *streamEventRenderer, cachePlan cacheWritePlan, businessID, userID int64, promptChars int, heartbeatInterval time.Duration) (snips.StreamToken, bool) {
	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case token, ok := <-tokenChan:
			if !ok {
				token := snips.StreamToken{Error: unexpectedStreamClosureError, Done: true}
				h.writeErrorResponse(w, token.Error)
				aicache.Record(aicache.Event{
					Outcome:         aicache.OutcomeError,
					Mode:            cachePlan.Mode,
					Provider:        cachePlan.Provider,
					Reason:          "provider stream closed before completion",
					CacheKeyPreview: aicache.KeyPreview(cachePlan.Key),
					BusinessID:      businessID,
					UserID:          userID,
					PromptChars:     promptChars,
					LatencyMs:       time.Since(cachePlan.StartedAt).Milliseconds(),
				})
				return token, false
			}

			if token.Error != "" {
				h.writeErrorResponse(w, token.Error)
				aicache.Record(aicache.Event{
					Outcome:         aicache.OutcomeError,
					Mode:            cachePlan.Mode,
					Provider:        token.Provider,
					Reason:          token.Error,
					CacheKeyPreview: aicache.KeyPreview(cachePlan.Key),
					BusinessID:      businessID,
					UserID:          userID,
					PromptChars:     promptChars,
					LatencyMs:       time.Since(cachePlan.StartedAt).Milliseconds(),
				})

				return token, false
			}

			if token.Text != "" {
				if err := renderer.AppendText(token.Text); err != nil {
					return token, false
				}
			}

			if token.Done {
				if err := renderer.FlushPending(); err != nil {
					return token, false
				}
				return token, true
			}

		case <-clientGone:
			return snips.StreamToken{}, false

		case <-ctx.Done():
			return snips.StreamToken{}, false

		case <-heartbeat.C:
			if err := WriteSSEHeartbeat(w, flushResponse(w)); err != nil {
				return snips.StreamToken{}, false
			}
		}
	}
}

func streamUsageTokens(token snips.StreamToken, promptChars, responseChars int) int {
	tokens := token.PromptTokens + token.CompletionTokens
	if tokens > 0 {
		return tokens
	}
	return estimateTokens(promptChars) + estimateTokens(responseChars)
}

func (h *GenerateTextStreamHandler) writeExactCacheEntry(ctx context.Context, token snips.StreamToken, response string, tokens int, plan cacheWritePlan, businessID, userID int64) {
	if !plan.Enabled || plan.Key == "" || response == "" {
		return
	}
	quality := assessCaretAnswer(response)
	if !quality.Cacheable {
		logger.Ctx(ctx).Warn("skipping exact cache write after answer quality gate",
			logger.Any("quality_score", quality.Score),
			logger.Any("findings", quality.Findings),
		)
		return
	}

	provider := token.Provider
	if provider == "" {
		provider = plan.Provider
	}

	writeCtx := context.WithoutCancel(ctx)
	aicache.PutExact(writeCtx, plan.Key, aicache.ExactResponse{
		Response:         response,
		Provider:         provider,
		Mode:             plan.Mode,
		PromptChars:      plan.PromptChars,
		ResponseChars:    len(response),
		PromptTokens:     token.PromptTokens,
		CompletionTokens: token.CompletionTokens,
		CreatedAt:        time.Now().UTC(),
	}, plan.Policy.TTL)

	aicache.Record(aicache.Event{
		Outcome:         aicache.OutcomeWrite,
		Mode:            plan.Mode,
		Provider:        provider,
		CacheKeyPreview: aicache.KeyPreview(plan.Key),
		BusinessID:      businessID,
		UserID:          userID,
		PromptChars:     plan.PromptChars,
		ResponseChars:   len(response),
		ProviderTokens:  tokens,
		LatencyMs:       time.Since(plan.StartedAt).Milliseconds(),
	})
}

func (h *GenerateTextStreamHandler) writeStreamPrelude(w http.ResponseWriter, sessionID, messageID string) error {
	if sessionID != "" {
		if err := WriteEvent(w, flushResponse(w), StreamEvent{Type: EventSession, SessionID: sessionID}); err != nil {
			return err
		}
	}
	return WriteEvent(w, flushResponse(w), StreamEvent{Type: EventMessageStart, MessageID: messageID, Role: "assistant"})
}

func (h *GenerateTextStreamHandler) writeAnalystToolEvents(w http.ResponseWriter, messageID string, grounding *analystGroundingResult) error {
	if grounding == nil {
		return nil
	}

	toolName := "analytics.query"
	if strings.TrimSpace(grounding.ToolName) != "" {
		toolName = strings.TrimSpace(grounding.ToolName)
	}
	blockID := fmt.Sprintf("tool-%s-%d", strings.ReplaceAll(toolName, ".", "-"), time.Now().UnixNano())
	input := map[string]any{
		"question": grounding.Question,
		"spec":     json.RawMessage(grounding.SpecJSON),
	}
	output := map[string]any{
		"row_count":       grounding.RowCount,
		"truncated":       grounding.Truncated,
		"rows":            grounding.Rows,
		"duration_ms":     grounding.DurationMs,
		"evidence_source": grounding.EvidenceSource,
		"generated_at":    grounding.GeneratedAt,
	}
	summary := summarizeToolResult(toolName, input, output, "success", "")
	if err := WriteEvent(w, flushResponse(w), StreamEvent{
		Type:             EventToolCall,
		MessageID:        messageID,
		BlockID:          blockID,
		Name:             toolName,
		Input:            input,
		RequiresApproval: false,
	}); err != nil {
		return err
	}
	return WriteEvent(w, flushResponse(w), StreamEvent{
		Type:       EventToolResult,
		MessageID:  messageID,
		BlockID:    blockID,
		Status:     "success",
		Summary:    summary,
		Output:     output,
		DurationMs: grounding.DurationMs,
	})
}

func flushResponse(w http.ResponseWriter) func() {
	return func() {
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}

// storeAIResponse persists the completed assistant turn before the stream's
// done event is acknowledged. The request context is still live at each call
// site; callers that need detached work should derive that context explicitly.
func (h *GenerateTextStreamHandler) storeAIResponse(ctx context.Context, sessionID string, userID int, response string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("store AI response panic: %v", r)
			logger.Ctx(ctx).Error("panic while storing AI response", logger.Any("panic", r))
		}
	}()

	// Keep the logger scope (trace ids) but apply a fresh, time-bounded
	// cancellation to the detached context.
	storeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err = h.MemoryService.StoreMessage(storeCtx, sessionID, userID, "assistant", response, len(response)/4)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to store AI response",
			logger.Error(err),
			logger.String("session_id", sessionID),
			logger.Int("user_id", userID))
		return err
	}
	if !h.claimSessionSummary(sessionID) {
		return nil
	}
	defer h.releaseSessionSummary(sessionID)

	// Keep the detached context and trace ids while bounding summarization.
	bg, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if summarizeErr := h.MemoryService.SummarizeSession(bg, sessionID, userID); summarizeErr != nil {
		logger.Ctx(bg).Warn("Failed to summarize session",
			logger.Error(summarizeErr),
			logger.String("session_id", sessionID),
			logger.Int("user_id", userID))
	}
	return nil
}

func (h *GenerateTextStreamHandler) claimSessionSummary(sessionID string) bool {
	if h == nil || strings.TrimSpace(sessionID) == "" {
		return false
	}
	h.summaryMu.Lock()
	defer h.summaryMu.Unlock()
	if h.summaryRunning == nil {
		h.summaryRunning = make(map[string]struct{})
	}
	if _, exists := h.summaryRunning[sessionID]; exists {
		return false
	}
	h.summaryRunning[sessionID] = struct{}{}
	return true
}

func (h *GenerateTextStreamHandler) releaseSessionSummary(sessionID string) {
	h.summaryMu.Lock()
	defer h.summaryMu.Unlock()
	delete(h.summaryRunning, sessionID)
}
