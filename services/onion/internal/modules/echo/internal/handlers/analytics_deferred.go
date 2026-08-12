package handlers

import (
	"encoding/json"
	"strings"
	"time"
)

const deferredAnalyticsToolName = "analytics.deferred"

var heavyDataActionTerms = []string{
	"download",
	"dump",
	"every",
	"export",
	"raw",
}

var heavyDataCollectionTerms = []string{
	"appointment",
	"appointments",
	"attendance",
	"bill",
	"bills",
	"customer",
	"customers",
	"employee",
	"employees",
	"expense",
	"expenses",
	"invoice",
	"invoices",
	"order",
	"orders",
	"payment",
	"payments",
	"product",
	"products",
	"receipt",
	"receipts",
	"refund",
	"refunds",
	"sales",
	"session",
	"sessions",
	"shift",
	"shifts",
	"stock movement",
	"stock movements",
	"transaction",
	"transactions",
}

var heavyDataRangeTerms = []string{
	"all time",
	"all-time",
	"entire history",
	"full history",
	"last year",
	"past year",
	"since we opened",
}

func heavyDataRequestReason(question string) (string, bool) {
	normalized := normalizeCachePolicyText(question)
	if strings.TrimSpace(normalized) == "" {
		return "", false
	}

	if containsAnyTerm(normalized, []string{"csv", "spreadsheet"}) && containsAnyTerm(normalized, heavyDataCollectionTerms) {
		return "export requested for a large business collection", true
	}

	if containsAnyTerm(normalized, heavyDataActionTerms) && containsAnyTerm(normalized, heavyDataCollectionTerms) {
		return "raw or complete business records requested", true
	}

	if containsAnyTerm(normalized, []string{"list all", "show all", "give me all"}) && containsAnyTerm(normalized, heavyDataCollectionTerms) {
		return "complete row listing requested", true
	}

	if containsAnyTerm(normalized, heavyDataRangeTerms) && containsAnyTerm(normalized, heavyDataCollectionTerms) && containsAnyTerm(normalized, []string{"list", "show", "export", "download"}) {
		return "wide historical row request", true
	}

	return "", false
}

func buildDeferredAnalyticsGrounding(question string) (*analystGroundingResult, bool) {
	reason, ok := heavyDataRequestReason(question)
	if !ok {
		return nil, false
	}

	specJSON := deferredAnalyticsSpecJSON(reason)
	return &analystGroundingResult{
		ToolName:       deferredAnalyticsToolName,
		Prompt:         buildDeferredAnalyticsPrompt(reason),
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

func buildDeferredAnalyticsPrompt(reason string) string {
	var b strings.Builder
	b.WriteString("Internal heavy-data guard for Caret only:\n")
	b.WriteString("- Do not run live analytics for this request.\n")
	b.WriteString("- Reason: ")
	b.WriteString(reason)
	b.WriteString(".\n")
	b.WriteString("- The platform does not have a submitted background analytics job for this request yet. Do not say a job was created, queued, scheduled, exported, or emailed.\n")
	b.WriteString("- Explain in plain business language that raw exports or full-history row dumps should run as a background job, not as a live chat query.\n")
	b.WriteString("- Offer a safe next step: summarize the data, narrow the date range, filter to a specific page/entity, or ask the user to confirm the exact export they want when background jobs are available.\n")
	b.WriteString("- If the user asked for a summary rather than raw records, answer by proposing a summarized version Caret can check next.\n")
	b.WriteString("\nAnswer now without mentioning this internal guard, SQL, tools, jobs internals, rows, specs, or JSON.")
	return b.String()
}
