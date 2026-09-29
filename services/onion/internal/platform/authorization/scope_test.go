package authorization

import (
	"context"
	"database/sql"
	"testing"
)

// failingQuerier proves a rejection happens before any lookup runs, so an
// unsupported scope type never reaches the scopes table or the constraint.
type failingQuerier struct {
	t *testing.T
}

func (f failingQuerier) QueryRowContext(context.Context, string, ...any) *sql.Row {
	f.t.Fatal("scope validation queried the database for an unsupported scope")
	return nil
}

func TestValidateAssignmentScopeRejectsUnsupportedTypes(t *testing.T) {
	// chk_auth_assignment_scope_type permits only pos_location and flow_project
	// with a positive numeric id. The role-access design plan still shows
	// `department:4` and `business:*` examples, neither of which the constraint
	// can store, so the validator must reject them before the write.
	for _, tc := range []struct {
		name      string
		scopeType string
		scopeID   string
	}{
		{name: "business scope", scopeType: "business", scopeID: "9"},
		{name: "department scope", scopeType: "department", scopeID: "4"},
		{name: "location scope", scopeType: "location", scopeID: "12"},
		{name: "hr department", scopeType: "hr_department", scopeID: "3"},
		{name: "non numeric id", scopeType: "pos_location", scopeID: "abc"},
		{name: "zero id", scopeType: "pos_location", scopeID: "0"},
		{name: "negative id", scopeType: "flow_project", scopeID: "-1"},
		{name: "empty type", scopeType: "", scopeID: "1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAssignmentScope(context.Background(), failingQuerier{t}, 9, tc.scopeType, tc.scopeID)
			if err == nil {
				t.Fatalf("expected %s:%s to be rejected", tc.scopeType, tc.scopeID)
			}
		})
	}
}
