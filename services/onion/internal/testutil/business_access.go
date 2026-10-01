package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

// GrantBusinessAccess creates a business owned by the first user and adds the
// remaining users as active members for tenant-resolution tests.
func GrantBusinessAccess(tb testing.TB, db *sql.DB, businessID int64, userIDs ...int64) {
	tb.Helper()
	if len(userIDs) == 0 {
		tb.Fatal("GrantBusinessAccess: need at least one user to own the business")
	}

	ctx := context.Background()
	ownerID := userIDs[0]
	for _, userID := range userIDs {
		if _, err := db.ExecContext(ctx,
			`INSERT IGNORE INTO users (id, email, password_hash, full_name) VALUES (?, ?, 'test', ?)`,
			userID, fmt.Sprintf("integration-test-%d@test.local", userID), fmt.Sprintf("Integration Test User %d", userID),
		); err != nil {
			tb.Fatalf("seed users row (%d): %v", userID, err)
		}
	}
	if _, err := db.ExecContext(ctx,
		`INSERT IGNORE INTO businesses (id, name, owner_id) VALUES (?, ?, ?)`,
		businessID, fmt.Sprintf("Integration Test Business %d", businessID), ownerID,
	); err != nil {
		tb.Fatalf("seed businesses row (%d): %v", businessID, err)
	}

	for _, userID := range userIDs {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO authorization_business_members (business_id, user_id, member_type, status)
			VALUES (?, ?, 'member', 'active')
			ON DUPLICATE KEY UPDATE member_type = VALUES(member_type), status = 'active'`,
			businessID, userID,
		); err != nil {
			tb.Fatalf("GrantBusinessAccess(%d, %d): %v", userID, businessID, err)
		}
	}

	tb.Cleanup(func() {
		for _, userID := range userIDs {
			_, _ = db.ExecContext(ctx,
				`DELETE FROM authorization_business_members WHERE user_id = ? AND business_id = ?`,
				userID, businessID,
			)
		}
		_, _ = db.ExecContext(ctx, `DELETE FROM businesses WHERE id = ?`, businessID)
	})
}

// EnsureBusiness ensures an active test business exists without registering
// cleanup, because shared fixture businesses are referenced by other tests.
func EnsureBusiness(tb testing.TB, db *sql.DB, businessID, ownerID int64) {
	tb.Helper()
	if _, err := db.ExecContext(context.Background(),
		`INSERT IGNORE INTO businesses (id, name, owner_id) VALUES (?, ?, ?)`,
		businessID, fmt.Sprintf("Integration Test Business %d", businessID), ownerID,
	); err != nil {
		tb.Fatalf("ensure businesses row (%d): %v", businessID, err)
	}
}
