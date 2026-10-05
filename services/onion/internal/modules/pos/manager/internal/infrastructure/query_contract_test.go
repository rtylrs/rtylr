package infrastructure

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// The manager credential must never be selected or compared as plaintext. A
// query that filters on `pin` is the pre-fix shape: an index seek on a
// cleartext secret, replayable from any dump of the table.
func TestManagerQueriesDoNotComparePlaintextPIN(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/pos_manager.sql")
	if err != nil {
		t.Fatalf("read manager queries: %v", err)
	}

	// Only the SQL is inspected. Comments legitimately discuss PINs in prose, and
	// matching that text would make this a comment lint rather than a
	// credential check.
	var statements []string
	for _, line := range strings.Split(string(query), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" && !strings.HasPrefix(trimmed, "--") {
			statements = append(statements, trimmed)
		}
	}
	statement := strings.ToLower(strings.Join(statements, "\n"))

	// \bpin\b matches a bare `pin` column but not `pin_hash` or
	// `pos_manager_pins`, where "pin" is part of a longer word.
	if bare := regexp.MustCompile(`\bpin\b`).FindString(statement); bare != "" {
		t.Errorf("manager queries still reference the bare plaintext pin column:\n%s", statement)
	}
	if !strings.Contains(statement, "pin_hash") {
		t.Error("manager queries do not read pin_hash; the credential is not hashed at rest")
	}
}

// The base schema is what a fresh database is built from, and it is also what
// sqlc compiles the module's queries against. A base table that still declares
// the legacy plaintext `pin` column is therefore not merely untidy: sqlc cannot
// generate at all (it has no `pin_hash` to select), and any deployment created
// from schema.sql alone would ship a cleartext manager-credential column that
// zz_pos_manager_pins_hash.sql only removes where it already ran.
//
// The convention for this is the POS users package: commit 3c411b335d ("fix(pos):
// remove legacy plaintext pin column") edited the base schema and added the zz_
// migration in the same change. The migration alone does not cover a new
// database, so both are needed and both are asserted here.
func TestBaseSchemaDeclaresNoPlaintextManagerPINColumn(t *testing.T) {
	const path = "../../../sql/schema/pos/schema.sql"
	schema, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read base POS schema: %v", err)
	}

	table := posManagerPinsTable(t, string(schema))
	// \bpin\b matches a bare `pin` column but not `pin_hash` or
	// `pos_manager_pins`, where "pin" is part of a longer word.
	if bare := regexp.MustCompile(`\bpin\b`).FindString(table); bare != "" {
		t.Errorf("pos_manager_pins in the base schema still declares the plaintext pin column:\n%s", table)
	}
	if !strings.Contains(table, "pin_hash") {
		t.Error("pos_manager_pins in the base schema does not declare pin_hash; a fresh database has no usable manager credential column")
	}
}

// posManagerPinsTable isolates the CREATE TABLE block for pos_manager_pins so the
// assertions above read the table and not the surrounding schema.
func posManagerPinsTable(t *testing.T, schema string) string {
	t.Helper()
	// Every table in this file is declared with backticked identifiers, so match
	// that form and tolerate the bare form rather than pinning one spelling.
	const quoted = "CREATE TABLE IF NOT EXISTS `pos_manager_pins`"
	const bare = "CREATE TABLE IF NOT EXISTS pos_manager_pins"
	_, after, found := strings.Cut(schema, quoted)
	if !found {
		_, after, found = strings.Cut(schema, bare)
	}
	if !found {
		t.Fatal("the base schema does not define pos_manager_pins at all")
	}
	// The statement ends at its first semicolon: a backticked table ends with the
	// ENGINE clause, not with the ");" a plain block would.
	if end := strings.Index(after, ";"); end >= 0 {
		return after[:end]
	}
	return after
}

// The migration that introduces the hash and removes the plaintext column must
// exist and must be idempotent, because schema runners re-apply it.
func TestManagerPINHashMigrationIsIdempotent(t *testing.T) {
	const path = "../../../sql/schema/pos/zz_pos_manager_pins_hash.sql"
	migration, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manager PIN hash migration: %v", err)
	}
	source := string(migration)

	if !strings.Contains(source, "IF(") || !strings.Contains(source, "INFORMATION_SCHEMA.COLUMNS") {
		t.Error("migration is not guarded by an INFORMATION_SCHEMA check, so re-running it fails")
	}
	for _, want := range []string{"ADD COLUMN pin_hash", "DROP COLUMN pin"} {
		if !strings.Contains(source, want) {
			t.Errorf("migration does not contain %q", want)
		}
	}
	// The operator has to be told, because there is no seamless conversion.
	if !strings.Contains(source, "OPERATOR ACTION REQUIRED") {
		t.Error("migration does not tell the operator that every existing manager PIN is invalidated")
	}
}

// The write path must exist. The bcrypt migration invalidated every manager
// credential, and with no INSERT or UPDATE against pos_manager_pins anywhere in
// the repository the override had no way to be re-provisioned at all: a
// credential nobody can set is a control that does not exist.
func TestManagerQueriesProvideAWritePath(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/pos_manager.sql")
	if err != nil {
		t.Fatalf("read manager queries: %v", err)
	}
	source := string(query)

	for _, want := range []string{
		"-- name: InsertManagerPIN",
		"-- name: UpdateManagerPINHash",
	} {
		if !strings.Contains(source, want) {
			t.Errorf("manager queries have no %q statement: the module can read a credential but never write one", want)
		}
	}

	// The write must land the hash, never a plaintext column.
	insert := statementContaining(t, source, "INSERT INTO pos_manager_pins")
	if bare := regexp.MustCompile(`\bpin\b`).FindString(insert); bare != "" {
		t.Errorf("the insert references the bare plaintext pin column:\n%s", insert)
	}
	if !strings.Contains(insert, "pin_hash") {
		t.Errorf("the insert does not write pin_hash:\n%s", insert)
	}
}

func statementContaining(t *testing.T, source, needle string) string {
	t.Helper()
	for _, statement := range strings.Split(source, ";") {
		if strings.Contains(statement, needle) {
			return statement
		}
	}
	t.Fatalf("no statement containing %q", needle)
	return ""
}
