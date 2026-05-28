package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql" // registers the mysql driver with database/sql
)

// OpenDB returns a *sql.DB connected to the test MySQL instance.
// It calls t.Skip when DB_HOST is not set (local dev without a DB),
// and t.Fatal when the connection fails in CI.
func OpenDB(t *testing.T) *sql.DB {
	t.Helper()

	host := envOr("DB_HOST", "127.0.0.1")

	port := envOr("DB_PORT", "3306")
	user := envOr("DB_USER", "onion")
	pass := envOr("DB_PASSWORD", "onionpw")
	name := envOr("DB_NAME", "onion_test")

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC&charset=utf8mb4&timeout=10s&readTimeout=15s&writeTimeout=15s",
		user, pass, host, port, name,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("db.PingContext: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// OwnerID is the fixture tenant used across all integration tests.
const OwnerID int64 = 1001
