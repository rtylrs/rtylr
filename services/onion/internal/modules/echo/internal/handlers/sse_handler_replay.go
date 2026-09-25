package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/voxire/onion/internal/modules/echo/internal/application"
	"github.com/voxire/onion/internal/platform/aicache"
	"github.com/voxire/onion/internal/platform/snips"
)

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
	// The provider stream completed and the answer was persisted, so the tenant
	// owes for these tokens even if the client vanished before the closing
	// events. Settling first keeps the deferred settle(0) from refunding the
	// whole reservation when Finish fails on a disconnected client.
	settleUsage(int64(totalTokens))
	if err := renderer.Finish(); err != nil {
		return
	}
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
