package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

func TestListUserGroupsHasServerBound(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/groups.sql")
	if err != nil {
		t.Fatalf("read group queries: %v", err)
	}
	if got := strings.Count(string(query), "ORDER BY g.name\nLIMIT 500"); got != 1 {
		t.Fatalf("expected ListUserGroups to have one server-side bound, found %d", got)
	}
}

func TestListGroupMembersHasServerBound(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/groups.sql")
	if err != nil {
		t.Fatalf("read group queries: %v", err)
	}
	if got := strings.Count(string(query), "ORDER BY m.added_at DESC\nLIMIT 500"); got != 1 {
		t.Fatalf("expected ListGroupMembers to have one server-side bound, found %d", got)
	}
}

func TestListGroupResourcePermissionsHasServerBound(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/groups.sql")
	if err != nil {
		t.Fatalf("read group queries: %v", err)
	}
	if got := strings.Count(string(query), "WHERE group_id = ?\nLIMIT 500"); got != 1 {
		t.Fatalf("expected ListGroupResourcePermissions to have one server-side bound, found %d", got)
	}
}
