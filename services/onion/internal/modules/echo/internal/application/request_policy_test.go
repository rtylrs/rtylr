package application

import "testing"

func TestBuildCaretResponsePolicyUsesOneGroundingDecision(t *testing.T) {
	cases := []struct {
		name        string
		mode        CaretMode
		request     string
		grounding   CaretGroundingMode
		fresh       bool
		plannerOnly bool
		cache       bool
		answerShape string
	}{
		{name: "conversation", mode: ModeAsk, request: "thanks", grounding: GroundingNone, cache: true, answerShape: "conversation"},
		{name: "snapshot", mode: ModeAsk, request: "what needs attention in my business?", grounding: GroundingSnapshot, plannerOnly: true, cache: false, answerShape: "operator_summary"},
		{name: "live analytics", mode: ModeAsk, request: "show today's revenue", grounding: GroundingAnalytics, plannerOnly: true, cache: false, answerShape: "direct_result"},
		{name: "mutation", mode: ModeAgent, request: "create a customer", grounding: GroundingNone, plannerOnly: true, cache: false, answerShape: "action_plan"},
		{name: "support", mode: ModeSupport, request: "what is a refund?", grounding: GroundingNone, cache: true, answerShape: "direct_result"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildCaretResponsePolicy(tc.mode, tc.request, "", nil)
			if got.Grounding != tc.grounding || got.GroundBeforeCache != tc.fresh || got.ProviderPlannerOnly != tc.plannerOnly || got.ExactCacheEligible != tc.cache || got.AnswerShape != tc.answerShape {
				t.Fatalf("policy = %+v", got)
			}
		})
	}
}

func TestBuildCaretResponsePolicyDisablesExecutionUntilClarification(t *testing.T) {
	got := BuildCaretResponsePolicy(ModeAgent, "update this customer", "", nil)
	if !got.AskClarification || got.ToolBudget != 0 || got.ExactCacheEligible {
		t.Fatalf("ambiguous mutation policy = %+v", got)
	}
}
