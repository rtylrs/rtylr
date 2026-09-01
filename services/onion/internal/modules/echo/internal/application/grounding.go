package application

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/snips"
)

// AnalystService is the safe NL-to-SQL engine (implemented by the aiquery
// module). Kept as a local interface with primitive types so echo doesn't depend
// on aiquery's internals.
type AnalystService interface {
	// SchemaPrompt returns the allowlisted views/columns to show the model.
	SchemaPrompt() string
	// RunJSON validates an LLM-produced query spec and runs it for the caller's
	// tenant, returning rows + the compiled SQL.
	RunJSON(ctx context.Context, businessID, userID int64, specJSON string) ([]map[string]any, string, error)
}

type TextCompletionService interface {
	CompleteTextWithOptions(ctx context.Context, prompt string, opts snips.CompletionOptions) (string, error)
}

// maxGroundingChars caps the server-side result retained for direct client
// rendering and event payloads. It is never used to build a provider prompt.
const maxGroundingChars = 12000

type AnalystGroundingResult struct {
	ToolName       string
	Question       string
	SpecJSON       string
	RowsJSON       string
	Rows           []map[string]any
	RowCount       int
	Truncated      bool
	DurationMs     int64
	EvidenceSource string
	RequestPlan    string
	GeneratedAt    time.Time
}

// buildSpecGenPrompt asks the model to turn a question into a strict JSON query
// spec (no prose) against the allowlisted schema. owner_id is never mentioned —
// it is enforced server-side.
func buildSpecGenPrompt(schema, question, today string) string {
	tomorrow := today
	if parsed, err := time.Parse("2006-01-02", today); err == nil {
		tomorrow = parsed.AddDate(0, 0, 1).Format("2006-01-02")
	}

	var b strings.Builder
	b.WriteString("You convert a business question into a JSON query against a read-only analytics API. ")
	b.WriteString("Output ONLY a single JSON object, nothing else — no prose, no markdown fences.\n\n")
	b.WriteString("Today is ")
	b.WriteString(today)
	b.WriteString(". Tomorrow is ")
	b.WriteString(tomorrow)
	b.WriteString(".\n\nViews and their columns:\n")
	b.WriteString(schema)
	b.WriteString("\nSpec shape:\n")
	b.WriteString(`{"view":"<one view>","columns":["<col>"],"aggregations":[{"func":"sum|count|avg|min|max","col":"<col>","as":"<alias>"}],"where":[{"col":"<col>","op":"=|!=|<|<=|>|>=|like|in","value":<value>,"value_col":"<other-col-if-comparing-columns>"}],"group_by":["<col>"],"order_by":[{"col":"<col-or-alias>","dir":"asc|desc"}],"limit":<int>}`)
	b.WriteString("\n\nRules: exactly one view; only columns listed for that view; never reference owner_id (applied automatically); limit <= 1000. For date filters use 'YYYY-MM-DD'. For today's datetime windows, filter >= today and < tomorrow. For safe column-to-column comparisons, use value_col instead of value. For daily revenue/sales, prefer paid ai_pos_orders total using payment_status='paid' and paid_at; use ai_invoices only when the user specifically asks about invoices or billing.\n\n")
	b.WriteString("Examples:\n")
	b.WriteString(`Q: "revenue today" -> {"view":"ai_pos_orders","aggregations":[{"func":"sum","col":"total","as":"revenue"}],"where":[{"col":"payment_status","op":"=","value":"paid"},{"col":"paid_at","op":">=","value":"` + today + `"},{"col":"paid_at","op":"<","value":"` + tomorrow + `"}]}`)
	b.WriteString("\n")
	b.WriteString(`Q: "how many employees" -> {"view":"ai_employees","aggregations":[{"func":"count","as":"employees"}]}`)
	b.WriteString("\n")
	b.WriteString(`Q: "low stock products" -> {"view":"ai_products","columns":["name","sku","stock_quantity","low_stock_threshold"],"where":[{"col":"is_active","op":"=","value":true},{"col":"track_inventory","op":"=","value":true},{"col":"low_stock_threshold","op":">=","value":0},{"col":"stock_quantity","op":"<=","value_col":"low_stock_threshold"}],"order_by":[{"col":"stock_quantity","dir":"asc"}],"limit":20}`)
	b.WriteString("\n\nIf the question is not answerable from these views (greetings, how-to, general chat), respond with exactly: {}\n\n")
	b.WriteString("Question: ")
	b.WriteString(question)
	return b.String()
}

// extractJSONObject returns the first complete {...} object in s using
// bracket counting, tolerating prose, markdown fences, and trailing text.
func extractJSONObject(s string) string {
	start := strings.Index(s, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	inStr := false
	escaped := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' && inStr {
			escaped = true
			continue
		}
		if c == '"' {
			inStr = !inStr
			continue
		}
		if inStr {
			continue
		}
		switch c {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}

// unquotedOpRe matches LLM-generated "op": = (unquoted operator value).
var unquotedOpRe = regexp.MustCompile(`"op"\s*:\s*(>=|<=|!=|>|<|=|like|in)\s*([,}\]])`)

// repairLLMJSON fixes the most common LLM JSON mistake: an unquoted comparison
// operator used as the value for an "op" field, e.g. "op": = or "op": >=.
func repairLLMJSON(s string) string {
	return unquotedOpRe.ReplaceAllString(s, `"op": "$1"$2`)
}

// analystGrounding runs the analyst pipeline: question -> query spec -> rows.
// On success it returns server-side result metadata and rows for the
// application-owned executor; it never returns a provider prompt containing
// those rows. On failure it returns (nil, false).
func GroundAnalytics(ctx context.Context, ai TextCompletionService, analyst AnalystService, question string, businessID, userID int64, systemContext string, recordUsage func(int)) (*AnalystGroundingResult, bool) {
	if analyst == nil {
		return nil, false
	}
	if grounding, ok := buildDeferredAnalyticsGrounding(question); ok {
		grounding.RequestPlan = ClassifyCaretRequest(question, systemContext).PromptText()
		return grounding, true
	}
	if shouldUseBusinessSnapshot(question) {
		if grounding, ok := BuildBusinessSnapshotGroundingWithContext(ctx, analyst, question, businessID, userID, time.Now().UTC(), systemContext); ok {
			return grounding, true
		}
	}
	if ai == nil {
		return nil, false
	}

	grounding, ok := runAnalyticsQuery(ctx, ai, analyst, question, businessID, userID, systemContext, recordUsage)
	if !ok {
		return nil, false
	}
	return grounding, true
}

// runAnalyticsQuery is the server-side analytics tool implementation. It may
// ask the provider for a query spec using schema-only context, but it never
// places database rows into a provider prompt.
func runAnalyticsQuery(ctx context.Context, ai TextCompletionService, analyst AnalystService, question string, businessID, userID int64, systemContext string, recordUsage func(int)) (*AnalystGroundingResult, bool) {
	if ai == nil || analyst == nil || strings.TrimSpace(question) == "" {
		return nil, false
	}

	started := time.Now()
	today := time.Now().UTC().Format("2006-01-02")
	specPrompt := buildSpecGenPrompt(analyst.SchemaPrompt(), question, today)
	raw, err := ai.CompleteTextWithOptions(ctx, specPrompt, snips.DeterministicJSONCompletionOptions())
	if err != nil {
		logger.Ctx(ctx).Warn("analyst spec generation failed", logger.Error(err))
		return nil, false
	}
	// Spec generation is a separate provider call from the user-facing answer.
	// The provider interface does not expose usage for non-streaming calls, so
	// meter it with the same fallback estimate used for streamed providers.
	if recordUsage != nil {
		recordUsage(EstimateTokens(len(specPrompt) + len(raw)))
	}

	specJSON := repairLLMJSON(extractJSONObject(raw))
	if specJSON == "" || specJSON == "{}" {
		return nil, false
	}

	rows, sql, err := analyst.RunJSON(ctx, businessID, userID, specJSON)
	if err != nil {
		logger.Ctx(ctx).Warn("analyst query failed", logger.Error(err), logger.String("spec", specJSON))
		return nil, false
	}

	logger.Ctx(ctx).Info("analyst query executed",
		logger.String("sql", sql), logger.Int("rows", len(rows)))

	rowsJSON, truncated := marshalGroundingRows(rows, maxGroundingChars)
	plan := ClassifyCaretRequest(question, systemContext)
	return &AnalystGroundingResult{
		ToolName:       "analytics.query",
		Question:       question,
		SpecJSON:       specJSON,
		RowsJSON:       rowsJSON,
		Rows:           rowsPreviewFromJSON(rowsJSON),
		RowCount:       len(rows),
		Truncated:      truncated,
		DurationMs:     time.Since(started).Milliseconds(),
		EvidenceSource: "allowlisted analytics view",
		RequestPlan:    plan.PromptText(),
		GeneratedAt:    time.Now().UTC(),
	}, true
}

func EstimateTokens(chars int) int {
	if chars <= 0 {
		return 0
	}
	return chars / 4
}

func rowsPreviewFromJSON(dataStr string) []map[string]any {
	var rows []map[string]any
	if err := json.Unmarshal([]byte(dataStr), &rows); err != nil {
		return nil
	}
	return rows
}

func marshalGroundingRows(rows []map[string]any, maxChars int) (string, bool) {
	if maxChars <= 2 {
		return "[]", len(rows) > 0
	}

	limit := len(rows)
	for limit > 0 {
		data, err := json.Marshal(rows[:limit])
		if err != nil {
			return "[]", true
		}
		if len(data) <= maxChars {
			return string(data), limit < len(rows)
		}
		limit--
	}

	return "[]", len(rows) > 0
}
