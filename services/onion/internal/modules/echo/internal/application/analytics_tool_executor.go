package application

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

// AnalyticsQueryToolExecutor keeps analytics execution on the server side.
// The provider may help translate the user's question into an allowlisted
// query spec, but it never receives the rows produced by that query.
type AnalyticsQueryToolExecutor struct {
	AI             TextCompletionService
	AnalystService AnalystService
}

func NewAnalyticsQueryToolExecutor(ai TextCompletionService, analystService AnalystService) *AnalyticsQueryToolExecutor {
	return &AnalyticsQueryToolExecutor{AI: ai, AnalystService: analystService}
}

func (e *AnalyticsQueryToolExecutor) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != "analytics.query" {
		return nil, false, nil
	}

	businessID, userID, err := businessSnapshotToolIdentity(ctx)
	if err != nil {
		return nil, true, err
	}
	question := strings.TrimSpace(businessSnapshotToolString(firstBusinessSnapshotToolValue(input, "question", "prompt", "query")))
	if question == "" {
		return nil, true, errors.New("analytics question is required")
	}

	grounding, ok := runAnalyticsQuery(ctx, e.AI, e.AnalystService, question, businessID, userID, "", nil)
	if !ok {
		return nil, true, errors.New("analytics query is unavailable")
	}

	output := map[string]any{
		"rows":            grounding.Rows,
		"row_count":       grounding.RowCount,
		"truncated":       grounding.Truncated,
		"duration_ms":     grounding.DurationMs,
		"evidence_source": grounding.EvidenceSource,
	}
	if strings.TrimSpace(grounding.SpecJSON) != "" {
		output["spec"] = json.RawMessage(grounding.SpecJSON)
	}
	return output, true, nil
}
