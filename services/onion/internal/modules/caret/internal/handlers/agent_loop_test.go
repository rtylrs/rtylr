package handlers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/caret/internal/application"
	"github.com/voxire/onion/internal/platform/snips"
)

func runLoop(t *testing.T, ai *sequencedAI, first string, executor ToolExecutor, continuations int) string {
	t.Helper()
	handler := &GenerateTextStreamHandler{AiService: ai, MemoryService: &detachedMemoryFake{}, ToolExecutors: []ToolExecutor{executor}}
	rec := httptest.NewRecorder()
	tokens := make(chan snips.StreamToken, 2)
	tokens <- snips.StreamToken{Text: first}
	tokens <- snips.StreamToken{Done: true}
	close(tokens)
	handler.streamTokensToClient(
		context.Background(), rec, tokens, make(chan bool),
		&GenerateTextStreamRequest{SessionID: "session-1", UserID: 7, EnhancedPrompt: "who is on my team and what is Rana's role?"},
		1, 7, cacheWritePlan{Mode: string(application.ModeAgent), ToolBudget: 10, ContinuationBudget: continuations}, nil,
	)
	return rec.Body.String()
}

// A task that keeps needing tools still ends in an answer: once the round
// budget is spent, a final round with tools off explains what ran.
func TestAgentLoopAlwaysEndsWithAnAnswer(t *testing.T) {
	ai := &sequencedAI{replies: []string{
		`<tool>{"name":"hr.list_employees","input":{}}</tool>`,
		`<tool>{"name":"hr.list_departments","input":{}}</tool>`,
		`Rana is on Operations. <tool>{"name":"hr.list_locations","input":{}}</tool>`,
	}}
	executor := namedToolExecutor{
		"tools.search":        {"tools": []string{"hr.list_employees"}},
		"hr.list_employees":   {"employees": []string{"Rana"}},
		"hr.list_departments": {"departments": []string{"Operations"}},
		"hr.list_locations":   {"locations": []string{"Beirut"}},
	}
	body := runLoop(t, ai, `<tool>{"name":"tools.search","input":{"query":"team"}}</tool>`, executor, 2)

	if len(ai.prompts) != 3 {
		t.Fatalf("expected 2 tool rounds plus a final answer round, got %d provider calls", len(ai.prompts))
	}
	final := ai.prompts[2]
	if !strings.Contains(final, "No more tools can run") {
		t.Fatalf("final round must ask for an answer with tools off:\n%.400s", final[len(final)-400:])
	}
	if strings.Contains(body, "hr.list_locations") {
		t.Fatal("a tool requested in the final round must not run")
	}
	if !strings.Contains(body, "Rana is on Operations.") {
		t.Fatalf("final answer missing: %s", body)
	}
}

// Each follow-up round sees what was already said, so it continues the
// reply instead of starting over.
func TestAgentLoopContinuesTheReplySoFar(t *testing.T) {
	ai := &sequencedAI{replies: []string{"Your team is Rana."}}
	executor := namedToolExecutor{"hr.list_employees": {"employees": []string{"Rana"}}}
	runLoop(t, ai, `Let me check your team. <tool>{"name":"hr.list_employees","input":{}}</tool>`, executor, 5)

	if len(ai.prompts) != 1 {
		t.Fatalf("expected one follow-up round, got %d", len(ai.prompts))
	}
	if !strings.Contains(ai.prompts[0], "<reply_so_far>Let me check your team.</reply_so_far>") {
		t.Fatalf("follow-up round did not see the reply so far")
	}
}
