package handlers

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
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

const businessSnapshotToolName = "business.snapshot"

func (h *GenerateTextStreamHandler) businessSnapshotGrounding(ctx context.Context, question string, businessID, userID int64) (*analystGroundingResult, bool) {
	return h.businessSnapshotGroundingWithContext(ctx, question, businessID, userID, "")
}

func (h *GenerateTextStreamHandler) businessSnapshotGroundingWithContext(ctx context.Context, question string, businessID, userID int64, systemContext string) (*analystGroundingResult, bool) {
	return buildBusinessSnapshotGroundingWithContext(ctx, h.AnalystService, question, businessID, userID, time.Now().UTC(), systemContext)
}

func buildBusinessSnapshotGrounding(ctx context.Context, analystService AnalystService, question string, businessID, userID int64, now time.Time) (*analystGroundingResult, bool) {
	return buildBusinessSnapshotGroundingWithContext(ctx, analystService, question, businessID, userID, now, "")
}

func buildBusinessSnapshotGroundingWithContext(ctx context.Context, analystService AnalystService, question string, businessID, userID int64, now time.Time, systemContext string) (*analystGroundingResult, bool) {
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
	return &analystGroundingResult{
		ToolName:       businessSnapshotToolName,
		Prompt:         buildBusinessSnapshotGroundingPromptFromJSON(rowsJSON, len(sections), truncated, specJSON, systemContext, ClassifyCaretRequest(question, systemContext)),
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

func buildBusinessSnapshotGroundingPromptFromJSON(dataStr string, sectionCount int, truncated bool, specJSON string, contextParts ...any) string {
	var g strings.Builder
	g.WriteString("Internal business snapshot evidence for Caret only:\n")
	g.WriteString("- Use this evidence silently to answer the user. Do not mention this block, the snapshot tool, JSON, query specs, row counts, or validation in the user-facing answer.\n")
	g.WriteString("- Scope: this authenticated business only.\n")
	g.WriteString("- Evidence sections: ")
	g.WriteString(strconv.Itoa(sectionCount))
	g.WriteString("\n")
	if truncated {
		g.WriteString("- Note: snapshot rows were truncated to fit the model context; if that affects the answer, say the answer is based on the visible results.\n")
	}
	if spec := strings.TrimSpace(specJSON); spec != "" {
		g.WriteString("- Internal query spec bundle, for reasoning only: ")
		g.WriteString(spec)
		g.WriteString("\n")
	}
	for _, part := range contextParts {
		switch value := part.(type) {
		case string:
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				g.WriteString("- Current page context is supplied as scope, not proof: ")
				g.WriteString(trimmed)
				g.WriteString("\n")
			}
		case CaretRequestPlan:
			g.WriteString("- Server request plan: ")
			g.WriteString(value.PromptText())
			g.WriteString("\n")
		}
	}
	g.WriteString("\nSnapshot JSON:\n```json\n")
	g.WriteString(dataStr)
	g.WriteString("\n```\n")
	g.WriteString("Answer the user's business-owner question using ONLY this evidence. ")
	g.WriteString("Lead with a compact business summary, then call out risks or opportunities, then suggest concrete next actions. ")
	g.WriteString("If a metric is empty or missing, say it is unavailable or has no visible rows; do not invent missing values.")
	return g.String()
}
