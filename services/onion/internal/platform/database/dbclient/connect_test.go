package dbclient

import (
	"context"
	"os"
	"sync"
	"testing"
)

// resetClient drops the process-wide client so a test can exercise Connect from
// a clean slate. Production never does this; the dbclient tests are sequential.
func resetClient() {
	clientInstance = &MySQLClient{}
	clientOnce = sync.Once{}
}

// Regression: a failed logging-pool connect used to return with client.mainDB
// still set, so the API restart loop's next Connect hit the mainDB != nil early
// return and booted "successfully" with LoggingPool()/ReplicaPool() silently
// falling back to the primary. A failed Connect must leave no pool behind.
func TestConnect_FailedLoggingPoolLeavesNoHalfInitializedClient(t *testing.T) {
	if os.Getenv("DB_HOST") == "" {
		t.Skip("DB_HOST not set; skipping database integration test")
	}

	// The main DSN is built from DB_*; a developer's DATABASE_URL must not
	// override it here.
	t.Setenv("DATABASE_URL", "")
	t.Setenv("LOGGING_DATABASE_URL", "unreachable:wrong@tcp(127.0.0.1:1)/no_such_db")

	resetClient()
	t.Cleanup(func() {
		Close()
		resetClient()
	})

	ctx := context.Background()
	if err := Connect(ctx); err == nil {
		t.Fatal("Connect with an unreachable logging database returned nil")
	}
	if getClient().mainDB != nil {
		t.Fatal("failed logging connect left client.mainDB set; the next Connect would report success without a logging pool")
	}

	// The retry the restart loop performs must now complete initialization.
	t.Setenv("LOGGING_DATABASE_URL", "")
	if err := Connect(ctx); err != nil {
		t.Fatalf("Connect retry after the logging failure: %v", err)
	}
	if getClient().mainDB == nil {
		t.Fatal("Connect retry did not establish the main pool")
	}
}
