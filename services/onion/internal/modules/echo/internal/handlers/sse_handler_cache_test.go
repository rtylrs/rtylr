package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	echoapp "github.com/voxire/onion/internal/modules/echo/internal/application"
)

func TestBuildCacheWritePlanIncludesDataFreshness(t *testing.T) {
	now := time.Date(2026, 6, 20, 11, 0, 0, 0, time.UTC)
	req := &GenerateTextStreamRequest{
		Prompt:        "Which products are low?",
		SystemContext: "User is in inventory.",
		PageContext: map[string]any{
			"module":        "erp",
			"dataFreshness": "products:1",
		},
	}

	first := echoapp.BuildCacheWritePlan(echoapp.ModeAsk, echoapp.TurnRequest{Prompt: req.Prompt, SystemContext: req.SystemContext, PageContext: req.PageContext}, "Current request: Which products are low?", 11, 22, "", now)
	req.PageContext["dataFreshness"] = "products:2"
	second := echoapp.BuildCacheWritePlan(echoapp.ModeAsk, echoapp.TurnRequest{Prompt: req.Prompt, SystemContext: req.SystemContext, PageContext: req.PageContext}, "Current request: Which products are low?", 11, 22, "", now)

	if first.Enabled || second.Enabled || first.Key != "" || second.Key != "" {
		t.Fatalf("business-data turns must not be exact cached: first=%+v second=%+v", first, second)
	}
	if !first.ProviderPlannerOnly || !second.ProviderPlannerOnly {
		t.Fatalf("business-data turns must be planner-only: first=%+v second=%+v", first, second)
	}
}

func TestBuildCacheWritePlanIncludesGroundingFreshness(t *testing.T) {
	now := time.Date(2026, 6, 20, 11, 0, 0, 0, time.UTC)
	req := &GenerateTextStreamRequest{
		Prompt:        "How much revenue today?",
		SystemContext: "User is in dashboard.",
	}

	firstGrounding := &analystGroundingResult{
		ToolName: "analytics.query",
		SpecJSON: `{"view":"ai_pos_orders","aggregations":[{"func":"sum","col":"total","as":"revenue"}]}`,
		RowsJSON: `[{"revenue":100}]`,
		RowCount: 1,
	}
	secondGrounding := &analystGroundingResult{
		ToolName: "analytics.query",
		SpecJSON: firstGrounding.SpecJSON,
		RowsJSON: `[{"revenue":120}]`,
		RowCount: 1,
	}

	turnRequest := echoapp.TurnRequest{Prompt: req.Prompt, SystemContext: req.SystemContext, PageContext: req.PageContext}
	first := echoapp.BuildCacheWritePlan(echoapp.ModeAsk, turnRequest, "grounded prompt", 11, 22, "", now, firstGrounding)
	second := echoapp.BuildCacheWritePlan(echoapp.ModeAsk, turnRequest, "grounded prompt", 11, 22, "", now, secondGrounding)

	if first.Enabled || second.Enabled || first.Key != "" || second.Key != "" {
		t.Fatalf("grounded data turns must not be exact cached: first=%+v second=%+v", first, second)
	}
}

func TestShouldGroundBeforeExactCacheForLiveBusinessQuestions(t *testing.T) {
	cases := []string{
		"How much revenue today?",
		"top 5 customers",
		"show my top customers",
		"How many employees do I have?",
		"Which products are low on stock?",
	}
	for _, prompt := range cases {
		policy := echoapp.BuildCaretResponsePolicy(echoapp.ModeAsk, prompt, "", nil)
		if policy.GroundBeforeCache {
			t.Fatalf("provider path must not pre-ground %q", prompt)
		}
		if !policy.ProviderPlannerOnly {
			t.Fatalf("expected %q to use planner-only execution", prompt)
		}
	}

	if echoapp.BuildCaretResponsePolicy(echoapp.ModeSupport, "How much revenue today?", "", nil).GroundBeforeCache {
		t.Fatal("support mode should not use business grounding")
	}
	if echoapp.BuildCaretResponsePolicy(echoapp.ModeAsk, "How do I use Caret?", "", nil).GroundBeforeCache {
		t.Fatal("product help should not force live business grounding")
	}
}

func TestCachedReplayChunksPreserveTextAndSplitNaturally(t *testing.T) {
	text := "Here is the first cached sentence. Here is the second cached sentence, with enough words to replay in more than one visible delta. Final."

	chunks := cachedReplayChunks(text)
	if len(chunks) < 2 {
		t.Fatalf("expected cached replay to split text into visible chunks, got %d", len(chunks))
	}
	if strings.Join(chunks, "") != text {
		t.Fatalf("cached replay chunks changed text: %#v", chunks)
	}
	for i, chunk := range chunks[:len(chunks)-1] {
		if len([]rune(chunk)) > cachedReplayMaxChunkRunes {
			t.Fatalf("chunk %d too large: %q", i, chunk)
		}
	}
}

func TestCachedReplayChunksSplitShortAnswers(t *testing.T) {
	text := "You have 2 employees."

	chunks := cachedReplayChunks(text)
	if len(chunks) < 2 {
		t.Fatalf("expected short cached answer to replay visibly, got %d chunk", len(chunks))
	}
	if strings.Join(chunks, "") != text {
		t.Fatalf("cached replay chunks changed text: %#v", chunks)
	}
}

func TestCachedReplayDelayIsBounded(t *testing.T) {
	if got := cachedReplayDelay(1); got != 0 {
		t.Fatalf("single chunk delay = %s, want 0", got)
	}
	if got := cachedReplayDelay(3); got != cachedReplayMaxDelay {
		t.Fatalf("short replay delay = %s, want %s", got, cachedReplayMaxDelay)
	}
	if got := cachedReplayDelay(500); got != cachedReplayMinDelay {
		t.Fatalf("long replay delay = %s, want %s", got, cachedReplayMinDelay)
	}
}

func TestStreamRendererExtractsWriteToolProposal(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full, withStreamMode(ModeAgent))

	err := renderer.AppendText(`I can do that. <tool>{"name":"crm.create_customer","input":{"name":"Ada"}}</tool> Waiting for approval.`)
	if err != nil {
		t.Fatalf("append tool chunk: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish stream: %v", err)
	}

	events := decodeSSEEvents(t, rec.Body.String())
	var toolCall *StreamEvent
	for i := range events {
		if events[i].Type == EventToolCall {
			toolCall = &events[i]
			break
		}
	}
	if toolCall == nil {
		t.Fatalf("expected tool_call event, got %#v", events)
	}
	if toolCall.Name != "crm.create_customer" || !toolCall.RequiresApproval {
		t.Fatalf("expected approval-backed create customer proposal, got %#v", *toolCall)
	}
	if toolCall.Mode != string(ModeAgent) {
		t.Fatalf("expected tool call to carry originating mode, got %#v", *toolCall)
	}
	if strings.Contains(full.String(), "<tool>") || strings.Contains(full.String(), "crm.create_customer") {
		t.Fatalf("public response should not include raw tool marker, got %q", full.String())
	}
}

func TestStreamRendererDropsWriteToolProposalOutsideAgentMode(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full, withStreamMode(ModeAsk))

	err := renderer.AppendText(`I can do that. <tool>{"name":"crm.create_customer","input":{"name":"Ada"}}</tool>`)
	if err != nil {
		t.Fatalf("append tool chunk: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish stream: %v", err)
	}

	events := decodeSSEEvents(t, rec.Body.String())
	for _, event := range events {
		if event.Type == EventToolCall || event.Type == EventToolResult || event.Type == EventError {
			t.Fatalf("ask mode must not surface write tool markers, got %#v", events)
		}
	}
	if full.String() != "I can do that. " {
		t.Fatalf("expected visible text before marker to remain, got %q", full.String())
	}
}

func TestStreamRendererDoesNotExposeSplitInternalCapabilityName(t *testing.T) {
	rec := httptest.NewRecorder()
	full := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", full)

	text := "I used insights.get_customer_kpis to check the customer data." + strings.Repeat(" ", markerHoldBytes+20)
	if err := renderer.AppendText(text); err != nil {
		t.Fatalf("append split internal detail: %v", err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatalf("finish stream: %v", err)
	}

	if strings.Contains(full.String(), "insights.get_customer_kpis") {
		t.Fatalf("public response exposed an internal capability name: %q", full.String())
	}
}
