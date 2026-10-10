package handlers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/caret/internal/application"
	"github.com/voxire/onion/internal/platform/snips"
)

// sequencedAI answers each provider call with the next scripted reply.
type sequencedAI struct {
	replies []string
	prompts []string
}

func (s *sequencedAI) IsAvailable() bool          { return true }
func (s *sequencedAI) GetPrimaryProvider() string { return "test-provider" }
func (s *sequencedAI) GenerateTextStream(_ context.Context, prompt string) (<-chan snips.StreamToken, error) {
	s.prompts = append(s.prompts, prompt)
	reply := ""
	if len(s.replies) > 0 {
		reply, s.replies = s.replies[0], s.replies[1:]
	}
	tokens := make(chan snips.StreamToken, 2)
	tokens <- snips.StreamToken{Text: reply}
	tokens <- snips.StreamToken{Done: true}
	close(tokens)
	return tokens, nil
}

type namedToolExecutor map[string]map[string]any

func (n namedToolExecutor) ExecuteCaretTool(_ context.Context, name string, _ map[string]any) (map[string]any, bool, error) {
	output, ok := n[name]
	return output, ok, nil
}

func TestToolSearchUnlocksToolsOutsideTheShortlist(t *testing.T) {
	ai := &sequencedAI{replies: []string{
		`<tool>{"name":"hr.list_employees","input":{}}</tool>`,
		"Your team is Rana and Karim.",
	}}
	handler := &GenerateTextStreamHandler{
		AiService:     ai,
		MemoryService: &detachedMemoryFake{},
		ToolExecutors: []ToolExecutor{
			application.NewAssistantToolExecutor(nil, nil),
			namedToolExecutor{"hr.list_employees": {"employees": []string{"Rana", "Karim"}}},
		},
	}
	rec := httptest.NewRecorder()
	tokens := make(chan snips.StreamToken, 2)
	tokens <- snips.StreamToken{Text: `<tool>{"name":"tools.search","input":{"query":"team members"}}</tool>`}
	tokens <- snips.StreamToken{Done: true}
	close(tokens)

	handler.streamTokensToClient(
		context.Background(), rec, tokens, make(chan bool),
		&GenerateTextStreamRequest{SessionID: "session-1", UserID: 7, EnhancedPrompt: "who do i have in the team?"},
		1, 7, cacheWritePlan{Mode: string(application.ModeAgent), ToolBudget: 6, ContinuationBudget: 3}, nil,
	)

	body := rec.Body.String()
	if !strings.Contains(body, "Your team is Rana and Karim.") {
		t.Fatalf("discovered tool did not run through to an answer: %s", body)
	}
	if len(ai.prompts) < 1 || !strings.Contains(ai.prompts[0], "hr.list_employees") {
		t.Fatalf("search results with the HR tool were not given to the model: %v", ai.prompts)
	}
	if len(ai.prompts) < 2 || !strings.Contains(ai.prompts[1], "Karim") {
		t.Fatalf("HR tool output was not given to the model: %v", ai.prompts)
	}
}
