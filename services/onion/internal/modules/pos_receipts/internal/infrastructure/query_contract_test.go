package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

func TestReceiptListQueriesHaveServerBounds(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/pos_receipts.sql")
	if err != nil {
		t.Fatalf("read receipt queries: %v", err)
	}
	if got := strings.Count(string(query), "LIMIT 500"); got != 3 {
		t.Fatalf("expected three bounded receipt list queries, found %d", got)
	}
}
