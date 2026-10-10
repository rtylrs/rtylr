package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/caret/internal/application"
	"github.com/voxire/onion/internal/platform/aicache"
	"github.com/voxire/onion/internal/platform/logger"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

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

	if err := renderer.FlushPending(); err != nil {
		return
	}
	if err := h.storeAIResponse(ctx, w, req.SessionID, int(req.UserID), fullResponse.String()); err != nil {
		h.writeErrorResponse(w, "Failed to persist response")
		return
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
