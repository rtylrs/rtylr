package authorization

import (
	"context"
	"testing"
)

func TestValidateAssignmentScopeRejectsUnsupportedTypes(t *testing.T) {
	t.Cleanup(SetScopeLookup(func(context.Context, int64, string, int64) (bool, error) {
		t.Fatal("unsupported scope reached the reader")
		return false, nil
	}))
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
			err := ValidateAssignmentScope(context.Background(), 9, tc.scopeType, tc.scopeID)
			if err == nil {
				t.Fatalf("expected %s:%s to be rejected", tc.scopeType, tc.scopeID)
			}
		})
	}
}
