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

	"github.com/voxire/onion/internal/modules/echo/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
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

const caretSystemPreamble = "You are Caret, an AI business assistant for rtylr."

var ErrInvalidRequestBody = errors.New("invalid request body")

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
}

type GenerateTextStreamRequest struct {
	Prompt string `json:"prompt"`
	// snake_case to match the frontend payload and onion's JSON convention.
	// camelCase aliases are accepted too so older clients keep working.
	SessionID     string `json:"session_id"`
	UserID        int64  `json:"user_id"`
	SystemContext string `json:"system_context"` // optional — app context injected as system message
	Mode          string `json:"mode"`           // optional — "analyst" routes through the NL-to-SQL layer

	// Accepted aliases (older camelCase clients).
	SessionIDAlt string `json:"sessionId"`
	UserIDAlt    int64  `json:"userId"`
}

type GenerateTextStreamResponse struct {
	Response   string `json:"response"`
	IsComplete bool   `json:"isComplete"`
	SessionID  string `json:"sessionId"`
}

func (h *GenerateTextStreamHandler) Validate(_ context.Context, req *GenerateTextStreamRequest) *errorModels.ServerError {
	if req.Prompt == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Prompt is required",
		}
	}

	if req.UserID <= 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "User ID is required",
		}
	}

	return nil
}

func (h *GenerateTextStreamHandler) Serve(_ context.Context, _ *GenerateTextStreamRequest) (GenerateTextStreamResponse, *errorModels.ServerError) {
	if !h.AiService.IsAvailable() {
		return GenerateTextStreamResponse{}, &errorModels.ServerError{
			Code:    http.StatusServiceUnavailable,
			Message: "AI service not available",
		}
	}

	return GenerateTextStreamResponse{}, nil
}

func (h *GenerateTextStreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Ctx(ctx)

	log.Info("Processing AI text stream request",
		logger.String("user_agent", r.UserAgent()),
		logger.String("remote_addr", r.RemoteAddr))

	req, err := h.parseAndValidateRequest(r)
	if err != nil {
		log.Error("Failed to parse request", logger.Error(err))
		http.Error(w, "Invalid request format", http.StatusBadRequest)

		return
	}

	if err := h.validateRequest(ctx, req); err != nil {
		log.Error("Request validation failed", logger.Error(err))
		http.Error(w, err.Message, err.Code)

		return
	}

	// Neutralize control markers and bound length before the input is stored or
	// folded into the prompt (prompt-injection defense).
	req.Prompt = sanitizeUserInput(req.Prompt, maxUserPromptLen)
	req.SystemContext = sanitizeUserInput(req.SystemContext, maxSystemContextLen)

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

	enhancedPrompt, contextErr := h.buildEnhancedPromptWithContext(ctx, req)
	if contextErr != nil {
		log.Error("Failed to get conversation context", logger.Error(contextErr))
		http.Error(w, "Failed to get context", http.StatusInternalServerError)

		return
	}

	// Analyst mode: turn the question into a validated, tenant-scoped query,
	// run it on the read replica, and ground the answer in the real rows. Falls
	// back to a normal chat answer if the data pipeline can't produce a result.
	if ParseMode(req.Mode) != ModeSupport {
		if grounding, ok := h.analystGrounding(ctx, req.Prompt, businessID, authUserID); ok {
			enhancedPrompt = grounding + "\n\n---\n\n" + enhancedPrompt
		}
	}

	log.Info("Starting AI text generation",
		logger.String("enhanced_prompt_length", strconv.Itoa(len(enhancedPrompt))))

	h.setupSSEHeaders(w)
	clientGone := h.monitorClientDisconnect(r)

	tokenChan, err := h.AiService.GenerateTextStream(ctx, enhancedPrompt)
	if err != nil {
		log.Error("Failed to generate text stream", logger.Error(err))
		h.writeErrorResponse(w, "Failed to generate response")

		return
	}

	h.streamTokensToClient(ctx, w, tokenChan, clientGone, req, businessID, authUserID, len(enhancedPrompt))
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

func (h *GenerateTextStreamHandler) parseAndValidateRequest(r *http.Request) (*GenerateTextStreamRequest, error) {
	var req GenerateTextStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, ErrInvalidRequestBody
	}

	// Accept camelCase aliases from older clients.
	if req.SessionID == "" {
		req.SessionID = req.SessionIDAlt
	}
	if req.UserID == 0 {
		req.UserID = req.UserIDAlt
	}

	return &req, nil
}

func (h *GenerateTextStreamHandler) validateRequest(ctx context.Context, req *GenerateTextStreamRequest) *errorModels.ServerError {
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

		sessionID, err := h.MemoryService.CreateSession(ctx, int(req.UserID), "Auto-generated session")
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

func (h *GenerateTextStreamHandler) buildEnhancedPromptWithContext(ctx context.Context, req *GenerateTextStreamRequest) (string, error) {
	logger.Ctx(ctx).Info("Building enhanced prompt with context",
		logger.String("session_id", req.SessionID),
		logger.Int("user_id", int(req.UserID)))

	enhancedPrompt := h.MemoryService.BuildEnhancedPrompt(ctx, req.SessionID, int(req.UserID), req.Prompt)

	if req.SystemContext == "" {
		return enhancedPrompt, nil
	}

	var sb strings.Builder
	sb.WriteString(caretSystemPreamble)
	sb.WriteString(" Today is ")
	sb.WriteString(time.Now().Format("2006-01-02"))
	sb.WriteString(".\nCurrent application context:\n")
	sb.WriteString(req.SystemContext)
	sb.WriteString("\n\n---\n\n")
	sb.WriteString(enhancedPrompt)

	return sb.String(), nil
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
	if _, writeErr := fmt.Fprintf(w, "event: error\ndata: %s\n\n", jsonEscape(errorMsg)); writeErr != nil {
		logger.Global().Error("Failed to write error to SSE response", logger.Error(writeErr))
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (h *GenerateTextStreamHandler) streamTokensToClient(ctx context.Context, w http.ResponseWriter, tokenChan <-chan snips.StreamToken, clientGone chan bool, req *GenerateTextStreamRequest, businessID, userID int64, promptChars int) {
	var fullResponse strings.Builder

	for {
		select {
		case token := <-tokenChan:
			if token.Error != "" {
				h.writeErrorResponse(w, token.Error)

				return
			}

			if err := h.processAndStreamToken(w, token, &fullResponse); err != nil {
				continue
			}

			if token.Done {
				if fullResponse.Len() > 0 {
					go h.storeAIResponse(ctx, req.SessionID, int(req.UserID), fullResponse.String())
				}

				// Meter usage against the quota. Prefer the provider's reported
				// token counts; fall back to a char/4 estimate when absent.
				tokens := token.PromptTokens + token.CompletionTokens
				if tokens == 0 {
					tokens = estimateTokens(promptChars) + estimateTokens(fullResponse.Len())
				}
				h.recordUsage(ctx, businessID, userID, tokens)

				return
			}

		case <-clientGone:
			return

		case <-ctx.Done():
			return
		}
	}
}

func (h *GenerateTextStreamHandler) processAndStreamToken(w http.ResponseWriter, token snips.StreamToken, fullResponse *strings.Builder) error {
	if token.Error != "" {
		return nil
	}

	data, err := json.Marshal(token)
	if err != nil {
		h.writeErrorResponse(w, "Failed to marshal token")

		return err
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		logger.Global().Error("Failed to write SSE data to response", logger.Error(err))

		return err
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	if token.Text != "" {
		fullResponse.WriteString(token.Text)
	}

	return nil
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

func jsonEscape(s string) string {
	return fmt.Sprintf("%q", s)
}
