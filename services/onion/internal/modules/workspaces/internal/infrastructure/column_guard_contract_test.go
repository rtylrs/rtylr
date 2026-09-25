package infrastructure_test

import (
	"strings"
	"testing"
)

// Every v2 row carries workspace_id = 0 and no flow_workspaces row has id 0, so
// a statement whose only authorization guard is the workspace pair
//
//	EXISTS (... flow_workspaces w        WHERE w.id  = <table>.workspace_id ...)
//	OR EXISTS (... flow_workspace_members wm WHERE wm.workspace_id = ...)
//
// can never be satisfied by any caller. It matches zero rows, requireAffected
// turns that into ErrUnauthorized, and the endpoint answers 403 to everyone
// including the business owner. That is what stopped board columns being
// deleted (#1062).
//
// Asserted on the query text rather than against a database, matching
// hour_store_contract_test.go alongside it. What this cannot check is that the
// project clauses admit the right people; that needs fixtures and a database.
func TestColumnDeletePathIsGuardedOnTheProject(t *testing.T) {
	source := readQuery(t, "../../sql/queries/columns.sql")

	// MoveTaskToColumn was repaired first and is the template the other two now
	// follow, so it is asserted here too: if it ever loses the project clauses,
	// the template is gone.
	for _, query := range []string{"DeleteColumn", "ClearTasksSectionID", "MoveTaskToColumn"} {
		t.Run(query, func(t *testing.T) {
			body := namedBlock(t, source, query)

			if !strings.Contains(body, "flow_project_members") {
				t.Errorf("%s has no project-membership clause, so it cannot admit a project member:\n%s", query, body)
			}
			if !strings.Contains(body, "flow_projects") {
				t.Errorf("%s has no project owner or creator clause:\n%s", query, body)
			}

			// The workspace clauses are harmless and are kept so a surviving
			// workspace-era row still resolves. They must not be the only guard.
			guardsWorkspace := strings.Contains(body, "flow_workspaces")
			guardsProject := strings.Contains(body, "flow_project_members") || strings.Contains(body, "flow_projects")
			if guardsWorkspace && !guardsProject {
				t.Errorf("%s is guarded on the workspace alone, which no v2 row can satisfy:\n%s", query, body)
			}
		})
	}
}

func TestMoveTaskToColumnRejectsAnotherProjectsColumn(t *testing.T) {
	body := namedBlock(t, readQuery(t, "../../sql/queries/columns.sql"), "MoveTaskToColumn")

	if !strings.Contains(body, "c.project_id = flow_tasks.project_id") {
		t.Fatalf("move guard does not require the destination project to match:\n%s", body)
	}
	if strings.Contains(body, "c.project_id = flow_tasks.project_id OR c.workspace_id = flow_tasks.workspace_id") {
		t.Fatalf("move guard still accepts a same-workspace column from another project:\n%s", body)
	}
	if !strings.Contains(body, "flow_tasks.workspace_id <> 0") {
		t.Fatalf("legacy workspace fallback is not restricted to non-zero workspaces:\n%s", body)
	}
}
