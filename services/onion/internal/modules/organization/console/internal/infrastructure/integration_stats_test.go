package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/console/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetStats_ReturnsNonNegativeCounts(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	stats, err := repo.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.TotalUsers < 0 {
		t.Errorf("expected non-negative TotalUsers, got %d", stats.TotalUsers)
	}
	if stats.TotalBusinesses < 0 {
		t.Errorf("expected non-negative TotalBusinesses, got %d", stats.TotalBusinesses)
	}
	if stats.TotalUsers < 2 {
		t.Errorf("expected at least 2 users in stats, got %d", stats.TotalUsers)
	}
	if stats.TotalBusinesses < 1 {
		t.Errorf("expected at least 1 business in stats, got %d", stats.TotalBusinesses)
	}
}

func TestGetStats_TracksGlobalCounterTransitions(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats before writes: %v", err)
	}

	suffix := time.Now().UnixNano()
	var userID, businessID int64
	if result, err := db.ExecContext(ctx,
		"INSERT INTO users (email, password_hash, full_name) VALUES (?, ?, ?)",
		fmt.Sprintf("console-stats-%d@example.com", suffix), "hash", "Console Stats Counter",
	); err != nil {
		t.Fatalf("insert user: %v", err)
	} else if userID, err = result.LastInsertId(); err != nil {
		t.Fatalf("user id: %v", err)
	}
	if result, err := db.ExecContext(ctx,
		"INSERT INTO businesses (name, owner_id) VALUES (?, ?)",
		fmt.Sprintf("Console Stats Business %d", suffix), userID,
	); err != nil {
		t.Fatalf("insert business: %v", err)
	} else if businessID, err = result.LastInsertId(); err != nil {
		t.Fatalf("business id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM manual_addon_grants WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM manual_access_grants WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM subscriptions WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM businesses WHERE id = ?", businessID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id = ?", userID)
	})
	if _, err := db.ExecContext(ctx, "INSERT INTO subscriptions (business_id, plan_id, status) VALUES (?, ?, 'trial')", businessID, 0); err != nil {
		t.Fatalf("insert subscription: %v", err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO manual_access_grants (business_id, plan_id, granted_by) VALUES (?, ?, ?)", businessID, 0, userID); err != nil {
		t.Fatalf("insert access grant: %v", err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO manual_addon_grants (business_id, addon_id, granted_by) VALUES (?, ?, ?)", businessID, "counter-test", userID); err != nil {
		t.Fatalf("insert addon grant: %v", err)
	}

	afterInsert, err := repo.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats after inserts: %v", err)
	}
	if afterInsert.TotalUsers != before.TotalUsers+1 || afterInsert.TotalBusinesses != before.TotalBusinesses+1 ||
		afterInsert.TrialSubs != before.TrialSubs+1 || afterInsert.ManualGrants != before.ManualGrants+1 ||
		afterInsert.ManualAddonGrants != before.ManualAddonGrants+1 {
		t.Fatalf("summary after inserts = %+v, before = %+v", afterInsert, before)
	}

	if _, err := db.ExecContext(ctx, "UPDATE users SET is_banned = TRUE, is_console_admin = TRUE WHERE id = ?", userID); err != nil {
		t.Fatalf("update user flags: %v", err)
	}
	if _, err := db.ExecContext(ctx, "UPDATE subscriptions SET status = 'active' WHERE business_id = ?", businessID); err != nil {
		t.Fatalf("activate subscription: %v", err)
	}
	if _, err := db.ExecContext(ctx, "UPDATE manual_access_grants SET revoked_at = NOW() WHERE business_id = ?", businessID); err != nil {
		t.Fatalf("revoke access grant: %v", err)
	}
	if _, err := db.ExecContext(ctx, "UPDATE manual_addon_grants SET revoked_at = NOW() WHERE business_id = ?", businessID); err != nil {
		t.Fatalf("revoke addon grant: %v", err)
	}

	afterTransitions, err := repo.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats after transitions: %v", err)
	}
	if afterTransitions.TotalUsers != before.TotalUsers+1 || afterTransitions.TotalBusinesses != before.TotalBusinesses+1 ||
		afterTransitions.BannedUsers != before.BannedUsers+1 || afterTransitions.ConsoleAdmins != before.ConsoleAdmins+1 ||
		afterTransitions.ActiveSubs != before.ActiveSubs+1 || afterTransitions.TrialSubs != before.TrialSubs ||
		afterTransitions.ManualGrants != before.ManualGrants || afterTransitions.ManualAddonGrants != before.ManualAddonGrants {
		t.Fatalf("summary after transitions = %+v, before = %+v", afterTransitions, before)
	}
}

func TestListPlans_ReturnsPlans(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	plans, err := repo.ListPlans(ctx)
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	// Plans may be empty in test DB but call must succeed.
	_ = plans
}

func TestGetGrant_ReturnsCreatedGrant(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	plans, err := repo.ListPlans(ctx)
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	if len(plans) == 0 {
		t.Skip("no plan fixtures available")
	}

	grantID, err := repo.CreateGrant(ctx, testutil.BusinessID, plans[0].ID, testutil.OwnerID, "lookup test", nil)
	if err != nil {
		t.Fatalf("CreateGrant: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM manual_access_grants WHERE id = ?", grantID)
	})

	grant, err := repo.GetGrant(ctx, grantID)
	if err != nil {
		t.Fatalf("GetGrant: %v", err)
	}
	if grant.ID != grantID {
		t.Fatalf("grant ID = %d, want %d", grant.ID, grantID)
	}
	if grant.PlanID != plans[0].ID {
		t.Fatalf("plan ID = %d, want %d", grant.PlanID, plans[0].ID)
	}
}

func TestResourceCRUD_HidesSensitiveUserColumns(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	columns, _, err := repo.DescribeResourceTable(ctx, "users")
	if err != nil {
		t.Fatalf("DescribeResourceTable users: %v", err)
	}
	assertNoResourceColumn(t, columns, "password_hash")

	rows, err := repo.ListResourceRows(ctx, "users", 10, 0)
	if err != nil {
		t.Fatalf("ListResourceRows users: %v", err)
	}
	assertNoResourceColumn(t, rows.Columns, "password_hash")
	for _, row := range rows.Rows {
		if _, ok := row["password_hash"]; ok {
			t.Fatal("generic resource rows exposed password_hash")
		}
	}

	row, rowColumns, _, err := repo.GetResourceRow(ctx, "users", fmt.Sprint(testutil.OwnerID))
	if err != nil {
		t.Fatalf("GetResourceRow users: %v", err)
	}
	assertNoResourceColumn(t, rowColumns, "password_hash")
	if _, ok := row["password_hash"]; ok {
		t.Fatal("generic resource row exposed password_hash")
	}
}

func assertNoResourceColumn(t *testing.T, columns []infrastructure.ResourceColumnRow, name string) {
	t.Helper()
	for _, column := range columns {
		if column.Name == name {
			t.Fatalf("generic resource metadata exposed %s", name)
		}
	}
}

func TestGetUserByEmail_OwnerEmailWorks(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Resolve the owner's email from the DB.
	var email string
	if err := db.QueryRowContext(ctx, `SELECT email FROM users WHERE id = ?`, testutil.OwnerID).Scan(&email); err != nil {
		t.Fatalf("resolve email: %v", err)
	}

	user, err := repo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if user.ID != testutil.OwnerID {
		t.Errorf("expected ID=%d, got %d", testutil.OwnerID, user.ID)
	}
}
