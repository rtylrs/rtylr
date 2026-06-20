package memory

import (
	"context"
	"strings"

	"github.com/voxire/onion/internal/modules/memory/internal/domain"
)

// Compile-time assertion that the prompt builder still references the
// expected fields on the conversation context. If the domain shape drifts,
// this function will fail to compile here rather than silently produce a
// blank prompt.
var _ = domain.ConversationContext{}

// MaxContextMessages caps how many prior messages we feed back into a
// generated prompt so the request payload stays bounded regardless of
// session length.
const MaxContextMessages = 12

// MaxMemoriesInPrompt caps how many user memories surface alongside the
// summary; the repository orders them by confidence so the most relevant
// ones are picked first.
const MaxMemoriesInPrompt = 6

// BuildEnhancedPrompt fetches the conversation context for (sessionID,
// userID) and renders it together with the current request into a single
// prompt string suitable for an LLM.
//
// On any error fetching context this function returns the bare current
// request — it never fails the caller, since a degraded prompt is strictly
// better than a failed generation.
func (m *Module) BuildEnhancedPrompt(ctx context.Context, sessionID string, userID int, currentRequest string) string {
	if m == nil || m.memoryService == nil || sessionID == "" || userID <= 0 {
		return "Current request: " + currentRequest
	}

	cc, err := m.memoryService.GetConversationContext(ctx, sessionID, userID)
	if err != nil || cc == nil {
		return "Current request: " + currentRequest
	}

	var b strings.Builder

	if summary := strings.TrimSpace(cc.Summary); summary != "" {
		b.WriteString("Conversation summary so far:\n")
		b.WriteString(summary)
		b.WriteString("\n\n")
	}

	if len(cc.Memories) > 0 {
		b.WriteString("Saved Caret business memory:\n")
		limit := len(cc.Memories)
		if limit > MaxMemoriesInPrompt {
			limit = MaxMemoriesInPrompt
		}
		for i := 0; i < limit; i++ {
			memoryType := promptSafeMemoryText(cc.Memories[i].MemoryType)
			key := promptSafeMemoryText(cc.Memories[i].Key)
			value := promptSafeMemoryText(cc.Memories[i].Value)
			if key == "" || value == "" {
				continue
			}
			b.WriteString("- ")
			if memoryType != "" {
				b.WriteString(memoryType)
				b.WriteString(": ")
			}
			b.WriteString(key)
			b.WriteString(": ")
			b.WriteString(value)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if len(cc.Messages) > 0 {
		b.WriteString("Recent messages (oldest first):\n")
		start := 0
		if len(cc.Messages) > MaxContextMessages {
			start = len(cc.Messages) - MaxContextMessages
		}
		for i := start; i < len(cc.Messages); i++ {
			role := strings.TrimSpace(cc.Messages[i].Role)
			content := strings.TrimSpace(cc.Messages[i].Content)
			if content == "" {
				continue
			}
			if role == "" {
				role = "user"
			}
			b.WriteString(role)
			b.WriteString(": ")
			b.WriteString(content)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("Current request: ")
	b.WriteString(currentRequest)
	return b.String()
}

func promptSafeMemoryText(value string) string {
	value = strings.NewReplacer(
		"<tool>", "[tool]",
		"</tool>", "[/tool]",
		"<thinking>", "[thinking]",
		"</thinking>", "[/thinking]",
	).Replace(value)
	return strings.Join(strings.Fields(value), " ")
}
