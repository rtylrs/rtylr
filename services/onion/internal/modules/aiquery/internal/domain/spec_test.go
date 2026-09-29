package domain

import (
	"strings"
	"testing"
)

func TestBuildSQL_AlwaysBindsBusinessID(t *testing.T) {
	sql, args, err := BuildSQL(QuerySpec{View: "ai_products", Columns: []string{"name", "sku"}}, 42)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(sql, "business_id = ?") {
		t.Errorf("business_id filter missing: %s", sql)
	}
	if strings.Contains(sql, "owner_id") {
		t.Errorf("owner_id must not scope tenant data, it is shared by every business of an account: %s", sql)
	}
	if len(args) == 0 || args[0] != int64(42) {
		t.Errorf("business_id arg not bound first: %v", args)
	}
	if !strings.Contains(sql, "FROM ai_products") {
		t.Errorf("wrong from: %s", sql)
	}
}

// owner_id is shared by every business of one account, so an owner-scoped query
// hands a member of business B1 the invoices of their sibling business B2. The
// bound predicate has to be the business.
func TestBuildSQL_InvoiceAggregateStaysInsideOneBusiness(t *testing.T) {
	sql, args, err := BuildSQL(QuerySpec{
		View:         "ai_invoices",
		Columns:      []string{"status"},
		Aggregations: []Aggregation{{Func: "count", As: "n"}},
		GroupBy:      []string{"status"},
	}, 4242)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(sql, "FROM ai_invoices WHERE business_id = ?") {
		t.Errorf("invoice query is not business scoped: %s", sql)
	}
	if len(args) != 1 || args[0] != int64(4242) {
		t.Errorf("expected the caller's business as the only bound arg, got %v", args)
	}
}

func TestBuildSQL_RejectsMissingBusinessScope(t *testing.T) {
	for _, businessID := range []int64{0, -1, -4242} {
		if _, _, err := BuildSQL(QuerySpec{View: "ai_invoices", Columns: []string{"id"}}, businessID); err == nil {
			t.Errorf("businessID %d: expected rejection, an unscoped query returns every business of the owner", businessID)
		}
	}
}

// The business predicate is injected, not negotiated: a spec — LLM-authored or
// otherwise — must not be able to name business_id and narrow or widen the
// scope it was already given.
func TestBuildSQL_RejectsSpecSuppliedBusinessScope(t *testing.T) {
	cases := []QuerySpec{
		{View: "ai_invoices", Columns: []string{"business_id"}},
		{View: "ai_invoices", Where: []Filter{{Col: "business_id", Op: "=", Value: 2}}},
		{View: "ai_invoices", Where: []Filter{{Col: "status", Op: "=", ValueColumn: "business_id"}}},
		{View: "ai_invoices", GroupBy: []string{"business_id"}},
		{View: "ai_invoices", OrderBy: []OrderBy{{Col: "business_id"}}},
		{View: "ai_employees", Aggregations: []Aggregation{{Func: "max", Col: "business_id"}}},
	}
	for i, spec := range cases {
		if _, _, err := BuildSQL(spec, 1); err == nil {
			t.Errorf("case %d: expected rejection of a spec naming business_id", i)
		}
	}
}

func TestBuildSQL_RejectsUnknownView(t *testing.T) {
	if _, _, err := BuildSQL(QuerySpec{View: "users"}, 1); err == nil {
		t.Fatal("expected rejection of non-allowlisted view")
	}
	if _, _, err := BuildSQL(QuerySpec{View: "erp_products"}, 1); err == nil {
		t.Fatal("expected rejection of base table")
	}
}

func TestBuildSQL_RejectsUnknownColumns(t *testing.T) {
	cases := []QuerySpec{
		{View: "ai_products", Columns: []string{"password"}},
		{View: "ai_products", Where: []Filter{{Col: "secret", Op: "=", Value: 1}}},
		{View: "ai_products", GroupBy: []string{"nope"}},
		{View: "ai_products", OrderBy: []OrderBy{{Col: "nope"}}},
		{View: "ai_products", Aggregations: []Aggregation{{Func: "sum", Col: "evil"}}},
	}
	for i, c := range cases {
		if _, _, err := BuildSQL(c, 1); err == nil {
			t.Errorf("case %d: expected unknown-column rejection", i)
		}
	}
}

func TestBuildSQL_RejectsBadAggregateAndOperator(t *testing.T) {
	if _, _, err := BuildSQL(QuerySpec{View: "ai_invoices", Aggregations: []Aggregation{{Func: "drop", Col: "total"}}}, 1); err == nil {
		t.Error("expected bad-aggregate rejection")
	}
	if _, _, err := BuildSQL(QuerySpec{View: "ai_invoices", Where: []Filter{{Col: "total", Op: "; DROP", Value: 1}}}, 1); err == nil {
		t.Error("expected bad-operator rejection")
	}
}

func TestBuildSQL_AggregatesAndGroupBy(t *testing.T) {
	sql, args, err := BuildSQL(QuerySpec{
		View:         "ai_invoices",
		Columns:      []string{"status"},
		Aggregations: []Aggregation{{Func: "sum", Col: "total", As: "revenue"}, {Func: "count", As: "cnt"}},
		Where:        []Filter{{Col: "status", Op: "=", Value: "paid"}},
		GroupBy:      []string{"status"},
		OrderBy:      []OrderBy{{Col: "revenue", Dir: "desc"}},
	}, 7)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	for _, want := range []string{"COALESCE(SUM(total), 0) AS revenue", "COUNT(*) AS cnt", "GROUP BY status", "ORDER BY revenue DESC", "business_id = ?"} {
		if !strings.Contains(sql, want) {
			t.Errorf("missing %q in: %s", want, sql)
		}
	}
	// business_id + status both bound.
	if len(args) != 2 {
		t.Errorf("expected 2 args (owner_id, status), got %v", args)
	}
}

func TestBuildSQL_SumAggregatesDefaultToZero(t *testing.T) {
	sql, _, err := BuildSQL(QuerySpec{
		View:         "ai_pos_orders",
		Aggregations: []Aggregation{{Func: "sum", Col: "total", As: "revenue"}},
	}, 7)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(sql, "COALESCE(SUM(total), 0) AS revenue") {
		t.Fatalf("expected SUM aggregate to default null to zero, got: %s", sql)
	}
}

func TestBuildSQL_ClampsLimit(t *testing.T) {
	sql, _, err := BuildSQL(QuerySpec{View: "ai_tasks", Limit: 999999}, 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(sql, "LIMIT 1000") {
		t.Errorf("limit not clamped to MaxRows: %s", sql)
	}

	sql2, _, _ := BuildSQL(QuerySpec{View: "ai_tasks"}, 1)
	if !strings.Contains(sql2, "LIMIT 100") {
		t.Errorf("default limit not applied: %s", sql2)
	}
}

func TestBuildSQL_InFilterRendersIn(t *testing.T) {
	sql, args, err := BuildSQL(QuerySpec{
		View:  "ai_pos_orders",
		Where: []Filter{{Col: "status", Op: "in", Value: []string{"paid", "refunded"}}},
	}, 5)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(strings.ToUpper(sql), "IN (?,?)") {
		t.Errorf("expected IN rendering: %s", sql)
	}
	if len(args) != 3 { // business_id + 2 statuses
		t.Errorf("expected 3 args, got %v", args)
	}
}

func TestBuildSQL_LikeFilterRendersBoundLike(t *testing.T) {
	sql, args, err := BuildSQL(QuerySpec{
		View:    "ai_products",
		Columns: []string{"name", "sku"},
		Where:   []Filter{{Col: "name", Op: "like", Value: "%whey%"}},
	}, 5)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(sql, "name LIKE ?") {
		t.Errorf("expected LIKE rendering: %s", sql)
	}
	if len(args) != 2 || args[0] != int64(5) || args[1] != "%whey%" {
		t.Errorf("expected business_id + bound like value, got %v", args)
	}
}

func TestBuildSQL_AllowsSafeColumnComparisonFilters(t *testing.T) {
	sql, args, err := BuildSQL(QuerySpec{
		View: "ai_products",
		Aggregations: []Aggregation{{
			Func: "count",
			As:   "low_stock_products",
		}},
		Where: []Filter{
			{Col: "track_inventory", Op: "=", Value: true},
			{Col: "stock_quantity", Op: "<=", ValueColumn: "low_stock_threshold"},
		},
	}, 9)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(sql, "stock_quantity <= low_stock_threshold") {
		t.Fatalf("expected column comparison in SQL, got: %s", sql)
	}
	if len(args) != 2 { // business_id + track_inventory
		t.Fatalf("expected only bound scalar args, got %v", args)
	}
}

func TestBuildSQL_RejectsUnsafeColumnComparisonFilters(t *testing.T) {
	cases := []QuerySpec{
		{
			View:  "ai_products",
			Where: []Filter{{Col: "stock_quantity", Op: "<=", ValueColumn: "password"}},
		},
		{
			View:  "ai_products",
			Where: []Filter{{Col: "name", Op: "like", ValueColumn: "sku"}},
		},
		{
			View:  "ai_products",
			Where: []Filter{{Col: "name", Op: "in", ValueColumn: "sku"}},
		},
	}
	for i, spec := range cases {
		if _, _, err := BuildSQL(spec, 1); err == nil {
			t.Fatalf("case %d: expected unsafe column comparison rejection", i)
		}
	}
}

func TestSchema_ExposesViewsWithoutTenantColumns(t *testing.T) {
	s := Schema()
	wantViews := map[string]struct{}{
		"ai_products":   {},
		"ai_customers":  {},
		"ai_invoices":   {},
		"ai_expenses":   {},
		"ai_pos_orders": {},
		"ai_employees":  {},
		"ai_tasks":      {},
	}
	if len(s) != len(wantViews) {
		t.Errorf("expected %d views, got %d", len(wantViews), len(s))
	}
	for view := range wantViews {
		if _, ok := s[view]; !ok {
			t.Errorf("expected allowlisted view %q", view)
		}
	}
	for view, cols := range s {
		for _, c := range cols {
			if c == "owner_id" || c == "business_id" {
				t.Errorf("%s exposes %s, which must stay internal", view, c)
			}
		}
	}
}
