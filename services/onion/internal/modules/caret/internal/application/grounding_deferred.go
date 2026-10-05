package application

import (
	"encoding/json"
	"time"
)

const deferredAnalyticsToolName = "analytics.deferred"

func buildDeferredAnalyticsGrounding(question string) (*AnalystGroundingResult, bool) {
	reason, ok := HeavyDataRequestReason(question)
	if !ok {
		return nil, false
	}

	specJSON := deferredAnalyticsSpecJSON(reason)
	return &AnalystGroundingResult{
		ToolName:       deferredAnalyticsToolName,
		Question:       question,
		SpecJSON:       specJSON,
		RowsJSON:       "[]",
		Rows:           nil,
		RowCount:       0,
		EvidenceSource: "deferred heavy-data guard",
		GeneratedAt:    time.Now().UTC(),
	}, true
}

func deferredAnalyticsSpecJSON(reason string) string {
	data, err := json.Marshal(struct {
		Kind   string `json:"kind"`
		Reason string `json:"reason"`
	}{
		Kind:   "deferred_analytics",
		Reason: reason,
	})
	if err != nil {
		return `{"kind":"deferred_analytics"}`
	}
	return string(data)
}
