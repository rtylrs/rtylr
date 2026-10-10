package application

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

// Context window budgets, in bytes of prompt text (roughly four per token).
// The transcript is filled newest-first; anything older than the window is
// covered by the session's rolling summary instead of being silently lost.
const (
	transcriptBudgetBytes = 40000
	messageClipBytes      = 10000
	profileBudgetBytes    = 4000
	memoriesBudgetBytes   = 4000
	summaryBudgetBytes    = 6000
	projectBudgetBytes    = 8000

	// Compaction starts once unsummarized history no longer fits the window,
	// and folds older turns in until the newest compactKeepBytes remain
	// verbatim, so it runs every few turns rather than on every reply.
	compactKeepBytes = transcriptBudgetBytes / 2

	compactedThroughKey = "compacted_through_id"
)

// sessionCompaction is the rolling summary of every message up to ThroughID.
type sessionCompaction struct {
	Summary   string
	ThroughID int
}

func readCompaction(chat *domain.ChatSession) sessionCompaction {
	if chat == nil {
		return sessionCompaction{}
	}
	var metadata map[string]any
	_ = json.Unmarshal([]byte(chat.Metadata), &metadata)
	through, _ := metadata[compactedThroughKey].(float64)
	if through <= 0 || strings.TrimSpace(chat.Summary) == "" {
		return sessionCompaction{}
	}
	return sessionCompaction{Summary: strings.TrimSpace(chat.Summary), ThroughID: int(through)}
}

// orderedHistory sorts by insertion order and drops the trailing user message,
// which is the request being answered and is supplied to the model separately.
func orderedHistory(messages []domain.ConversationMessage) []domain.ConversationMessage {
	ordered := append([]domain.ConversationMessage(nil), messages...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	if n := len(ordered); n > 0 && strings.EqualFold(ordered[n-1].Role, "user") {
		ordered = ordered[:n-1]
	}
	return ordered
}

func uncoveredMessages(messages []domain.ConversationMessage, through int) []domain.ConversationMessage {
	for i, message := range messages {
		if message.ID > through {
			return messages[i:]
		}
	}
	return nil
}

// renderTurn writes one message as a transcript entry, with a note for each
// tool step the reply took so follow-ups like "search that again" make sense.
func renderTurn(message domain.ConversationMessage) string {
	speaker := "User"
	switch strings.ToLower(message.Role) {
	case "assistant":
		speaker = "Caret"
	case "system":
		speaker = "System"
	}
	var turn strings.Builder
	turn.WriteString(speaker)
	if !message.CreatedAt.IsZero() {
		turn.WriteString(" (" + message.CreatedAt.UTC().Format("2006-01-02 15:04") + " UTC)")
	}
	turn.WriteString(":\n")
	for _, note := range stepNotes(message.Metadata) {
		turn.WriteString("[" + note + "]\n")
	}
	turn.WriteString(clipMiddle(strings.TrimSpace(message.Content), messageClipBytes))
	return turn.String()
}

func stepNotes(metadata string) []string {
	var stored struct {
		Blocks []struct {
			Type    string          `json:"type"`
			Name    string          `json:"name"`
			Status  string          `json:"status"`
			Summary string          `json:"summary"`
			Input   json.RawMessage `json:"input"`
		} `json:"blocks"`
	}
	if json.Unmarshal([]byte(metadata), &stored) != nil {
		return nil
	}
	notes := make([]string, 0, len(stored.Blocks))
	for _, block := range stored.Blocks {
		if block.Type != "tool_call" || block.Name == "" {
			continue
		}
		note := "used " + block.Name
		if block.Name == "web.search" {
			var input struct {
				Query string `json:"query"`
			}
			_ = json.Unmarshal(block.Input, &input)
			note = "searched the web"
			if query := strings.TrimSpace(input.Query); query != "" {
				note += fmt.Sprintf(" for %q", clipMiddle(query, 200))
			}
		}
		if block.Status != "" && block.Status != "success" {
			note += " (" + block.Status + ")"
		} else if summary := strings.TrimSpace(block.Summary); summary != "" {
			note += ": " + clipMiddle(summary, 300)
		}
		notes = append(notes, note)
	}
	return notes
}

// transcriptWindow keeps the newest turns that fit in budget, oldest first.
// It reports how many uncovered turns did not fit.
func transcriptWindow(messages []domain.ConversationMessage, budget int) (string, int) {
	turns := make([]string, 0, len(messages))
	used := 0
	for i := len(messages) - 1; i >= 0; i-- {
		turn := renderTurn(messages[i])
		if used+len(turn) > budget && len(turns) > 0 {
			return joinOldestFirst(turns), i + 1
		}
		turns = append(turns, turn)
		used += len(turn) + 2
	}
	return joinOldestFirst(turns), 0
}

func joinOldestFirst(newestFirst []string) string {
	for i, j := 0, len(newestFirst)-1; i < j; i, j = i+1, j-1 {
		newestFirst[i], newestFirst[j] = newestFirst[j], newestFirst[i]
	}
	return strings.Join(newestFirst, "\n\n")
}

// compactionSplit returns the oldest uncovered messages to fold into the
// summary, or nil when the unsummarized history still fits the window.
func compactionSplit(uncovered []domain.ConversationMessage) []domain.ConversationMessage {
	total := 0
	for _, message := range uncovered {
		total += len(renderTurn(message)) + 2
	}
	if total <= transcriptBudgetBytes {
		return nil
	}
	kept := 0
	for i := len(uncovered) - 1; i > 0; i-- {
		kept += len(renderTurn(uncovered[i])) + 2
		if kept > compactKeepBytes {
			return uncovered[:i+1]
		}
	}
	return uncovered[:1]
}

func renderMemories(memories []domain.UserMemory) string {
	var out strings.Builder
	for _, memory := range memories {
		value := strings.TrimSpace(memory.Value)
		if value == "" {
			continue
		}
		line := "- " + value + "\n"
		if out.Len()+len(line) > memoriesBudgetBytes {
			break
		}
		out.WriteString(line)
	}
	return strings.TrimSpace(out.String())
}

// clipMiddle keeps the start and end of long text, where requests and
// conclusions usually are, and marks what was removed.
func clipMiddle(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	const marker = "\n[... middle of message omitted ...]\n"
	head := validUTF8Prefix(value, (maxBytes-len(marker))*2/3)
	tail := validUTF8Suffix(value, maxBytes-len(marker)-len(head))
	return head + marker + tail
}

func clipEnd(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	return validUTF8Prefix(value, maxBytes-3) + "..."
}

func validUTF8Prefix(value string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(value) <= maxBytes {
		return value
	}
	for maxBytes > 0 && !utf8.RuneStart(value[maxBytes]) {
		maxBytes--
	}
	return value[:maxBytes]
}

func validUTF8Suffix(value string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(value) <= maxBytes {
		return value
	}
	start := len(value) - maxBytes
	for start < len(value) && !utf8.RuneStart(value[start]) {
		start++
	}
	return value[start:]
}
