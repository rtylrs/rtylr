package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

// The lead lists and counts repeat one filter block, since sqlc has no shared
// fragments; a filter added to one query but not the others would make the
// tab counts disagree with the list.
func TestLeadQueriesShareOneFilterBlock(t *testing.T) {
	sql, err := os.ReadFile("../../sql/queries/leads.sql")
	if err != nil {
		t.Fatal(err)
	}
	const first, last = "  AND (CAST(sqlc.arg('filter_reps')", "phone_digits LIKE sqlc.narg('search_digits'))\n"
	blocks := map[string]string{}
	for _, query := range strings.Split(string(sql), "-- name: ")[1:] {
		name := strings.Fields(query)[0]
		if !strings.HasPrefix(name, "ListLeads") && name != "CountLeadsByStatus" {
			continue
		}
		start, end := strings.Index(query, first), strings.Index(query, last)
		if start < 0 || end < start {
			t.Fatalf("%s has no filter block", name)
		}
		blocks[name] = query[start : end+len(last)]
	}
	if len(blocks) != 5 {
		t.Fatalf("found filter blocks in %d queries, want 5", len(blocks))
	}
	want := blocks["ListLeadsNewest"]
	for name, block := range blocks {
		if block != want {
			t.Errorf("%s's filter block differs from ListLeadsNewest's", name)
		}
	}
}
