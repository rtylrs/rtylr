package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// RevokeBusinessUserPermissions sent the literal text sqlc.arg('business_id')
// to MySQL, which answered "Error 1305 (42000): FUNCTION sqlc.arg does not
// exist" on every call (#1046). sqlc rewrites sqlc.arg in a WHERE expression
// but left it untouched inside the JOIN ON clause.
//
// The ids below deliberately match nothing. What is under test is whether the
// statement is valid SQL at all, and a parse error fires regardless of whether
// any row qualifies. That also keeps the test independent of fixtures, which
// matters here: the neighbouring TestTerminatedAccessReconciliation currently
// skips because its fixture permission row for user 1002 is absent from
// onion_test, so this whole area was passing vacuously.
func TestRevokeBusinessUserPermissionsIsValidSQL(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	const noSuchBusiness, noSuchUser = int64(999999), int64(999999)

	rows, err := repo.RevokeBusinessUserPermissions(context.Background(), noSuchBusiness, noSuchUser)
	if err != nil {
		t.Fatalf("RevokeBusinessUserPermissions: %v", err)
	}
	if rows != 0 {
		t.Errorf("revoked %d rows for a business and user that do not exist, want 0", rows)
	}
}
