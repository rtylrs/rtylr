package handlers

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/platform/snips"
)

type compactingMemoryFake struct {
	detachedMemoryFake
	pending   bool
	saved     chan string
	savedFrom int
	savedTo   int
}

func (m *compactingMemoryFake) PendingCompaction(context.Context, string, int) (string, string, int, int, bool, error) {
	return "Earlier: user picked Friday.", "User:\nmove it to Monday", 4, 9, m.pending, nil
}

func (m *compactingMemoryFake) SaveCompaction(_ context.Context, _ string, _ int, from, to int, summary string) error {
	m.savedFrom, m.savedTo = from, to
	m.saved <- summary
	return nil
}

func TestStoreAIResponseCompactsLongChatsAfterTheReply(t *testing.T) {
	memory := &compactingMemoryFake{pending: true, saved: make(chan string, 1)}
	ai := &fakeAIStreamService{tokens: []snips.StreamToken{
		{Text: "<thinking>summarizing</thinking>Launch moved "},
		{Text: "from Friday to Monday."},
		{Done: true},
	}}
	handler := &GenerateTextStreamHandler{AiService: ai, MemoryService: memory}

	if err := handler.storeAIResponse(context.Background(), nil, "session-1", 7, "answer"); err != nil {
		t.Fatalf("storeAIResponse: %v", err)
	}

	select {
	case summary := <-memory.saved:
		if summary != "Launch moved from Friday to Monday." {
			t.Fatalf("summary = %q", summary)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("compaction did not run")
	}
	if memory.savedFrom != 4 || memory.savedTo != 9 {
		t.Fatalf("saved range = %d..%d, want 4..9", memory.savedFrom, memory.savedTo)
	}
	if len(ai.prompts) != 1 || !strings.Contains(ai.prompts[0], "Earlier: user picked Friday.") || !strings.Contains(ai.prompts[0], "move it to Monday") {
		t.Fatalf("compaction prompt missing summary or turns: %v", ai.prompts)
	}
	if memory.summarizeContextErr != nil {
		t.Fatal("legacy summary should not run when compaction is available")
	}
}

func TestContinuationPromptDropsOldestResultsAndKeepsInstruction(t *testing.T) {
	base := "BASE PROMPT"
	results := "OLDEST" + strings.Repeat("r", maxContinuationBytes) + "NEWEST"
	got := continuationPrompt(base, results)

	if len(got) > maxContinuationBytes {
		t.Fatalf("continuation is %d bytes, over %d", len(got), maxContinuationBytes)
	}
	if !strings.HasPrefix(got, base) || !strings.Contains(got, "NEWEST") || strings.Contains(got, "OLDEST") {
		t.Fatal("continuation must keep the base prompt and newest results, dropping the oldest")
	}
	if !strings.HasSuffix(got, "Do not repeat completed calls.") {
		t.Fatal("closing instruction must survive truncation")
	}
}
