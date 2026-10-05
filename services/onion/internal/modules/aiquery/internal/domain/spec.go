package domain

import "errors"

var (
	ErrUnknownView   = errors.New("unknown or disallowed view")
	ErrUnknownColumn = errors.New("unknown or disallowed column")
	ErrBadAggregate  = errors.New("unsupported aggregate function")
	ErrBadOperator   = errors.New("unsupported filter operator")
	ErrEmptySelect   = errors.New("query selects nothing")
	// ErrNoBusinessScope is returned when the query carries no resolvable
	// business. It must be a denial: a zero scope would match the legacy
	// business_id = 0 bucket or nothing at all, and a missing predicate would
	// return every business of the owner.
	ErrNoBusinessScope = errors.New("query has no business scope")
)

const (
	// MaxRows caps any result set the AI can pull back.
	MaxRows = 1000
	// DefaultLimit is applied when the spec omits a limit.
	DefaultLimit = 100
)

// allowedViews is the entire surface the AI may query: read-only ai_* views,
// each mapped to the columns it may select/filter/group by. owner_id and
// business_id are intentionally absent — neither is ever selectable, and
// business_id is always injected as the tenant filter, so a spec cannot narrow
// or widen its own scope. Keep this in lockstep with
// sql/ai_views.sql and sql/zz_ai_views_business_scope.sql.
var allowedViews = map[string]map[string]struct{}{
	"ai_products":   set("id", "name", "sku", "category_id", "brand_id", "supplier_id", "cost_price", "sale_price", "compare_at_price", "currency", "stock_quantity", "low_stock_threshold", "reorder_quantity", "is_active", "track_inventory", "lead_time_days", "created_at", "updated_at"),
	"ai_customers":  set("id", "first_name", "last_name", "email", "phone", "city", "country", "loyalty_points", "total_spent", "visit_count", "last_visit_at", "loyalty_tier", "is_active", "created_at", "updated_at"),
	"ai_invoices":   set("id", "invoice_number", "client_name", "status", "issue_date", "due_date", "paid_date", "currency", "subtotal", "discount_amount", "tax_rate", "tax_amount", "total", "amount_paid", "amount_due", "created_at", "updated_at"),
	"ai_expenses":   set("id", "category_id", "description", "amount", "currency", "expense_date", "vendor", "payment_method", "is_tax_deductible", "tax_amount", "status", "created_at", "updated_at"),
	"ai_pos_orders": set("id", "order_number", "type", "status", "customer_id", "customer_name", "staff_id", "location_id", "subtotal", "discount_amount", "tax_amount", "total", "tip_amount", "payment_method", "payment_status", "created_at", "paid_at"),
	"ai_employees":  set("id", "first_name", "last_name", "email", "department_id", "location_id", "employment_type", "status", "salary", "hourly_rate", "overtime_rate", "start_date", "end_date", "created_at", "updated_at"),
	"ai_tasks":      set("id", "project_id", "workspace_id", "title", "status", "priority", "assignee_id", "due_date", "start_date", "completed_at", "estimated_hours", "actual_hours", "created_at", "updated_at"),
}

// Aggregation is e.g. {"func":"sum","col":"total","as":"revenue"}. For count,
// col may be empty (COUNT(*)).
type Aggregation struct {
	Func string
	Col  string
	As   string
}

// Filter is a single WHERE condition. Value is a scalar, or a slice for "in".
type Filter struct {
	Col         string
	Op          string
	Value       any
	ValueColumn string
}

// OrderBy references a selected column or an aggregation alias.
type OrderBy struct {
	Col string
	Dir string
}

// QuerySpec is the structured query the LLM produces.
type QuerySpec struct {
	View         string
	Columns      []string
	Aggregations []Aggregation
	Where        []Filter
	GroupBy      []string
	OrderBy      []OrderBy
	Limit        int
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

func set(items ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(items))
	for _, item := range items {
		m[item] = struct{}{}
	}
	return m
}
