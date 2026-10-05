package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

func TestReviewListQueryHasServerBoundAndStableOrder(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/reviews.sql")
	if err != nil {
		t.Fatalf("read review queries: %v", err)
	}
	sql := string(query)
	if !strings.Contains(sql, "ORDER BY created_at DESC, id DESC") || !strings.Contains(sql, "LIMIT 500") {
		t.Fatal("review history must have a server-side bound and unique ordering")
	}
}
