//go:build ignore

// Run with: go run scripts/reset_hr_setup.go
// Resets HR setup state so the wizard can be retested from scratch.
// Reads DB credentials from environment (source .env first).
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

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, password, host, port, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	deleted, reset, err := resetHRSetup(db)
	if err != nil {
		log.Fatalf("reset HR setup: %v", err)
	}

	fmt.Printf("  ✓ Deleted %d rows from hr_setup_steps\n", deleted)
	fmt.Printf("  ✓ Reset setup_completed_at on %d hr_configs rows\n", reset)
	fmt.Println("\n✅ HR setup state cleared. Reload the app to see the wizard.")
}

// resetHRSetup performs both writes in one transaction. In particular, do not
// let the DELETE autocommit before the hr_configs update: a failed second
// statement must not leave the setup wizard permanently half-reset.
func resetHRSetup(db *sql.DB) (deleted, reset int64, err error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, 0, fmt.Errorf("begin transaction: %w", err)
	}
	// Rollback is a no-op after Commit and is the safety net for every error
	// path below (including RowsAffected failures).
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`DELETE FROM hr_setup_steps`)
	if err != nil {
		return 0, 0, fmt.Errorf("delete hr_setup_steps: %w", err)
	}
	deleted, err = res.RowsAffected()
	if err != nil {
		return 0, 0, fmt.Errorf("read deleted hr_setup_steps count: %w", err)
	}

	res, err = tx.Exec(`UPDATE hr_configs SET setup_completed_at = NULL`)
	if err != nil {
		return 0, 0, fmt.Errorf("update hr_configs: %w", err)
	}
	reset, err = res.RowsAffected()
	if err != nil {
		return 0, 0, fmt.Errorf("read reset hr_configs count: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("commit transaction: %w", err)
	}
	return deleted, reset, nil
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
