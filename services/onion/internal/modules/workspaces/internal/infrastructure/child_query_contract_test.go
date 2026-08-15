package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

func TestWorkspaceTaskChildListsHaveServerBounds(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/workspaces.sql")
	if err != nil {
		t.Fatalf("read workspace queries: %v", err)
	}
	if got := strings.Count(string(query), "LIMIT 500"); got != 3 {
		t.Fatalf("expected three bounded task child lists, found %d", got)
	}
}
