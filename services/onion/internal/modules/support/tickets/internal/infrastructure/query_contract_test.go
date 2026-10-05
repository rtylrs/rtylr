package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

func TestFindSupportCustomerByEmailHasStableOrder(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/support_tickets.sql")
	if err != nil {
		t.Fatalf("read support-ticket queries: %v", err)
	}

	sql := string(query)
	start := strings.Index(sql, "-- name: FindSupportCustomerByEmail")
	if start < 0 {
		t.Fatal("FindSupportCustomerByEmail query is missing")
	}
	queryMarker := "-- name:"
	end := strings.Index(sql[start+len(queryMarker):], queryMarker)
	if end < 0 {
		end = len(sql) - start
	} else {
		end += len(queryMarker)
	}
	queryText := sql[start : start+end]
	if !strings.Contains(queryText, "ORDER BY id ASC") {
		t.Fatal("FindSupportCustomerByEmail must order by id for deterministic selection")
	}
}
