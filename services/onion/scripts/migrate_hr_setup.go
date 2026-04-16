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

	statements := []struct {
		name string
		sql  string
	}{
		{
			name: "hr_setup_steps",
			sql: `CREATE TABLE IF NOT EXISTS hr_setup_steps (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  owner_id BIGINT NOT NULL,
  step_key VARCHAR(64) NOT NULL,
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
  setup_completed_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
)`,
		},
	}

	for _, s := range statements {
		if _, err := db.Exec(s.sql); err != nil {
			log.Fatalf("create %s: %v", s.name, err)
		}
		fmt.Printf("  ✓ %s\n", s.name)
	}

	fmt.Println("\n✅ hr_setup schema applied.")
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
