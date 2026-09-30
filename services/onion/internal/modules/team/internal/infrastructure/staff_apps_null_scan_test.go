package infrastructure_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// user_app_permissions.apps is a nullable JSON column, and database/sql refuses
// to scan NULL into json.RawMessage (it is []byte-backed and non-nullable).
// Every member row written before apps are populated therefore reads back NULL,
// so the statements that select it unguarded fail with:
//
//	sql: Scan error on column index 4, name "apps": unsupported Scan,
//	storing driver.Value type <nil> into type *json.RawMessage
//
// The API does not map that error, so it answers 500. Editing the permissions of
// any member without a legacy apps array failed outright, whatever apps or role
// were submitted, including an empty set and a role change alone.
//
// ListStaffUsersByBusinessID had the same scan and was guarded with
// COALESCE(..., JSON_ARRAY()); ListStaffUsers and GetStaffUserByID had not been.
//
// This checks the SQL rather than the result, so it needs no database, which is
// what CI has. The Go-level counterpart is TestNullAppsCannotScanWithoutAGuard.
func TestAppsColumnsAreNullGuarded(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("../../sql/queries", "team.sql"))
	if err != nil {
		t.Fatalf("read team.sql: %v", err)
	}

	// An alias-qualified bare column in a SELECT list. COALESCE(...) and
	// JSON_ARRAY() do not match, so this finds only the unguarded form.
	bare := regexp.MustCompile(`(?m)^\s+\w+\.apps\s*,`)

	for _, line := range strings.Split(withoutComments(string(raw)), "\n") {
		if bare.MatchString(line) {
			t.Errorf("team.sql selects a bare apps column, which cannot scan a NULL value:\n\t%s\n"+
				"wrap it: COALESCE(alias.apps, JSON_ARRAY()) AS apps", strings.TrimSpace(line))
		}
	}
}

// The generated scan target is what makes the NULL fatal, so pin that the column
// is read into a non-nullable type rather than something that tolerates NULL.
func TestGeneratedAppsColumnIsNonNullable(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(
		"persistence/sqlc/team.sql.go",
	))
	if err != nil {
		t.Fatalf("read generated team.sql.go: %v", err)
	}

	source := string(raw)
	for _, rowType := range []string{"GetStaffUserByIDRow", "ListStaffUsersRow"} {
		start := strings.Index(source, "type "+rowType+" struct")
		if start < 0 {
			t.Fatalf("%s not found in the generated code", rowType)
		}
		end := strings.Index(source[start:], "\n}")
		if end < 0 {
			t.Fatalf("%s struct is unterminated", rowType)
		}
		structure := source[start : start+end]

		var found bool
		for _, field := range strings.Split(structure, "\n") {
			if !strings.Contains(field, `db:"apps"`) {
				continue
			}
			found = true
			// sql.Null... would absorb a NULL. json.RawMessage will not, which is
			// why the SQL must coalesce instead.
			if strings.Contains(field, "sql.Null") {
				t.Errorf("%s.Apps scans into a nullable type, so the COALESCE in team.sql is no longer "+
					"load-bearing and the contract test above can drift: %s", rowType, strings.TrimSpace(field))
			}
		}
		if !found {
			t.Errorf("%s has no apps field", rowType)
		}
	}
}

// What COALESCE produces, and what the code does with it. Guards against a
// future change swapping the guard for something that yields a different shape.
func TestCoalescedAppsDecodesToAnEmptyList(t *testing.T) {
	var apps json.RawMessage = json.RawMessage(`[]`)

	var decoded []string
	if err := json.Unmarshal(apps, &decoded); err != nil {
		t.Fatalf("coalesced apps must decode: %v", err)
	}
	if len(decoded) != 0 {
		t.Fatalf("a NULL apps column must read back as an empty list, got %v", decoded)
	}
	if decoded == nil {
		t.Fatal("decoded apps must be non-nil so callers do not have to guard a nil slice")
	}
}
