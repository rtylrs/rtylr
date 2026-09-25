package dbclient

import "database/sql"

// SetPoolForTesting points the package-level pool at a caller-supplied handle
// and returns a function restoring the previous one.
//
// Integration tests open their own *sql.DB through internal/testutil, but a lot
// of platform code reads Pool() instead of taking a handle. tenant.ResolveBusinessID
// is the one that bites: it returns ErrNoBusinessScope when Pool() is nil, and
// every caller maps that to a 403. So a handler that resolves the tenant fails
// with "permission denied" in tests for want of a pool, which looks exactly like
// a real authorization failure and is why a batch of tests in the workspaces
// module has been red.
//
// Deliberately not implemented by calling Connect: that resolves a DSN from the
// environment, where DATABASE_URL takes precedence over the DB_* variables the
// test harness sets. A machine with DATABASE_URL pointing at a deployed database
// would have its test suite connect there. This takes the handle the test already
// opened and resolves nothing.
func SetPoolForTesting(db *sql.DB) (restore func()) {
	client := getClient()
	previous := client.mainDB
	client.mainDB = db
	return func() { client.mainDB = previous }
}
