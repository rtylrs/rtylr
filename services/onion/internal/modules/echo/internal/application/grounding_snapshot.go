package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

type businessSnapshotQuery struct {
	Metric string
	Label  string
	Spec   string
}

type businessSnapshotSpecItem struct {
	Metric string          `json:"metric"`
	Label  string          `json:"label"`
	Spec   json.RawMessage `json:"spec"`
}

func shouldUseBusinessSnapshot(question string) bool {
	plan := ClassifyCaretRequest(question, "")
	return plan.Kind == IntentSnapshot || plan.IsBroad
}

const BusinessSnapshotToolName = "business.snapshot"

func BuildBusinessSnapshotGrounding(ctx context.Context, analystService AnalystService, question string, businessID, userID int64, now time.Time) (*AnalystGroundingResult, bool) {
	return BuildBusinessSnapshotGroundingWithContext(ctx, analystService, question, businessID, userID, now, "")
}

func BuildBusinessSnapshotGroundingWithContext(ctx context.Context, analystService AnalystService, question string, businessID, userID int64, now time.Time, systemContext string) (*AnalystGroundingResult, bool) {
	if analystService == nil {
		return nil, false
	}

	started := time.Now()
	queries := buildBusinessSnapshotQueries(now)
	sections := make([]map[string]any, 0, len(queries))
	specItems := make([]businessSnapshotSpecItem, 0, len(queries))

	for _, query := range queries {
		rows, _, err := analystService.RunJSON(ctx, businessID, userID, query.Spec)
		if err != nil {
			logger.Ctx(ctx).Warn("business snapshot query failed",
				logger.String("metric", query.Metric),
				logger.Error(err),
			)
			continue
		}

		sections = append(sections, map[string]any{
			"metric":    query.Metric,
			"label":     query.Label,
			"row_count": len(rows),
			"rows":      rows,
		})
		specItems = append(specItems, businessSnapshotSpecItem{
			Metric: query.Metric,
			Label:  query.Label,
			Spec:   json.RawMessage(query.Spec),
		})
	}

	if len(sections) == 0 {
		return nil, false
	}

	rowsJSON, truncated := marshalGroundingRows(sections, maxGroundingChars)
	specJSON := buildBusinessSnapshotSpecJSON(now.Format("2006-01-02"), specItems)
	return &AnalystGroundingResult{
		ToolName:       BusinessSnapshotToolName,
		Question:       question,
		SpecJSON:       specJSON,
		RowsJSON:       rowsJSON,
		Rows:           rowsPreviewFromJSON(rowsJSON),
		RowCount:       len(sections),
		Truncated:      truncated,
		DurationMs:     time.Since(started).Milliseconds(),
		EvidenceSource: "business snapshot views",
		RequestPlan:    ClassifyCaretRequest(question, systemContext).PromptText(),
		GeneratedAt:    time.Now().UTC(),
	}, true
}

func buildBusinessSnapshotQueries(now time.Time) []businessSnapshotQuery {
	today := now.Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	return []businessSnapshotQuery{
		{
			Metric: "today_revenue",
			Label:  "Paid POS revenue today",
			Spec:   `{"view":"ai_pos_orders","aggregations":[{"func":"sum","col":"total","as":"revenue"}],"where":[{"col":"payment_status","op":"=","value":"paid"},{"col":"paid_at","op":">=","value":"` + today + `"},{"col":"paid_at","op":"<","value":"` + tomorrow + `"}]}`,
		},
		{
			Metric: "today_orders",
			Label:  "Paid POS orders today",
			Spec:   `{"view":"ai_pos_orders","aggregations":[{"func":"count","as":"orders"}],"where":[{"col":"payment_status","op":"=","value":"paid"},{"col":"paid_at","op":">=","value":"` + today + `"},{"col":"paid_at","op":"<","value":"` + tomorrow + `"}]}`,
		},
		{
			Metric: "open_invoice_exposure",
			Label:  "Unpaid invoice exposure",
			Spec:   `{"view":"ai_invoices","aggregations":[{"func":"sum","col":"amount_due","as":"amount_due"},{"func":"count","as":"invoice_count"}],"where":[{"col":"amount_due","op":">","value":0}]}`,
		},
		{
			Metric: "month_expenses",
			Label:  "Month-to-date expenses",
			Spec:   `{"view":"ai_expenses","aggregations":[{"func":"sum","col":"amount","as":"expenses"}],"where":[{"col":"expense_date","op":">=","value":"` + monthStart + `"},{"col":"expense_date","op":"<","value":"` + tomorrow + `"}]}`,
		},
		{
			Metric: "customer_base",
			Label:  "Customer base",
			Spec:   `{"view":"ai_customers","aggregations":[{"func":"count","as":"customers"}]}`,
		},
		{
			Metric: "active_employees",
			Label:  "Active employees",
			Spec:   `{"view":"ai_employees","aggregations":[{"func":"count","as":"employees"}],"where":[{"col":"status","op":"=","value":"active"}]}`,
		},
		{
			Metric: "low_stock_products",
			Label:  "Low-stock inventory",
			Spec:   `{"view":"ai_products","columns":["name","sku","stock_quantity","low_stock_threshold"],"where":[{"col":"is_active","op":"=","value":true},{"col":"track_inventory","op":"=","value":true},{"col":"low_stock_threshold","op":">=","value":0},{"col":"stock_quantity","op":"<=","value_col":"low_stock_threshold"}],"order_by":[{"col":"stock_quantity","dir":"asc"}],"limit":10}`,
		},
		{
			Metric: "task_status",
			Label:  "Flow tasks by status",
			Spec:   `{"view":"ai_tasks","columns":["status"],"aggregations":[{"func":"count","as":"tasks"}],"group_by":["status"],"order_by":[{"col":"tasks","dir":"desc"}],"limit":10}`,
		},
	}
}

func buildBusinessSnapshotSpecJSON(date string, items []businessSnapshotSpecItem) string {
	doc := struct {
		Kind    string                     `json:"kind"`
		Date    string                     `json:"date"`
		Queries []businessSnapshotSpecItem `json:"queries"`
	}{
		Kind:    "business_snapshot",
		Date:    date,
		Queries: items,
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return `{"kind":"business_snapshot"}`
	}
	return string(data)
}
