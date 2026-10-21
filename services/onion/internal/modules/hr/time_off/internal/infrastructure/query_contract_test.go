package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

func TestTimeOffListQueryHasServerBoundAndStableOrder(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/time_off.sql")
	if err != nil {
		t.Fatalf("read time-off queries: %v", err)
	}
	sql := string(query)
	if !strings.Contains(sql, "ORDER BY t.created_at DESC, t.id DESC") || !strings.Contains(sql, "LIMIT 500") {
		t.Fatal("time-off history must have a server-side bound and unique ordering")
	}
}
