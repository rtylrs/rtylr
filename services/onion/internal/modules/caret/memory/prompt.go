package memory

import (
	"context"
	"strings"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

// Compile-time assertion that the prompt builder still references the
// expected fields on the conversation context. If the domain shape drifts,
// this function will fail to compile here rather than silently produce a
// blank prompt.
var _ = domain.ConversationContext{}

const (
	maxMessagePromptRunes = 2000
)

// BuildPlannerPrompt deliberately excludes conversation messages, summaries,
// saved memory values, page rows, and prior tool results. Those values remain
// available to server-side product features, but are not sent to an external
// provider merely to choose a tool.
func (m *Module) BuildPlannerPrompt(_ context.Context, _ string, _ int, currentRequest string) string {
	currentRequest = promptSafeContextText(currentRequest, maxMessagePromptRunes)
	return "Current request: " + currentRequest
}

// BuildEnhancedPrompt remains as a compatibility alias for callers outside
// the Caret turn preparation path. It is data-free for the same reason as
// BuildPlannerPrompt.
func (m *Module) BuildEnhancedPrompt(ctx context.Context, sessionID string, userID int, currentRequest string) string {
	return m.BuildPlannerPrompt(ctx, sessionID, userID, currentRequest)
}

func promptSafeContextText(value string, maxRunes int) string {
	value = strings.NewReplacer(
		"<tool>", "[tool]",
		"</tool>", "[/tool]",
		"<thinking>", "[thinking]",
		"</thinking>", "[/thinking]",
	).Replace(value)
	value = strings.Join(strings.Fields(value), " ")
	if maxRunes <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[:maxRunes]) + "..."
}
