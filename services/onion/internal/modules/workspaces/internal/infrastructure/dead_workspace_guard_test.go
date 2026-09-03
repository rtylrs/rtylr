package infrastructure_test

import (
	"strings"
	"testing"
)

// A v2 row carries workspace_id = 0 and no flow_workspaces row has id 0, so a
// statement whose only authorization guard is
//
//	EXISTS (... flow_workspaces w        WHERE w.id  = <table>.workspace_id ...)
//	OR EXISTS (... flow_workspace_members wm WHERE wm.workspace_id = ...)
//
// matches nothing for any caller. Each call site turns that into a permission
// error or a silent no-op, which is why the symptoms looked unrelated: a column
// that will not delete, a label that will not delete, and a task whose logged
// hours never reach actual_hours (#1065).
//
// These assert on the query text, matching hour_store_contract_test.go beside
// them. Each query below must carry a guard that a v2 row can actually satisfy,
// alongside the workspace pair which is kept so legacy rows still resolve.
func TestGuardsAV2RowCanSatisfy(t *testing.T) {
	cases := []struct {
		query string
		file  string
		// satisfiable is the clause a v2 row can meet: a project join for rows
		// that carry project_id, or the business filter for rows that do not.
		satisfiable string
	}{
		{"UpdateColumn", "../../sql/queries/columns.sql", "flow_project_members"},
		{"DeleteColumn", "../../sql/queries/columns.sql", "flow_project_members"},
		{"ClearTasksSectionID", "../../sql/queries/columns.sql", "flow_project_members"},
		{"MoveTaskToColumn", "../../sql/queries/columns.sql", "flow_project_members"},
		{"SetTaskLabels", "../../sql/queries/labels.sql", "flow_project_members"},
		{"RecalcActualHours", "../../sql/queries/helpers.sql", "flow_project_members"},
		{"UpdateLabel", "../../sql/queries/labels.sql", "flow_labels.business_id"},
		{"DeleteLabel", "../../sql/queries/labels.sql", "flow_labels.business_id"},
	}

	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			body := namedBlock(t, readQuery(t, tc.file), tc.query)

			if !strings.Contains(body, "flow_workspace") {
				return // no workspace guard to be dead in the first place
			}
			if !strings.Contains(body, tc.satisfiable) {
				t.Errorf("%s is guarded on the workspace alone, which no v2 row can satisfy; expected a %s clause:\n%s",
					tc.query, tc.satisfiable, body)
			}
		})
	}
}

// flow_labels rows written before the business column was populated carry
// business_id = 0. Without this guard a caller whose business resolved to 0
// would match every one of them.
func TestLabelWritesRejectABusinessOfZero(t *testing.T) {
	source := readQuery(t, "../../sql/queries/labels.sql")
	for _, query := range []string{"UpdateLabel", "DeleteLabel"} {
		body := namedBlock(t, source, query)
		if !strings.Contains(body, "flow_labels.business_id <> 0") {
			t.Errorf("%s does not exclude business_id = 0, so an unresolved business matches every legacy label:\n%s",
				query, body)
		}
	}
}
