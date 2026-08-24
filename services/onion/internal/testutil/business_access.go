package testutil

import (
	"context"
	"database/sql"
	"testing"
)

// GrantBusinessAccess makes businessID a real, active business owned by the
// first user given, and grants every user access to it. Rows are removed
// afterwards.
//
// Needed because v2 handlers resolve the tenant through
// tenant.ResolveBusinessID, and that check starts FROM the businesses table:
//
//	SELECT 1 FROM businesses b
//	WHERE b.id = ? AND COALESCE(b.is_active, TRUE) = TRUE
//	  AND (b.owner_id = ? OR EXISTS (SELECT 1 FROM user_business_access ...))
//
// So seeding user_business_access alone is not enough - with no businesses row
// the query matches nothing, returns sql.ErrNoRows, and ResolveBusinessID maps
// that to ErrNoBusinessScope. Callers turn that into a 403.
//
// That is the trap worth naming: the fixture business id every integration test
// scopes by has never existed in the businesses table, so any handler resolving
// the tenant is denied for a reason that has nothing to do with the permission
// it is testing, and the error is indistinguishable from a real denial.
func GrantBusinessAccess(tb testing.TB, db *sql.DB, businessID int64, userIDs ...int64) {
	tb.Helper()

	if len(userIDs) == 0 {
		tb.Fatal("GrantBusinessAccess: need at least one user to own the business")
	}

	ctx := context.Background()
	ownerID := userIDs[0]

	if _, err := db.ExecContext(ctx,
		`INSERT IGNORE INTO businesses (id, name, owner_id) VALUES (?, ?, ?)`,
		businessID, "Integration Test Business", ownerID,
	); err != nil {
		tb.Fatalf("seed businesses row (%d): %v", businessID, err)
	}

	for _, userID := range userIDs {
		if _, err := db.ExecContext(ctx,
			`INSERT IGNORE INTO user_business_access (user_id, business_id, role) VALUES (?, ?, 'owner')`,
			userID, businessID,
		); err != nil {
			tb.Fatalf("GrantBusinessAccess(%d, %d): %v", userID, businessID, err)
		}
	}

	tb.Cleanup(func() {
		for _, userID := range userIDs {
			_, _ = db.ExecContext(ctx,
				`DELETE FROM user_business_access WHERE user_id = ? AND business_id = ?`,
				userID, businessID,
			)
		}
		_, _ = db.ExecContext(ctx, `DELETE FROM businesses WHERE id = ?`, businessID)
	})
}
