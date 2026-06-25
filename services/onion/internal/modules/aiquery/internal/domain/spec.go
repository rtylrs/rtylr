// Package domain defines the safe, structured query layer the AI uses to read
// business data. The LLM never emits raw SQL — it emits a QuerySpec, which we
// validate against an allowlist of read-only views/columns and compile to a
// parameterized statement that ALWAYS binds owner_id. This makes a missing
// tenant filter (a cross-tenant data leak) impossible by construction.
package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

var (
	ErrUnknownView   = errors.New("unknown or disallowed view")
	ErrUnknownColumn = errors.New("unknown or disallowed column")
	ErrBadAggregate  = errors.New("unsupported aggregate function")
	ErrBadOperator   = errors.New("unsupported filter operator")
	ErrEmptySelect   = errors.New("query selects nothing")
)

const (
	// MaxRows caps any result set the AI can pull back.
	MaxRows = 1000
	// DefaultLimit is applied when the spec omits a limit.
	DefaultLimit = 100
)

// allowedViews is the entire surface the AI may query: read-only ai_* views,
// each mapped to the columns it may select/filter/group by. owner_id is
// intentionally absent — it is never selectable and is always injected as the
// tenant filter. Keep this in lockstep with shared/ai_views.sql.
var allowedViews = map[string]map[string]struct{}{
	"ai_products":   set("id", "name", "sku", "category_id", "brand_id", "supplier_id", "cost_price", "sale_price", "compare_at_price", "currency", "stock_quantity", "low_stock_threshold", "reorder_quantity", "is_active", "track_inventory", "lead_time_days", "created_at", "updated_at"),
	"ai_customers":  set("id", "first_name", "last_name", "email", "phone", "city", "country", "loyalty_points", "total_spent", "visit_count", "last_visit_at", "loyalty_tier", "is_active", "created_at", "updated_at"),
	"ai_invoices":   set("id", "invoice_number", "client_name", "status", "issue_date", "due_date", "paid_date", "currency", "subtotal", "discount_amount", "tax_rate", "tax_amount", "total", "amount_paid", "amount_due", "created_at", "updated_at"),
	"ai_expenses":   set("id", "category_id", "description", "amount", "currency", "expense_date", "vendor", "payment_method", "is_tax_deductible", "tax_amount", "status", "created_at", "updated_at"),
	"ai_pos_orders": set("id", "order_number", "type", "status", "customer_id", "customer_name", "staff_id", "location_id", "subtotal", "discount_amount", "tax_amount", "total", "tip_amount", "payment_method", "payment_status", "created_at", "paid_at"),
	"ai_employees":  set("id", "first_name", "last_name", "email", "department_id", "location_id", "employment_type", "status", "salary", "hourly_rate", "overtime_rate", "start_date", "end_date", "created_at", "updated_at"),
	"ai_tasks":      set("id", "project_id", "workspace_id", "title", "status", "priority", "assignee_id", "due_date", "start_date", "completed_at", "estimated_hours", "actual_hours", "created_at", "updated_at"),
}

var (
	allowedAggregates = set("sum", "count", "avg", "min", "max")
	allowedOperators  = set("=", "!=", "<", "<=", ">", ">=", "like", "in")
	aliasRe           = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

// Aggregation is e.g. {"func":"sum","col":"total","as":"revenue"}. For count,
// col may be empty (COUNT(*)).
type Aggregation struct {
	Func string `json:"func"`
	Col  string `json:"col,omitempty"`
	As   string `json:"as,omitempty"`
}

// Filter is a single WHERE condition. Value is a scalar, or a slice for "in".
type Filter struct {
	Col         string `json:"col"`
	Op          string `json:"op"`
	Value       any    `json:"value,omitempty"`
	ValueColumn string `json:"value_col,omitempty"`
}

// OrderBy references a selected column or an aggregation alias.
type OrderBy struct {
	Col string `json:"col"`
	Dir string `json:"dir,omitempty"`
}

// QuerySpec is the structured query the LLM produces.
type QuerySpec struct {
	View         string        `json:"view"`
	Columns      []string      `json:"columns,omitempty"`
	Aggregations []Aggregation `json:"aggregations,omitempty"`
	Where        []Filter      `json:"where,omitempty"`
	GroupBy      []string      `json:"group_by,omitempty"`
	OrderBy      []OrderBy     `json:"order_by,omitempty"`
	Limit        int           `json:"limit,omitempty"`
}

// Schema returns the allowlisted views and their columns, for prompting the LLM.
func Schema() map[string][]string {
	out := make(map[string][]string, len(allowedViews))
	for view, cols := range allowedViews {
		list := make([]string, 0, len(cols))
		for c := range cols {
			list = append(list, c)
		}
		out[view] = list
	}
	return out
}

// BuildSQL validates the spec against the allowlist and compiles it to a
// parameterized SELECT that is always constrained to ownerID. Returns the SQL
// and its bound args. Any reference outside the allowlist is rejected.
func BuildSQL(spec QuerySpec, ownerID int64) (string, []any, error) {
	cols, ok := allowedViews[spec.View]
	if !ok {
		return "", nil, fmt.Errorf("%w: %q", ErrUnknownView, spec.View)
	}

	selects, err := buildSelects(spec, cols)
	if err != nil {
		return "", nil, err
	}

	// owner_id is always bound — the core isolation guarantee.
	b := sq.Select(selects...).From(spec.View).Where(sq.Eq{"owner_id": ownerID})

	for _, f := range spec.Where {
		pred, perr := filterPredicate(f, cols)
		if perr != nil {
			return "", nil, perr
		}
		b = b.Where(pred)
	}

	for _, g := range spec.GroupBy {
		if !allowed(cols, g) {
			return "", nil, fmt.Errorf("%w: group by %q", ErrUnknownColumn, g)
		}
		b = b.GroupBy(g)
	}

	aliases := aggregationAliases(spec)
	for _, o := range spec.OrderBy {
		if !allowed(cols, o.Col) && !aliases[o.Col] {
			return "", nil, fmt.Errorf("%w: order by %q", ErrUnknownColumn, o.Col)
		}
		dir := "ASC"
		if strings.EqualFold(o.Dir, "desc") {
			dir = "DESC"
		}
		b = b.OrderBy(o.Col + " " + dir)
	}

	limit := spec.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxRows {
		limit = MaxRows
	}
	b = b.Limit(uint64(limit))

	return b.PlaceholderFormat(sq.Question).ToSql()
}

func buildSelects(spec QuerySpec, cols map[string]struct{}) ([]string, error) {
	var selects []string

	for _, c := range spec.Columns {
		if !allowed(cols, c) {
			return nil, fmt.Errorf("%w: %q", ErrUnknownColumn, c)
		}
		selects = append(selects, c)
	}

	for _, a := range spec.Aggregations {
		fn := strings.ToLower(strings.TrimSpace(a.Func))
		if !inSet(allowedAggregates, fn) {
			return nil, fmt.Errorf("%w: %q", ErrBadAggregate, a.Func)
		}
		alias := sanitizeAlias(a.As)
		if fn == "count" && a.Col == "" {
			selects = append(selects, fmt.Sprintf("COUNT(*) AS %s", orDefault(alias, "count")))
			continue
		}
		if !allowed(cols, a.Col) {
			return nil, fmt.Errorf("%w: aggregate on %q", ErrUnknownColumn, a.Col)
		}
		selects = append(selects, fmt.Sprintf("%s AS %s", aggregateExpr(fn, a.Col), orDefault(alias, fn+"_"+a.Col)))
	}

	if len(selects) == 0 {
		// Default to all allowed columns rather than erroring on a bare view.
		for c := range cols {
			selects = append(selects, c)
		}
		if len(selects) == 0 {
			return nil, ErrEmptySelect
		}
	}

	return selects, nil
}

func aggregateExpr(fn, col string) string {
	switch fn {
	case "sum":
		return fmt.Sprintf("COALESCE(SUM(%s), 0)", col)
	default:
		return fmt.Sprintf("%s(%s)", strings.ToUpper(fn), col)
	}
}

func filterPredicate(f Filter, cols map[string]struct{}) (sq.Sqlizer, error) {
	if !allowed(cols, f.Col) {
		return nil, fmt.Errorf("%w: filter on %q", ErrUnknownColumn, f.Col)
	}
	op := strings.ToLower(strings.TrimSpace(f.Op))
	if !inSet(allowedOperators, op) {
		return nil, fmt.Errorf("%w: %q", ErrBadOperator, f.Op)
	}
	if strings.TrimSpace(f.ValueColumn) != "" {
		valueCol := strings.TrimSpace(f.ValueColumn)
		if !allowed(cols, valueCol) {
			return nil, fmt.Errorf("%w: value column %q", ErrUnknownColumn, valueCol)
		}
		if op == "in" || op == "like" {
			return nil, fmt.Errorf("%w: column comparison with %q", ErrBadOperator, f.Op)
		}
		return sq.Expr(fmt.Sprintf("%s %s %s", strings.TrimSpace(f.Col), op, valueCol)), nil
	}

	switch op {
	case "=":
		return sq.Eq{f.Col: f.Value}, nil
	case "!=":
		return sq.NotEq{f.Col: f.Value}, nil
	case "<":
		return sq.Lt{f.Col: f.Value}, nil
	case "<=":
		return sq.LtOrEq{f.Col: f.Value}, nil
	case ">":
		return sq.Gt{f.Col: f.Value}, nil
	case ">=":
		return sq.GtOrEq{f.Col: f.Value}, nil
	case "like":
		return sq.Like{f.Col: f.Value}, nil
	case "in":
		return sq.Eq{f.Col: f.Value}, nil // squirrel renders a slice value as IN (...)
	}
	return nil, fmt.Errorf("%w: %q", ErrBadOperator, f.Op)
}

func aggregationAliases(spec QuerySpec) map[string]bool {
	out := map[string]bool{}
	for _, a := range spec.Aggregations {
		alias := sanitizeAlias(a.As)
		if alias == "" {
			if strings.ToLower(a.Func) == "count" && a.Col == "" {
				alias = "count"
			} else {
				alias = strings.ToLower(a.Func) + "_" + a.Col
			}
		}
		out[alias] = true
	}
	return out
}

func set(items ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(items))
	for _, i := range items {
		m[i] = struct{}{}
	}
	return m
}

func inSet(m map[string]struct{}, k string) bool { _, ok := m[k]; return ok }
func allowed(cols map[string]struct{}, c string) bool {
	_, ok := cols[strings.TrimSpace(c)]
	return ok
}

func sanitizeAlias(s string) string { return aliasRe.ReplaceAllString(strings.TrimSpace(s), "") }
func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
