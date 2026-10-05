package infrastructure_test

import (
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

// ensureFixtureBusiness makes sure the fixture business row exists before a
// test writes rows that carry a business_id foreign key (the authorization_*
// tables reference businesses). Other suites share this database and drop the
// row in their cleanup, so this only re-seeds it and deliberately never deletes
// it: removing it again could break a suite running in parallel.
func ensureFixtureBusiness(t *testing.T, db *sql.DB) {
	t.Helper()
	testutil.EnsureBusiness(t, db, testutil.BusinessID, testutil.OwnerID)
}
