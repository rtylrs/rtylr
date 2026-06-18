package handlers

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
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

// maxGroundingChars caps the retrieved data folded into the prompt.
const maxGroundingChars = 12000

// buildSpecGenPrompt asks the model to turn a question into a strict JSON query
// spec (no prose) against the allowlisted schema. owner_id is never mentioned —
// it is enforced server-side.
func buildSpecGenPrompt(schema, question, today string) string {
	var b strings.Builder
	b.WriteString("You convert a business question into a JSON query against a read-only analytics API. ")
	b.WriteString("Output ONLY a single JSON object, nothing else — no prose, no markdown fences.\n\n")
	b.WriteString("Today is ")
	b.WriteString(today)
	b.WriteString(".\n\nViews and their columns:\n")
	b.WriteString(schema)
	b.WriteString("\nSpec shape:\n")
	b.WriteString(`{"view":"<one view>","columns":["<col>"],"aggregations":[{"func":"sum|count|avg|min|max","col":"<col>","as":"<alias>"}],"where":[{"col":"<col>","op":"=|!=|<|<=|>|>=|like|in","value":<value>}],"group_by":["<col>"],"order_by":[{"col":"<col-or-alias>","dir":"asc|desc"}],"limit":<int>}`)
	b.WriteString("\n\nRules: exactly one view; only columns listed for that view; never reference owner_id (applied automatically); limit <= 1000. For date filters use 'YYYY-MM-DD'.\n\n")
	b.WriteString("Examples:\n")
	b.WriteString(`Q: "revenue today" -> {"view":"ai_invoices","aggregations":[{"func":"sum","col":"total","as":"revenue"}],"where":[{"col":"issue_date","op":">=","value":"` + today + `"}]}`)
	b.WriteString("\n")
	b.WriteString(`Q: "how many employees" -> {"view":"ai_employees","aggregations":[{"func":"count","as":"employees"}]}`)
	b.WriteString("\n")
	b.WriteString(`Q: "low stock products" -> {"view":"ai_products","columns":["name","sku","stock_quantity","low_stock_threshold"],"order_by":[{"col":"stock_quantity","dir":"asc"}],"limit":20}`)
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
// On success it returns a data-grounded prompt prefix and true; on any failure
// it returns ("", false) so the caller falls back to a normal chat answer.
func (h *GenerateTextStreamHandler) analystGrounding(ctx context.Context, question string, businessID, userID int64) (string, bool) {
	if h.AnalystService == nil || h.AiService == nil {
		return "", false
	}

	today := time.Now().UTC().Format("2006-01-02")
	raw, err := h.AiService.CompleteText(ctx, buildSpecGenPrompt(h.AnalystService.SchemaPrompt(), question, today))
	if err != nil {
		logger.Ctx(ctx).Warn("analyst spec generation failed", logger.Error(err))
		return "", false
	}

	specJSON := repairLLMJSON(extractJSONObject(raw))
	if specJSON == "" || specJSON == "{}" {
		return "", false
	}

	rows, sql, err := h.AnalystService.RunJSON(ctx, businessID, userID, specJSON)
	if err != nil {
		logger.Ctx(ctx).Warn("analyst query failed", logger.Error(err), logger.String("spec", specJSON))
		return "", false
	}

	data, err := json.Marshal(rows)
	if err != nil {
		return "", false
	}
	dataStr := string(data)
	if len(dataStr) > maxGroundingChars {
		dataStr = dataStr[:maxGroundingChars]
	}

	logger.Ctx(ctx).Info("analyst query executed",
		logger.String("sql", sql), logger.Int("rows", len(rows)))

	var g strings.Builder
	g.WriteString("You are answering an analytics question about the user's own business. ")
	g.WriteString("The following rows were retrieved from their database for this question:\n```json\n")
	g.WriteString(dataStr)
	g.WriteString("\n```\nAnswer the question using ONLY this data. Be concise; use a small table or bullets when it helps. If the data is empty, say so plainly.")
	return g.String(), true
}
