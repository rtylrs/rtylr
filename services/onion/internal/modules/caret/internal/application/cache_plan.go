package application

import (
	"time"

	"github.com/voxire/onion/internal/platform/aicache"
)

// CacheWritePlan is the application decision consumed by the SSE transport.
// The handler may read it to replay or persist a response, but it does not
// decide eligibility, identity, freshness, or budgets.
type CacheWritePlan struct {
	Enabled             bool
	Key                 string
	Mode                string
	Provider            string
	Policy              aicache.ExactPolicy
	StartedAt           time.Time
	PromptChars         int
	ToolBudget          int
	ContinuationBudget  int
	ProviderPlannerOnly bool
}

func BuildCacheWritePlan(mode CaretMode, request TurnRequest, enhancedPrompt string, businessID, userID int64, provider string, now time.Time, groundings ...*AnalystGroundingResult) CacheWritePlan {
	policy := aicache.ExactPolicyForMode(string(mode))
	requestPolicy := BuildCaretResponsePolicy(mode, request.Prompt, request.SystemContext, request.PageContext)
	dataFreshness := BuildCacheFreshness(request.PageContext)
	if len(groundings) > 0 {
		dataFreshness = CombineCacheFreshness(dataFreshness, BuildGroundingCacheFreshness(groundings[0]))
	}

	plan := CacheWritePlan{
		Enabled:             policy.Enabled,
		Mode:                string(mode),
		Provider:            provider,
		Policy:              policy,
		StartedAt:           time.Now(),
		PromptChars:         len(enhancedPrompt),
		ToolBudget:          requestPolicy.ToolBudget,
		ContinuationBudget:  requestPolicy.ContinuationBudget,
		ProviderPlannerOnly: requestPolicy.ProviderPlannerOnly,
	}
	if policy.Enabled && requestPolicy.ExactCacheEligible {
		plan.Key = aicache.ExactKey(aicache.ExactKeyParts{
			BusinessID:    businessID,
			UserID:        userID,
			Mode:          string(mode),
			Provider:      provider,
			RequestDate:   now.UTC().Format("2006-01-02"),
			SystemContext: request.SystemContext,
			DataFreshness: dataFreshness,
			Prompt:        request.Prompt,
		})
	}
	if requestPolicy.AskClarification || requestPolicy.ProviderPlannerOnly || !requestPolicy.ExactCacheEligible || requestPolicy.Plan.Kind == IntentAction || requestPolicy.Plan.Kind == IntentMemory {
		plan = DisableExactCachePlan(plan, requestPolicy.CacheReason)
	}
	return plan
}

func DisableExactCachePlan(plan CacheWritePlan, reason string) CacheWritePlan {
	plan.Enabled = false
	plan.Key = ""
	plan.Policy.Enabled = false
	plan.Policy.Reason = reason
	return plan
}
