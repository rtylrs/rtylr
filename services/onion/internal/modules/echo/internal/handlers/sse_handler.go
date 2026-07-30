package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

type MemoryService interface {
	CreateSession(ctx context.Context, userID int, title string) (string, error)
	StoreMessage(ctx context.Context, sessionID string, userID int, role, content string, tokens int) (any, error)
	SummarizeSession(ctx context.Context, sessionID string, userID int) error
	BuildEnhancedPrompt(ctx context.Context, sessionID string, userID int, currentRequest string) string
}

// QuotaService gates AI requests against the tenant's monthly token allowance
// and records consumption. Implemented by the aiquota module; kept as a local
// interface (primitive signatures) so echo doesn't depend on its internals.
// Authorize fails open (returns allowed) on error.
type QuotaService interface {
	Authorize(ctx context.Context, businessID, userID int64) (allowed bool, planSlug string, allowanceTokens, usedTokens, remainingTokens int64)
	RecordUsage(ctx context.Context, businessID, userID int64, tokens int)
}

const (
	caretSystemPreamble  = "You are Caret, an AI business assistant for rtylr."
	maxSessionTitleRunes = 80
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
	maxToolResultContinuations    = 2
	maxToolResultPromptBytes      = 12000
	unexpectedStreamClosureError  = "Response stream ended before completion"
)

var ErrInvalidRequestBody = errors.New("invalid request body")

type cacheWritePlan struct {
	Enabled     bool
	Key         string
	Mode        string
	Provider    string
	Policy      aicache.ExactPolicy
	StartedAt   time.Time
	PromptChars int
}

func GenerateTextStreamRoutes() []string {
	return []string{
		"/ai/generate-text-stream",
	}
}

type GenerateTextStreamHandler struct {
	AiService      *application.Service
	MemoryService  MemoryService
	QuotaService   QuotaService
	AnalystService AnalystService
	ToolExecutors  []ToolExecutor
}

type GenerateTextStreamRequest struct {
	Prompt string `json:"prompt"`
	// snake_case to match the frontend payload and onion's JSON convention.
	// camelCase aliases are accepted too so older clients keep working.
	SessionID     string         `json:"session_id"`
	UserID        int64          `json:"user_id"`
	SystemContext string         `json:"system_context"` // optional — app context injected as system message
	Mode          string         `json:"mode"`           // optional — "analyst" routes through the NL-to-SQL layer
	PageContext   map[string]any `json:"page_context,omitempty"`

	// Accepted aliases (older camelCase clients).
	SessionIDAlt   string         `json:"sessionId"`
	UserIDAlt      int64          `json:"userId"`
	PageContextAlt map[string]any `json:"pageContext,omitempty"`
}

type GenerateTextStreamResponse struct {
	Response   string `json:"response"`
	IsComplete bool   `json:"isComplete"`
	SessionID  string `json:"sessionId"`
}

func (h *GenerateTextStreamHandler) Validate(_ context.Context, req *GenerateTextStreamRequest) error {
	if req.Prompt == "" {
		return onekiterrors.New(http.StatusBadRequest, "Prompt is required")
	}

	if req.UserID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "User ID is required")
	}

	return nil
}

func (h *GenerateTextStreamHandler) Serve(_ context.Context, _ *GenerateTextStreamRequest) (GenerateTextStreamResponse, error) {
	if !h.AiService.IsAvailable() {
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

	// Neutralize control markers and bound length before the input is stored or
	// folded into the prompt (prompt-injection defense).
	req.Prompt = sanitizeUserInput(req.Prompt, maxUserPromptLen)
	req.SystemContext = buildRequestSystemContext(req.SystemContext, req.PageContext)

	if !h.AiService.IsAvailable() {
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
	// Trust the authenticated identity over the client-supplied user_id for the
	// session/memory owner too — prevents one user reading another's history.
	if authUserID > 0 {
		req.UserID = authUserID
	}
	if h.QuotaService != nil {
		allowed, plan, allowance, used, remaining := h.QuotaService.Authorize(ctx, businessID, authUserID)
		if !allowed {
			log.Info("AI request blocked by quota",
				logger.Any("owner_business_id", businessID), logger.String("plan", plan))
			h.writeQuotaExceeded(ctx, w, plan, allowance, used, remaining)
			return
		}
	}

	if setupErr := h.setupSessionAndStoreMessage(ctx, req); setupErr != nil {
		log.Error("Failed to setup session and store message", logger.Error(setupErr))
		http.Error(w, "Failed to setup session", http.StatusInternalServerError)

		return
	}

	mode := ParseMode(req.Mode)
	requestNow := time.Now()

	enhancedPrompt, contextErr := h.buildEnhancedPromptWithContext(ctx, req, mode, requestNow)
	if contextErr != nil {
		log.Error("Failed to get conversation context", logger.Error(contextErr))
		http.Error(w, "Failed to get context", http.StatusInternalServerError)

		return
	}

	var liveGrounding *analystGroundingResult
	groundBeforeCache := shouldGroundBeforeExactCache(mode, req.Prompt, req.PageContext)
	if groundBeforeCache {
		if grounding, ok := h.analystGrounding(ctx, req.Prompt, businessID, authUserID); ok {
			enhancedPrompt = grounding.Prompt + "\n\n---\n\n" + enhancedPrompt
			liveGrounding = grounding
		}
	}

	cachePlan := h.buildCacheWritePlan(ctx, mode, req, enhancedPrompt, businessID, authUserID, requestNow, liveGrounding)
	if groundBeforeCache && liveGrounding == nil {
		cachePlan = disableExactCachePlan(cachePlan, "live business request needs fresh grounding")
	}

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

	// Analyst grounding may call an internal model and the read replica. For
	// obvious live business questions we do it before the exact cache lookup so
	// the key includes a fingerprint of the rows. If a question only resolves
	// as live data after the lookup, keep the answer uncached rather than risk a
	// stale replay on the next request.
	if !groundBeforeCache && mode != ModeSupport {
		if grounding, ok := h.analystGrounding(ctx, req.Prompt, businessID, authUserID); ok {
			enhancedPrompt = grounding.Prompt + "\n\n---\n\n" + enhancedPrompt
			liveGrounding = grounding
			cachePlan = disableExactCachePlan(cachePlan, "live business grounding resolved after exact cache lookup")
			cachePlan.PromptChars = len(enhancedPrompt)
		}
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

	h.streamTokensToClient(ctx, w, tokenChan, clientGone, req, businessID, authUserID, cachePlan, liveGrounding, enhancedPrompt)
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
func (h *GenerateTextStreamHandler) recordUsage(ctx context.Context, businessID, userID int64, tokens int) {
	if h.QuotaService == nil || tokens <= 0 {
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
		h.QuotaService.RecordUsage(bg, businessID, userID, tokens)
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

	// Accept camelCase aliases from older clients.
	if req.SessionID == "" {
		req.SessionID = req.SessionIDAlt
	}
	if req.UserID == 0 {
		req.UserID = req.UserIDAlt
	}
	if req.PageContext == nil {
		req.PageContext = req.PageContextAlt
	}

	return &req, nil
}

func (h *GenerateTextStreamHandler) validateRequest(ctx context.Context, req *GenerateTextStreamRequest) error {
	return h.Validate(ctx, req)
}

func (h *GenerateTextStreamHandler) setupSessionAndStoreMessage(ctx context.Context, req *GenerateTextStreamRequest) error {
	log := logger.Ctx(ctx)

	log.Info("Setting up session and storing message",
		logger.String("session_id", req.SessionID),
		logger.Int("user_id", int(req.UserID)),
		logger.String("prompt_length", strconv.Itoa(len(req.Prompt))))

	if req.SessionID == "" {
		log.Info("Creating new session for user", logger.Int("user_id", int(req.UserID)))

		sessionID, err := h.MemoryService.CreateSession(ctx, int(req.UserID), sessionTitleFromPrompt(req.Prompt))
		if err != nil {
			log.Error("Failed to create session", logger.Error(err), logger.Int("user_id", int(req.UserID)))

			return fmt.Errorf("failed to create session: %w", err)
		}

		req.SessionID = sessionID
		log.Info("Session created successfully", logger.String("session_id", req.SessionID))
	} else {
		log.Info("Using existing session", logger.String("session_id", req.SessionID))
	}

	log.Info("Storing user message",
		logger.String("session_id", req.SessionID),
		logger.Int("user_id", int(req.UserID)),
		logger.String("role", "user"))

	_, err := h.MemoryService.StoreMessage(ctx, req.SessionID, int(req.UserID), "user", req.Prompt, len(req.Prompt)/4)
	if err != nil {
		log.Error("Failed to store message", logger.Error(err),
			logger.String("session_id", req.SessionID),
			logger.Int("user_id", int(req.UserID)))

		return fmt.Errorf("failed to store message: %w", err)
	}

	log.Info("Message stored successfully")

	return nil
}

func sessionTitleFromPrompt(prompt string) string {
	title := strings.Join(strings.Fields(prompt), " ")
	if title == "" {
		return "New chat"
	}

	runes := []rune(title)
	if len(runes) <= maxSessionTitleRunes {
		return title
	}
	return string(runes[:maxSessionTitleRunes]) + "..."
}

func (h *GenerateTextStreamHandler) buildEnhancedPromptWithContext(ctx context.Context, req *GenerateTextStreamRequest, mode CaretMode, now time.Time) (string, error) {
	logger.Ctx(ctx).Info("Building enhanced prompt with context",
		logger.String("session_id", req.SessionID),
		logger.Int("user_id", int(req.UserID)))

	conversationPrompt := h.MemoryService.BuildEnhancedPrompt(ctx, req.SessionID, int(req.UserID), req.Prompt)
	return buildCaretPrompt(caretPromptInput{
		Mode:               mode,
		ConversationPrompt: conversationPrompt,
		CurrentRequest:     req.Prompt,
		SystemContext:      req.SystemContext,
		Now:                now,
	}), nil
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

func (h *GenerateTextStreamHandler) buildCacheWritePlan(_ context.Context, mode CaretMode, req *GenerateTextStreamRequest, enhancedPrompt string, businessID, userID int64, now time.Time, groundings ...*analystGroundingResult) cacheWritePlan {
	policy := aicache.ExactPolicyForMode(string(mode))
	provider := ""
	if h.AiService != nil {
		provider = h.AiService.GetPrimaryProvider()
	}
	dataFreshness := buildCacheFreshness(req.PageContext)
	if len(groundings) > 0 {
		dataFreshness = combineCacheFreshness(dataFreshness, buildGroundingCacheFreshness(groundings[0]))
	}
	plan := cacheWritePlan{
		Enabled:     policy.Enabled,
		Mode:        string(mode),
		Provider:    provider,
		Policy:      policy,
		StartedAt:   time.Now(),
		PromptChars: len(enhancedPrompt),
	}
	if policy.Enabled {
		plan.Key = aicache.ExactKey(aicache.ExactKeyParts{
			BusinessID:    businessID,
			UserID:        userID,
			Mode:          string(mode),
			Provider:      provider,
			RequestDate:   now.UTC().Format("2006-01-02"),
			SystemContext: req.SystemContext,
			DataFreshness: dataFreshness,
			Prompt:        req.Prompt,
		})
	}
	return plan
}

func disableExactCachePlan(plan cacheWritePlan, reason string) cacheWritePlan {
	plan.Enabled = false
	plan.Key = ""
	plan.Policy.Enabled = false
	plan.Policy.Reason = reason
	return plan
}

func (h *GenerateTextStreamHandler) streamCachedResponse(ctx context.Context, w http.ResponseWriter, clientGone chan bool, req *GenerateTextStreamRequest, cached aicache.ExactResponse, plan cacheWritePlan, businessID, userID int64) {
	started := time.Now()
	messageID := newAssistantMessageID()
	if err := h.writeStreamPrelude(w, req.SessionID, messageID); err != nil {
		return
	}

	var fullResponse strings.Builder
	renderer := newStreamEventRenderer(w, flushResponse(w), messageID, &fullResponse, withStreamMode(ParseMode(plan.Mode)))

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

	if cached.Response != "" {
		go h.storeAIResponse(ctx, req.SessionID, int(req.UserID), cached.Response)
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

func (h *GenerateTextStreamHandler) streamTokensToClient(ctx context.Context, w http.ResponseWriter, tokenChan <-chan snips.StreamToken, clientGone chan bool, req *GenerateTextStreamRequest, businessID, userID int64, cachePlan cacheWritePlan, grounding *analystGroundingResult, basePrompt string) {
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
		withStreamMode(ParseMode(cachePlan.Mode)),
		withStreamToolExecution(ctx, h.ToolExecutors),
	)
	if grounding != nil {
		if err := h.writeAnalystToolEvents(w, messageID, grounding); err != nil {
			return
		}
	}

	currentPromptChars := cachePlan.PromptChars
	totalTokens := 0
	var finalToken snips.StreamToken
	for round := 0; ; round++ {
		beforeToolResults := renderer.ToolResultCount()
		beforeResponseChars := fullResponse.Len()
		token, ok := h.drainProviderStream(ctx, w, tokenChan, clientGone, renderer, cachePlan, businessID, userID, currentPromptChars)
		if !ok {
			partialResponseChars := fullResponse.Len() - beforeResponseChars + len(renderer.pending)
			if token.Error != "" || partialResponseChars > 0 {
				// Providers may consume the full prompt before a transport failure or
				// client interruption. They do not reliably report usage on failed
				// streams, so meter the observed partial attempt conservatively.
				totalTokens += estimateTokens(currentPromptChars) + estimateTokens(partialResponseChars)
				h.recordUsage(ctx, businessID, userID, totalTokens)
			}
			return
		}
		finalToken = token
		totalTokens += streamUsageTokens(token, currentPromptChars, fullResponse.Len()-beforeResponseChars)

		newToolResults := renderer.ToolResultsSince(beforeToolResults)
		if len(newToolResults) > 0 {
			// A live read can change independently of the request/page context, so
			// never replay an answer that was informed by a tool result.
			cachePlan = disableExactCachePlan(cachePlan, "live tool result informed the response")
		}
		if len(newToolResults) == 0 || round >= maxToolResultContinuations {
			if err := renderer.Finish(); err != nil {
				return
			}
			if fullResponse.Len() > 0 {
				go h.storeAIResponse(ctx, req.SessionID, int(req.UserID), fullResponse.String())
			}
			h.recordUsage(ctx, businessID, userID, totalTokens)
			h.writeExactCacheEntry(ctx, finalToken, fullResponse.String(), totalTokens, cachePlan, businessID, userID)
			return
		}

		currentPrompt := buildToolResultContinuationPrompt(basePrompt, fullResponse.String(), newToolResults)
		currentPromptChars = len(currentPrompt)
		nextTokenChan, err := h.AiService.GenerateTextStream(ctx, currentPrompt)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to generate tool-result continuation", logger.Error(err))
			if err := renderer.Finish(); err != nil {
				return
			}
			if fullResponse.Len() > 0 {
				go h.storeAIResponse(ctx, req.SessionID, int(req.UserID), fullResponse.String())
			}
			h.recordUsage(ctx, businessID, userID, totalTokens)
			return
		}
		tokenChan = nextTokenChan
	}
}

func (h *GenerateTextStreamHandler) drainProviderStream(ctx context.Context, w http.ResponseWriter, tokenChan <-chan snips.StreamToken, clientGone chan bool, renderer *streamEventRenderer, cachePlan cacheWritePlan, businessID, userID int64, promptChars int) (snips.StreamToken, bool) {
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

func buildToolResultContinuationPrompt(basePrompt, assistantSoFar string, results []streamToolExecution) string {
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		resultsJSON = []byte("[]")
	}
	if len(resultsJSON) > maxToolResultPromptBytes {
		resultsJSON = append(resultsJSON[:maxToolResultPromptBytes], []byte(`...truncated`)...)
	}

	var b strings.Builder
	b.WriteString(basePrompt)
	b.WriteString("\n\nTool results are now available for the same user request.\n")
	b.WriteString("Continue the same assistant answer using these results as evidence. Prefer each result summary for business language, then use raw output only for exact values. Do not repeat text already sent. Do not mention JSON, tool internals, rows, tenant, or implementation details unless the user asks.\n")
	if trimmed := strings.TrimSpace(assistantSoFar); trimmed != "" {
		b.WriteString("\nAssistant text already sent:\n")
		b.WriteString(trimmed)
		b.WriteString("\n")
	}
	b.WriteString("\nTool results:\n")
	b.Write(resultsJSON)
	b.WriteString("\n\nContinue now with the final user-facing answer.")
	return b.String()
}

func (h *GenerateTextStreamHandler) writeExactCacheEntry(ctx context.Context, token snips.StreamToken, response string, tokens int, plan cacheWritePlan, businessID, userID int64) {
	if !plan.Enabled || plan.Key == "" || response == "" {
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

	toolName := grounding.toolName()
	blockID := fmt.Sprintf("tool-%s-%d", strings.ReplaceAll(toolName, ".", "-"), time.Now().UnixNano())
	input := map[string]any{
		"question": grounding.Question,
		"spec":     json.RawMessage(grounding.SpecJSON),
	}
	output := map[string]any{
		"row_count":   grounding.RowCount,
		"truncated":   grounding.Truncated,
		"rows":        grounding.Rows,
		"duration_ms": grounding.DurationMs,
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

func (h *GenerateTextStreamHandler) storeAIResponse(ctx context.Context, sessionID string, userID int, response string) {
	// Runs in its own goroutine — a panic here must not crash the API process.
	defer func() {
		if r := recover(); r != nil {
			logger.Ctx(ctx).Error("panic in storeAIResponse goroutine", logger.Any("panic", r))
		}
	}()

	// Decouple from the request context: the SSE handler has already returned by
	// the time this runs, so the request ctx is canceled and the DB write +
	// cache invalidation would fail with "context canceled". Keep the logger
	// scope (trace ids) but a fresh, time-bounded cancellation.
	storeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()

	_, err := h.MemoryService.StoreMessage(storeCtx, sessionID, userID, "assistant", response, len(response)/4)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to store AI response",
			logger.Error(err),
			logger.String("session_id", sessionID),
			logger.Int("user_id", userID))
	}

	// Decouple from the request context — when the SSE handler returns,
	// the request context is canceled and the summarize call would
	// abort before it ever starts. We still keep the parent's logger
	// scope so trace ids survive.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Global().Error("panic in summarize goroutine", logger.Any("panic", r))
			}
		}()
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()

		if summarizeErr := h.MemoryService.SummarizeSession(bg, sessionID, userID); summarizeErr != nil {
			logger.Ctx(bg).Warn("Failed to summarize session",
				logger.Error(summarizeErr),
				logger.String("session_id", sessionID),
				logger.Int("user_id", userID))
		}
	}()
}
