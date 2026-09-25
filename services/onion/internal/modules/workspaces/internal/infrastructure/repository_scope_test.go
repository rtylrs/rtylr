package infrastructure

import (
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
)

func TestTaskListConditionsScopeByBusiness(t *testing.T) {
	query, args, err := taskListConditions("t.workspace_id", 42, domain.ListTasksFilter{
		BusinessID: 7,
	}).ToSql()
	if err != nil {
		t.Fatalf("build conditions: %v", err)
	}

	if !strings.Contains(query, "t.workspace_id = ?") || !strings.Contains(query, "t.business_id = ?") {
		t.Fatalf("conditions are missing workspace/business scope: %s", query)
	}
	if len(args) != 2 || args[0] != int64(42) || args[1] != int64(7) {
		t.Fatalf("unexpected condition args: %#v", args)
	}
}
