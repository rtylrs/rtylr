package handlers

import "github.com/voxire/onion/internal/platform/aicache"

type CaretGroundingMode string

const (
	GroundingNone      CaretGroundingMode = "none"
	GroundingSnapshot  CaretGroundingMode = "snapshot"
	GroundingAnalytics CaretGroundingMode = "analytics"
	GroundingDeferred  CaretGroundingMode = "deferred"
)

// CaretResponsePolicy is the backend execution decision for one user turn.
// Keeping it explicit prevents cache, grounding, tool, and clarification code
// from independently making contradictory guesses about the same request.
type CaretResponsePolicy struct {
	Plan               CaretRequestPlan
	Grounding          CaretGroundingMode
	GroundBeforeCache  bool
	ExactCacheEligible bool
	CacheReason        string
	ToolBudget         int
	ContinuationBudget int
	AskClarification   bool
	AnswerShape        string
}

func BuildCaretResponsePolicy(mode CaretMode, request, systemContext string, pageContext map[string]any) CaretResponsePolicy {
	plan := ClassifyCaretRequest(request, systemContext)
	policy := aicache.ExactPolicyForMode(string(mode))
	result := CaretResponsePolicy{
		Plan:               plan,
		Grounding:          GroundingNone,
		ExactCacheEligible: policy.Enabled,
		CacheReason:        policy.Reason,
		ToolBudget:         maxToolCallsPerTurn,
		ContinuationBudget: maxToolResultContinuations,
		AskClarification:   plan.NeedsClarification || ClarificationHintForRequest(request, systemContext) != "",
		AnswerShape:        answerShapeForPlan(plan),
	}
	heavyReason, heavyData := heavyDataRequestReason(request)

	switch {
	case mode == ModeSupport || plan.IsSmokeCheck || plan.IsConversational:
		result.Grounding = GroundingNone
	case heavyData && heavyReason != "":
		result.Grounding = GroundingDeferred
	case plan.Kind == IntentSnapshot || plan.IsBroad:
		result.Grounding = GroundingSnapshot
	case plan.NeedsLiveData:
		result.Grounding = GroundingAnalytics
	case hasDynamicPageContext(pageContext):
		result.Grounding = GroundingAnalytics
	}

	result.GroundBeforeCache = result.Grounding != GroundingNone && mode != ModeSupport
	if result.GroundBeforeCache {
		result.ExactCacheEligible = false
		result.CacheReason = "fresh business evidence is required before replay"
	}
	if result.AskClarification {
		result.ToolBudget = 0
		result.ContinuationBudget = 0
		result.ExactCacheEligible = false
		result.CacheReason = "request needs one clarification before execution"
	}
	if plan.Kind == IntentAction || plan.Kind == IntentMemory {
		result.ToolBudget = minInt(result.ToolBudget, 3)
		result.ExactCacheEligible = false
		result.CacheReason = "mutating requests are never replayed"
	}
	return result
}

func answerShapeForPlan(plan CaretRequestPlan) string {
	switch {
	case plan.NeedsClarification:
		return "clarification"
	case plan.Kind == IntentConversation:
		return "conversation"
	case plan.Kind == IntentAction || plan.Kind == IntentMemory:
		return "action_plan"
	case plan.Action == "compare":
		return "comparison"
	case plan.IsBroad || plan.Kind == IntentSnapshot:
		return "operator_summary"
	default:
		return "direct_result"
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
