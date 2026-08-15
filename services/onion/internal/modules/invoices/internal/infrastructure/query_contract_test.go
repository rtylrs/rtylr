package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

func TestInvoiceListQueriesUseStableOrdering(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/invoices.sql")
	if err != nil {
		t.Fatalf("read invoice queries: %v", err)
	}
	sql := string(query)
	if !strings.Contains(sql, "ORDER BY created_at DESC, id DESC") {
		t.Fatal("invoice list queries must use a unique created_at tie-breaker")
	}
	if !strings.Contains(sql, "ORDER BY due_date ASC, id ASC") {
		t.Fatal("overdue invoice pagination must use a unique due_date tie-breaker")
	}
}
