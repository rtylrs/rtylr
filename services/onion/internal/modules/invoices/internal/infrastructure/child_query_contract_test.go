package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

func TestInvoiceChildListsHaveServerBounds(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/invoices.sql")
	if err != nil {
		t.Fatalf("read invoice queries: %v", err)
	}
	if got := strings.Count(string(query), "LIMIT 500"); got != 2 {
		t.Fatalf("expected two bounded invoice child lists, found %d", got)
	}
}
