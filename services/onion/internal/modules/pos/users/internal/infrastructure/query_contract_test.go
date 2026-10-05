package infrastructure

import (
	"os"
	"strings"
	"testing"
)

func TestPOSUserListQueriesDoNotSelectPINCredentials(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/pos_users.sql")
	if err != nil {
		t.Fatalf("read POS-user queries: %v", err)
	}

	source := string(query)
	for _, name := range []string{"ListPOSUsersByOwner", "ListPOSUsersByLocation"} {
		start := strings.Index(source, "-- name: "+name)
		if start < 0 {
			t.Fatalf("query %s not found", name)
		}
		end := strings.Index(source[start+1:], "-- name:")
		if end < 0 {
			end = len(source) - start - 1
		}
		section := source[start : start+1+end]
		if strings.Contains(strings.ToLower(section), "pin_hash") || strings.Contains(strings.ToLower(section), " pin") {
			t.Errorf("query %s carries PIN credential columns", name)
		}
	}
}
