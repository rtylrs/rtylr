package infrastructure

import (
	"os"
	"strings"
	"testing"
)

func TestBillListQueriesUseStableOrdering(t *testing.T) {
	queryFile, err := os.ReadFile("../../sql/queries/bills.sql")
	if err != nil {
		t.Fatalf("read bill queries: %v", err)
	}

	if got := strings.Count(string(queryFile), "ORDER BY due_date ASC, id ASC"); got != 3 {
		t.Fatalf("stable bill list ordering appears %d times, want 3", got)
	}
}
