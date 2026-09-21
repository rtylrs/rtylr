package testutil

import (
	"context"
	"database/sql"
	"fmt"
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

	// The business name must be unique per id: businesses.name has a UNIQUE key,
	// and cleanup can fail on FK constraints, so a fixed name would make every
	// later INSERT IGNORE for a different id silently match nothing and leave
	// ResolveBusinessID denying with a 403.
	if _, err := db.ExecContext(ctx,
		`INSERT IGNORE INTO businesses (id, name, owner_id) VALUES (?, ?, ?)`,
		businessID, fmt.Sprintf("Integration Test Business %d", businessID), ownerID,
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

// EnsureBusiness makes businessID exist as an active business owned by ownerID
// without registering a cleanup that deletes it. Use it for tests that only
// need the shared fixture business to exist (FK targets, tenant resolution);
// GrantBusinessAccess is for tests that own the business exclusively.
//
// The delete in GrantBusinessAccess can fail on FK constraints and is ignored,
// which historically leaked fixture rows and broke later suites; not deleting
// keeps the fixture stable for everything that follows.
func EnsureBusiness(tb testing.TB, db *sql.DB, businessID, ownerID int64) {
	tb.Helper()

	ctx := context.Background()
	if _, err := db.ExecContext(ctx,
		`INSERT IGNORE INTO businesses (id, name, owner_id) VALUES (?, ?, ?)`,
		businessID, fmt.Sprintf("Integration Test Business %d", businessID), ownerID,
	); err != nil {
		tb.Fatalf("ensure businesses row (%d): %v", businessID, err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT IGNORE INTO user_business_access (user_id, business_id, role) VALUES (?, ?, 'owner')`,
		ownerID, businessID,
	); err != nil {
		tb.Fatalf("ensure business access (%d, %d): %v", ownerID, businessID, err)
	}
}
