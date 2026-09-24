// Package segmentrules evaluates stored CRM segment filters against the
// crm_customers table.
//
// A segment's filters are persisted as JSON of the shape
//
//	{"rules": ["totalSpent > 500", "visits > 5", "country contains US"]}
//
// produced by the CRM segment builder. Each rule is "<field> <operator>
// <value>"; every rule must match (AND). Filter returns a parameterized WHERE
// fragment joined with AND plus its args, or empty values when no valid rule
// is present (callers then fall back to their unfiltered behavior).
package segmentrules

import (
	"encoding/json"
	"fmt"
	"strings"
)

type fieldSpec struct {
	column    string
	numeric   bool
	boolean   bool
	jsonArray bool
}

var fields = map[string]fieldSpec{
	"totalspent":    {column: "total_spent", numeric: true},
	"visits":        {column: "visit_count", numeric: true},
	"visitcount":    {column: "visit_count", numeric: true},
	"points":        {column: "loyalty_points", numeric: true},
	"loyaltypoints": {column: "loyalty_points", numeric: true},
	"tier":          {column: "loyalty_tier"},
	"loyaltytier":   {column: "loyalty_tier"},
	"email":         {column: "email"},
	"phone":         {column: "phone"},
	"city":          {column: "city"},
	"country":       {column: "country"},
	"firstname":     {column: "first_name"},
	"lastname":      {column: "last_name"},
	"isactive":      {column: "is_active", boolean: true},
	"tags":          {column: "tags", jsonArray: true},
}

type parsedRule struct {
	spec      fieldSpec
	op        string
	val       string
	boolValue bool
}

type rulesDoc struct {
	Rules []string `json:"rules"`
}

// Filter converts stored segment rules JSON into a SQL WHERE fragment and
// args. The fragment is safe to AND into any query on crm_customers: values
// are always passed as parameters, identifiers come only from the fixed
// column map, and LIKE wildcards inside user values are escaped.
func Filter(rulesJSON []byte) (string, []any) {
	if len(rulesJSON) == 0 {
		return "", nil
	}

	var doc rulesDoc
	if err := json.Unmarshal(rulesJSON, &doc); err != nil {
		return "", nil
	}

	var conds []string
	var args []any
	for _, raw := range doc.Rules {
		rule, ok := parseRule(raw)
		if !ok {
			continue
		}
		frag, rargs := rule.sql()
		if frag == "" {
			continue
		}
		conds = append(conds, frag)
		args = append(args, rargs...)
	}

	return strings.Join(conds, " AND "), args
}

func parseRule(raw string) (parsedRule, bool) {
	parts := strings.Fields(strings.TrimSpace(raw))
	if len(parts) < 3 {
		return parsedRule{}, false
	}
	field := strings.NewReplacer("_", "", "-", "").Replace(strings.ToLower(parts[0]))
	spec, ok := fields[field]
	if !ok {
		return parsedRule{}, false
	}
	op := strings.ToLower(parts[1])
	value := strings.Join(parts[2:], " ")
	if spec.numeric && !isNumeric(value) {
		return parsedRule{}, false
	}
	boolValue := false
	if spec.boolean {
		boolValue = normalizeBool(value)
	}
	return parsedRule{spec: spec, op: op, val: value, boolValue: boolValue}, true
}

func isNumeric(value string) bool {
	if value == "" {
		return false
	}
	dot := false
	for i, ch := range value {
		switch {
		case ch >= '0' && ch <= '9':
		case ch == '.' && !dot:
			dot = true
		case ch == '-' && i == 0:
		default:
			return false
		}
	}
	return true
}

func normalizeBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}

func likeEscape(value string) string {
	replacer := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_")
	return replacer.Replace(value)
}

func (r parsedRule) valueArg() any {
	if r.spec.boolean {
		return r.boolValue
	}
	return r.val
}

func (r parsedRule) sql() (string, []any) {
	col := r.spec.column
	switch r.op {
	case ">", ">=", "<", "<=":
		if !r.spec.numeric {
			break
		}
		return fmt.Sprintf("%s %s ?", col, r.op), []any{r.val}
	case "=", "==":
		if r.spec.jsonArray {
			encoded, err := json.Marshal(r.val)
			if err != nil {
				break
			}
			return "JSON_CONTAINS(tags, ?)", []any{string(encoded)}
		}
		return fmt.Sprintf("%s = ?", col), []any{r.valueArg()}
	case "!=", "<>":
		return fmt.Sprintf("%s <> ?", col), []any{r.valueArg()}
	case "contains":
		if r.spec.numeric || r.spec.boolean {
			break
		}
		if r.spec.jsonArray {
			encoded, err := json.Marshal(r.val)
			if err != nil {
				break
			}
			return "JSON_CONTAINS(tags, ?)", []any{string(encoded)}
		}
		return fmt.Sprintf("%s LIKE ? ESCAPE '!'", col), []any{"%" + likeEscape(r.val) + "%"}
	case "starts", "startswith", "starts_with":
		if r.spec.numeric || r.spec.boolean || r.spec.jsonArray {
			break
		}
		return fmt.Sprintf("%s LIKE ? ESCAPE '!'", col), []any{likeEscape(r.val) + "%"}
	case "ends", "endswith", "ends_with":
		if r.spec.numeric || r.spec.boolean || r.spec.jsonArray {
			break
		}
		return fmt.Sprintf("%s LIKE ? ESCAPE '!'", col), []any{"%" + likeEscape(r.val)}
	}
	return "", nil
}
