package dbclient

import "database/sql"

// UseForTesting installs db as the primary pool without connecting, so tests
// can construct modules (which only keep the handle) without a database.
// Never call it from server code.
func UseForTesting(db *sql.DB) {
	client := getClient()
	client.mainDB = db
}
