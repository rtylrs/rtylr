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

	// Clear all setup step completions
	res, err := db.Exec(`DELETE FROM hr_setup_steps`)
	if err != nil {
		log.Fatalf("delete hr_setup_steps: %v", err)
	}
	rows, _ := res.RowsAffected()
	fmt.Printf("  ✓ Deleted %d rows from hr_setup_steps\n", rows)

	// Reset setup_completed_at on hr_configs
	res, err = db.Exec(`UPDATE hr_configs SET setup_completed_at = NULL`)
	if err != nil {
		log.Fatalf("update hr_configs: %v", err)
	}
	rows, _ = res.RowsAffected()
	fmt.Printf("  ✓ Reset setup_completed_at on %d hr_configs rows\n", rows)

	fmt.Println("\n✅ HR setup state cleared. Reload the app to see the wizard.")
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
