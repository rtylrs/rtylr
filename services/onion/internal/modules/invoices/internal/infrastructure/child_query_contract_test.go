package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

// The child lists of an invoice are unbounded by nature: an invoice can accrue
// any number of items or payments, so these two reads must carry a server-side
// cap rather than trusting the caller to page.
//
// This asserts a bound on each named query rather than counting "LIMIT 500"
// across the file. The count form failed the moment #259 correctly bounded
// ListInvoicesByOwnerAndStatus, which made the test punish exactly the change it
// exists to encourage.
func TestInvoiceChildListsHaveServerBounds(t *testing.T) {
	const queryFile = "../../sql/queries/invoices.sql"

	raw, err := os.ReadFile(queryFile)
	if err != nil {
		t.Fatalf("read invoice queries: %v", err)
	}

	for _, name := range []string{"GetInvoiceItems", "GetPaymentsByInvoice"} {
		t.Run(name, func(t *testing.T) {
			body, ok := namedQuery(string(raw), name)
			if !ok {
				t.Fatalf("query %s not found in %s; rename the test or restore the bound", name, queryFile)
			}
			if !strings.Contains(body, "LIMIT") {
				t.Fatalf("query %s has no LIMIT, so one invoice with enough children is an unbounded read:\n%s", name, body)
			}
		})
	}
}

// namedQuery returns the body of a sqlc query, from its "-- name: X" marker to
// the next marker or end of file.
func namedQuery(source, name string) (string, bool) {
	marker := "-- name: " + name + " "
	start := strings.Index(source, marker)
	if start < 0 {
		return "", false
	}
	rest := source[start+len(marker):]
	if next := strings.Index(rest, "-- name: "); next >= 0 {
		return rest[:next], true
	}
	return rest, true
}
