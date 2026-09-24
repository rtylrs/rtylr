//go:build ignore

// Run with: go run scripts/migrate_hr_setup.go
// Reads DB credentials from environment variables (source .env first).
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	host := env("DB_HOST", "localhost")
	port := env("DB_PORT", "3306")
	user := mustEnv("DB_USER")
	password := mustEnv("DB_PASSWORD")
	dbName := mustEnv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		user, password, host, port, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	// Keep this standalone helper in sync with the canonical HR schemas. The
	// hr_setup and hr_core readers both select the newer columns, so creating
	// the original two-column tables here makes a fresh database unusable.
	statements := []struct {
		name string
		sql  string
	}{
		{
			name: "hr_setup_steps",
			sql: `CREATE TABLE IF NOT EXISTS hr_setup_steps (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  owner_id BIGINT NOT NULL,
  business_id BIGINT NOT NULL DEFAULT 0,
  step_key VARCHAR(100) NOT NULL,
  step_label VARCHAR(255) NOT NULL DEFAULT '',
  is_completed BOOLEAN NOT NULL DEFAULT FALSE,
  completed_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_hr_setup_steps_owner_step (owner_id, step_key)
)`,
		},
		{
			name: "hr_configs",
			sql: `CREATE TABLE IF NOT EXISTS hr_configs (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  owner_id BIGINT NOT NULL UNIQUE,
  business_id BIGINT NOT NULL DEFAULT 0,
  setup_completed_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
)`,
		},
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement.sql); err != nil {
			log.Fatalf("create %s: %v", statement.name, err)
		}
		fmt.Printf("  ✓ %s\n", statement.name)
	}

	// CREATE TABLE IF NOT EXISTS deliberately does not repair tables created
	// by an older version of this helper. Add the columns conditionally so
	// the same command works for both fresh and existing installations.
	columns := []struct {
		table      string
		name       string
		definition string
	}{
		{table: "hr_setup_steps", name: "business_id", definition: "BIGINT NOT NULL DEFAULT 0"},
		{table: "hr_setup_steps", name: "step_label", definition: "VARCHAR(255) NOT NULL DEFAULT ''"},
		{table: "hr_setup_steps", name: "is_completed", definition: "BOOLEAN NOT NULL DEFAULT FALSE"},
		{table: "hr_configs", name: "business_id", definition: "BIGINT NOT NULL DEFAULT 0"},
	}
	for _, column := range columns {
		if err := ensureColumn(db, column.table, column.name, column.definition); err != nil {
			log.Fatalf("ensure %s.%s: %v", column.table, column.name, err)
		}
	}

	// The legacy tables used owner_id as their tenant scope. Preserve those
	// rows when introducing business_id, and derive the canonical completion
	// flag from the timestamp that older versions wrote.
	backfills := []struct {
		name string
		sql  string
	}{
		{
			name: "hr_setup_steps business scope",
			sql:  `UPDATE hr_setup_steps SET business_id = owner_id WHERE business_id = 0`,
		},
		{
			name: "hr_setup_steps completion flag",
			sql:  `UPDATE hr_setup_steps SET is_completed = TRUE WHERE is_completed = FALSE AND completed_at IS NOT NULL`,
		},
		{
			name: "hr_configs business scope",
			sql:  `UPDATE hr_configs SET business_id = owner_id WHERE business_id = 0`,
		},
	}
	for _, backfill := range backfills {
		if _, err := db.Exec(backfill.sql); err != nil {
			log.Fatalf("backfill %s: %v", backfill.name, err)
		}
	}

	fmt.Println("\n✅ hr_setup schema applied.")
}

func ensureColumn(db *sql.DB, table, column, definition string) error {
	var count int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = DATABASE()
		  AND table_name = ?
		  AND column_name = ?`, table, column).Scan(&count); err != nil {
		return fmt.Errorf("inspect column: %w", err)
	}
	if count != 0 {
		return nil
	}

	// table and column are fixed constants in this migration, not user input.
	statement := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", table, column, definition)
	if _, err := db.Exec(statement); err != nil {
		return fmt.Errorf("add column: %w", err)
	}
	return nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set (did you source .env?)", key)
	}
	return v
}
