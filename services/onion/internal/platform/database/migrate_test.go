package database_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func migrationFixture(t *testing.T, registry string, sources map[string]string) (string, func() ([]byte, error)) {
	t.Helper()
	root := t.TempDir()
	write := func(path, content string, mode os.FileMode) {
		t.Helper()
		path = filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), mode); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"migrate.sh", "dsn.sh"} {
		data, err := os.ReadFile(filepath.Join("../../../scripts", name))
		if err != nil {
			t.Fatal(err)
		}
		write("scripts/"+name, string(data), 0o755)
	}
	write("internal/platform/database/schema_registry.tsv", registry, 0o644)
	for path, sql := range sources {
		write(path, sql, 0o644)
	}
	write("bin/mysql", `#!/bin/bash
set -eu
query=""
while [ "$#" -gt 0 ]; do
    if [ "$1" = "-e" ]; then query="$2"; shift; fi
    shift
done
case "$query" in
    *"SELECT COUNT(*) FROM schema_migrations"*)
        identity=$(printf '%s' "$query" | cut -d "'" -f 2)
        if grep -Fxq "$identity" "$FIXTURE_APPLIED"; then echo 1; else echo 0; fi
        ;;
    *"INSERT IGNORE INTO schema_migrations"*)
        printf '%s\n' "$query" | cut -d "'" -f 2 >> "$FIXTURE_APPLIED"
        ;;
    *)
        sql=$(cat)
        printf '%s\n' "$sql" >> "$FIXTURE_SQL"
        if [[ "$sql" == *FAIL_FIXTURE* ]]; then exit 9; fi
        ;;
esac
`, 0o755)
	write("applied", "", 0o644)
	write("sql", "", 0o644)
	run := func() ([]byte, error) {
		cmd := exec.Command("bash", "scripts/migrate.sh")
		cmd.Dir = root
		cmd.Env = append(os.Environ(),
			"PATH="+filepath.Join(root, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
			"DATABASE_URL=", "DB_HOST=127.0.0.1", "DB_PORT=3306", "DB_USER=fixture", "DB_PASSWORD=fixture", "DB_NAME=fixture",
			"FIXTURE_APPLIED="+filepath.Join(root, "applied"), "FIXTURE_SQL="+filepath.Join(root, "sql"))
		return cmd.CombinedOutput()
	}
	return root, run
}

func TestMigrationRegistryPreservesIdentityAcrossRelocation(t *testing.T) {
	const schema = "internal/modules/sample/sql/schema/schema.sql"
	const migration = "internal/modules/sample/sql/migration_20261004_example.sql"
	root, run := migrationFixture(t, "legacy/schema.sql\t"+schema+"\n"+migration+"\t"+migration+"\n", map[string]string{
		schema:    "CREATE TABLE example (id BIGINT PRIMARY KEY);",
		migration: "ALTER TABLE example ADD COLUMN label TEXT;",
	})
	for i := 0; i < 2; i++ {
		if output, err := run(); err != nil {
			t.Fatalf("migration run %d: %v\n%s", i, err, output)
		}
	}
	applied, err := os.ReadFile(filepath.Join(root, "applied"))
	if err != nil {
		t.Fatal(err)
	}
	if string(applied) != "legacy/schema.sql\n"+migration+"\n" {
		t.Fatalf("migration identities changed or were applied twice: %s", applied)
	}
	sql, err := os.ReadFile(filepath.Join(root, "sql"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(sql), "CREATE TABLE example") != 1 || strings.Count(string(sql), "ALTER TABLE example") != 1 || strings.Index(string(sql), "CREATE TABLE example") > strings.Index(string(sql), "ALTER TABLE example") {
		t.Fatalf("schema and migration must execute once, in order: %s", sql)
	}
}

func TestMigrationRunnerDiscoversNestedModuleMigrations(t *testing.T) {
	const migration = "internal/modules/crm/internal/outbox/sql/migration_20261004_outbox.sql"
	root, run := migrationFixture(t, "", map[string]string{
		migration: "CREATE TABLE crm_outbox (id BIGINT PRIMARY KEY);",
	})
	if output, err := run(); err != nil {
		t.Fatalf("migration run: %v\n%s", err, output)
	}
	applied, err := os.ReadFile(filepath.Join(root, "applied"))
	if err != nil {
		t.Fatal(err)
	}
	if string(applied) != migration+"\n" {
		t.Fatalf("nested module migration was not recorded: %s", applied)
	}
	sql, err := os.ReadFile(filepath.Join(root, "sql"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(sql), "CREATE TABLE crm_outbox") {
		t.Fatalf("nested module migration was not applied: %s", sql)
	}
}

func TestMigrationFailureDoesNotRecordOrContinue(t *testing.T) {
	const failed = "internal/modules/sample/sql/schema/failed.sql"
	const later = "internal/modules/another/sql/schema/later.sql"
	root, run := migrationFixture(t, "failed.sql\t"+failed+"\nlater.sql\t"+later+"\n", map[string]string{
		failed: "FAIL_FIXTURE;",
		later:  "CREATE TABLE unreachable (id BIGINT);",
	})
	if output, err := run(); err == nil {
		t.Fatalf("failed SQL was reported as successful: %s", output)
	}
	applied, err := os.ReadFile(filepath.Join(root, "applied"))
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 0 {
		t.Fatalf("failed migration was recorded: %s", applied)
	}
	sql, err := os.ReadFile(filepath.Join(root, "sql"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(sql), "unreachable") {
		t.Fatalf("migration runner continued after SQL failure: %s", sql)
	}
}
