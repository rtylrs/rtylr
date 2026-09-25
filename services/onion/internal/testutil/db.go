package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql" // registers the mysql driver with database/sql

	"github.com/voxire/onion/internal/platform/database/dbclient"
)

// OpenDB returns a *sql.DB connected to a MySQL instance for integration tests.
// Database integration tests are opt-in: when DB_HOST is not set (CI and local
// dev without a database) the test is skipped. Set DB_HOST to run them against
// a real MySQL. When DB_HOST is set, a connection failure is fatal.
func OpenDB(tb testing.TB) *sql.DB {
	tb.Helper()

	host := os.Getenv("DB_HOST")
	if host == "" {
		tb.Skip("DB_HOST not set; skipping database integration test")
	}

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
		tb.Fatalf("sql.Open: %v", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		tb.Fatalf("db.PingContext: %v", err)
	}

	// Point the package-level pool at this handle too. Platform code that reads
	// dbclient.Pool() rather than taking a *sql.DB (tenant.ResolveBusinessID
	// most notably) otherwise sees nil and fails closed, which surfaces as a
	// 403 and is indistinguishable from a real permission error.
	tb.Cleanup(dbclient.SetPoolForTesting(db))

	tb.Cleanup(func() { _ = db.Close() })
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

// BusinessID is the fixture business scoping used across all integration tests.
// The fixture intentionally uses the canonical owner/business identity until
// the repository suite has a separate multi-business fixture for every module.
const BusinessID int64 = OwnerID
