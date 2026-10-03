package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/aicache"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/snips"
)

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

	// Bound storage while preserving request cancellation and trace context.
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

	// Bound summarization within the same request.
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
