package application

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
	// ProviderPlannerOnly means the provider may select a capability but must
	// not author a data-dependent answer. The application renders tool results
	// directly to the client.
	ProviderPlannerOnly bool
	AskClarification    bool
	AnswerShape         string
}

func BuildCaretResponsePolicy(mode CaretMode, request, systemContext string, pageContext map[string]any) CaretResponsePolicy {
	plan := ClassifyCaretRequest(request, systemContext)
	policy := aicache.ExactPolicyForMode(string(mode))
	result := CaretResponsePolicy{
		Plan:               plan,
		Grounding:          GroundingNone,
		ExactCacheEligible: policy.Enabled,
		CacheReason:        policy.Reason,
		ToolBudget:         DefaultToolCallBudget,
		ContinuationBudget: DefaultContinuationBudget,
		AskClarification:   plan.NeedsClarification || ClarificationHintForRequest(request, systemContext) != "",
		AnswerShape:        answerShapeForPlan(plan),
	}
	heavyReason, heavyData := HeavyDataRequestReason(request)

	switch {
	case mode == ModeSupport || plan.IsSmokeCheck || plan.IsConversational:
		result.Grounding = GroundingNone
	case heavyData && heavyReason != "":
		result.Grounding = GroundingDeferred
	case plan.Kind == IntentSnapshot || plan.IsBroad:
		result.Grounding = GroundingSnapshot
	case plan.NeedsLiveData:
		result.Grounding = GroundingAnalytics
	case HasDynamicPageContext(pageContext):
		result.Grounding = GroundingAnalytics
	}

	// Business-data turns are executed through server-side tools. Keep the old
	// grounding classification for policy/telemetry compatibility, but never
	// prefetch rows into the provider prompt.
	result.GroundBeforeCache = false
	result.ProviderPlannerOnly = mode.AllowsTools() && (plan.NeedsTool || result.Grounding != GroundingNone)
	if result.ProviderPlannerOnly {
		result.ExactCacheEligible = false
		result.CacheReason = "business data is resolved by a server-side tool"
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
	// There is no safe provider continuation after a tool result: the result is
	// delivered to the client directly and never serialized into another prompt.
	if result.ProviderPlannerOnly {
		result.ContinuationBudget = 0
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
