package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/caret/internal/application"
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
func (h *GenerateTextStreamHandler) storeAIResponse(ctx context.Context, w http.ResponseWriter, sessionID string, userID int, response string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("store AI response panic: %v", r)
			logger.Ctx(ctx).Error("panic while storing AI response", logger.Any("panic", r))
		}
	}()

	// Bound storage while preserving request cancellation and trace context.
	storeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if recorder, ok := w.(*transcriptWriter); ok {
		if store, ok := h.MemoryService.(application.TranscriptMessageService); ok {
			transcript := recorder.Transcript()
			blocks, marshalErr := json.Marshal(transcript.Blocks)
			if marshalErr != nil {
				return marshalErr
			}
			_, err = store.StoreMessageWithTranscript(storeCtx, sessionID, userID, "assistant", response, len(response)/4, blocks, transcript.ResponseMs)
		} else {
			_, err = h.MemoryService.StoreMessage(storeCtx, sessionID, userID, "assistant", response, len(response)/4)
		}
	} else {
		_, err = h.MemoryService.StoreMessage(storeCtx, sessionID, userID, "assistant", response, len(response)/4)
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to store AI response",
			logger.Error(err),
			logger.String("session_id", sessionID),
			logger.Int("user_id", userID))
		return err
	}
	if titler, ok := h.MemoryService.(application.ChatTitler); ok && h.AiService != nil && h.AiService.IsAvailable() {
		go h.titleChat(context.WithoutCancel(ctx), titler, sessionID, userID)
	}
	if !h.claimSessionSummary(sessionID) {
		return nil
	}
	if compactor, ok := h.MemoryService.(application.ConversationCompactor); ok && h.AiService != nil && h.AiService.IsAvailable() {
		// Compaction calls the model, so it runs after the reply instead of
		// holding the done event. The claim keeps one compaction per chat.
		go func() {
			defer h.releaseSessionSummary(sessionID)
			h.compactConversation(context.WithoutCancel(ctx), compactor, sessionID, userID)
		}()
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

const compactionTimeout = 90 * time.Second

// compactConversation folds chat history that has outgrown the context window
// into the chat's rolling summary. Failure only means the next turn shows
// fewer old messages; it never affects the reply already sent.
func (h *GenerateTextStreamHandler) compactConversation(ctx context.Context, compactor application.ConversationCompactor, sessionID string, userID int) {
	defer func() {
		if r := recover(); r != nil {
			logger.Ctx(ctx).Error("panic while compacting chat", logger.Any("panic", r))
		}
	}()
	ctx, cancel := context.WithTimeout(ctx, compactionTimeout)
	defer cancel()
	log := logger.Ctx(ctx)

	previousSummary, transcript, previousThrough, newThrough, ok, err := compactor.PendingCompaction(ctx, sessionID, userID)
	if err != nil {
		log.Warn("Failed to read chat for compaction", logger.Error(err), logger.String("session_id", sessionID))
		return
	}
	if !ok {
		return
	}
	tokens, err := h.AiService.GenerateTextStream(ctx, application.BuildCompactionPrompt(previousSummary, transcript))
	if err != nil {
		log.Warn("Failed to start chat compaction", logger.Error(err), logger.String("session_id", sessionID))
		return
	}
	var raw strings.Builder
	for token := range tokens {
		if token.Error != "" {
			log.Warn("Chat compaction stream failed", logger.String("error", token.Error), logger.String("session_id", sessionID))
			return
		}
		raw.WriteString(token.Text)
	}
	summary := application.CleanCompactionSummary(raw.String())
	if summary == "" {
		log.Warn("Chat compaction returned an empty summary", logger.String("session_id", sessionID))
		return
	}
	if err := compactor.SaveCompaction(ctx, sessionID, userID, previousThrough, newThrough, summary); err != nil {
		log.Warn("Failed to save chat compaction", logger.Error(err), logger.String("session_id", sessionID))
		return
	}
	log.Info("Compacted chat history",
		logger.String("session_id", sessionID),
		logger.Int("through_message_id", newThrough),
		logger.Int("summary_bytes", len(summary)))
}

const chatTitleTimeout = 30 * time.Second

// titleChat gives a chat a short written title after its first reply. It
// runs after the reply is sent; a failure keeps the title derived from the
// first message.
func (h *GenerateTextStreamHandler) titleChat(ctx context.Context, titler application.ChatTitler, sessionID string, userID int) {
	defer func() {
		if r := recover(); r != nil {
			logger.Ctx(ctx).Error("panic while titling chat", logger.Any("panic", r))
		}
	}()
	ctx, cancel := context.WithTimeout(ctx, chatTitleTimeout)
	defer cancel()

	current, prompt, reply, ok, err := titler.PendingChatTitle(ctx, sessionID, userID)
	if err != nil || !ok {
		return
	}
	tokens, err := h.AiService.GenerateTextStream(ctx, application.BuildChatTitlePrompt(prompt, reply))
	if err != nil {
		return
	}
	var raw strings.Builder
	for token := range tokens {
		if token.Error != "" {
			return
		}
		raw.WriteString(token.Text)
	}
	title := application.CleanChatTitle(raw.String())
	if title == "" {
		return
	}
	if err := titler.SaveChatTitle(ctx, sessionID, userID, current, title); err != nil {
		logger.Ctx(ctx).Warn("Failed to save chat title", logger.Error(err), logger.String("session_id", sessionID))
	}
}
