package domain

import (
	"strings"
	"testing"
)

func TestBuildSQL_AlwaysBindsOwnerID(t *testing.T) {
	sql, args, err := BuildSQL(QuerySpec{View: "ai_products", Columns: []string{"name", "sku"}}, 42)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !strings.Contains(sql, "owner_id = ?") {
		t.Errorf("owner_id filter missing: %s", sql)
	}
	if len(args) == 0 || args[0] != int64(42) {
		t.Errorf("owner_id arg not bound first: %v", args)
	}
	if !strings.Contains(sql, "FROM ai_products") {
		t.Errorf("wrong from: %s", sql)
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
	for _, want := range []string{"COALESCE(SUM(total), 0) AS revenue", "COUNT(*) AS cnt", "GROUP BY status", "ORDER BY revenue DESC", "owner_id = ?"} {
		if !strings.Contains(sql, want) {
			t.Errorf("missing %q in: %s", want, sql)
		}
	}
	// owner_id + status both bound.
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
	if len(args) != 3 { // owner_id + 2 statuses
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
		t.Errorf("expected owner_id + bound like value, got %v", args)
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
	if len(args) != 2 { // owner_id + track_inventory
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

func TestSchema_ExposesViewsWithoutOwnerID(t *testing.T) {
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
			if c == "owner_id" {
				t.Errorf("%s exposes owner_id, which must stay internal", view)
			}
		}
	}
}
