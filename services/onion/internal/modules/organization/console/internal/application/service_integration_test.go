package application_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/console/internal/application"
	"github.com/voxire/onion/internal/modules/organization/console/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// Database-backed regression tests for the console grant flows. Ids live far
// outside the fixture range (1001-3702) so seeded rows never collide.
const (
	addonGrantTestOwnerID     int64 = 9_999_930
	addonGrantTestBusinessID  int64 = 9_999_931
	accessGrantTestOwnerID    int64 = 9_999_932
	accessGrantTestBusinessID int64 = 9_999_933
)

func seedOwnerBusiness(t *testing.T, db *sql.DB, businessID, ownerID int64) {
	t.Helper()
	ctx := context.Background()

	if _, err := db.ExecContext(ctx,
		`INSERT INTO businesses (id, name, owner_id) VALUES (?, ?, ?)`,
		businessID, fmt.Sprintf("Triage Test Business %d", businessID), ownerID,
	); err != nil {
		t.Fatalf("seed business: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM businesses WHERE id = ?`, businessID)
	})
}

func resetAddonGrantState(t *testing.T, db *sql.DB, businessID int64, addonID string) {
	t.Helper()
	ctx := context.Background()

	_, _ = db.ExecContext(ctx, `DELETE FROM manual_addon_grants WHERE business_id = ? AND addon_id = ?`, businessID, addonID)
	_, _ = db.ExecContext(ctx, `UPDATE business_addons SET is_enabled = FALSE WHERE business_id = ? AND addon_id = ?`, businessID, addonID)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM manual_addon_grants WHERE business_id = ? AND addon_id = ?`, businessID, addonID)
		_, _ = db.ExecContext(ctx, `DELETE FROM business_addons WHERE business_id = ? AND addon_id = ?`, businessID, addonID)
	})
}

func assertAddonEnabled(t *testing.T, repo *infrastructure.Repository, businessID int64, addonID string, want bool) {
	t.Helper()
	got, err := repo.IsBusinessAddonEnabled(context.Background(), businessID, addonID)
	if err != nil {
		t.Fatalf("IsBusinessAddonEnabled: %v", err)
	}
	if got != want {
		t.Errorf("addon %q enabled = %v, want %v", addonID, got, want)
	}
}

func createAddonGrant(t *testing.T, svc *application.Service, businessID, ownerID int64, addonID, note string) *application.AddonGrantDTO {
	t.Helper()
	grant, err := svc.CreateAddonGrant(context.Background(), ownerID, businessID, addonID, ownerID, note, nil)
	if err != nil {
		t.Fatalf("CreateAddonGrant(%s): %v", note, err)
	}
	return grant
}

func TestRevokeAddonGrant_LastGrantDisablesGrantDrivenAddon(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	svc := application.NewService(ctx, infrastructure.NewRepository(db))
	repo := infrastructure.NewRepository(db)

	seedOwnerBusiness(t, db, addonGrantTestBusinessID, addonGrantTestOwnerID)
	resetAddonGrantState(t, db, addonGrantTestBusinessID, "loyalty")

	// The first grant turns the paid add-on on; the second is created while it
	// is already enabled, so only the first owns the enablement.
	first := createAddonGrant(t, svc, addonGrantTestBusinessID, addonGrantTestOwnerID, "loyalty", "enabler")
	second := createAddonGrant(t, svc, addonGrantTestBusinessID, addonGrantTestOwnerID, "loyalty", "follower")
	if !first.EnabledByGrant || second.EnabledByGrant {
		t.Fatalf("unexpected enablement ownership: first=%v second=%v", first.EnabledByGrant, second.EnabledByGrant)
	}
	assertAddonEnabled(t, repo, addonGrantTestBusinessID, "loyalty", true)

	// Revoking the enabler must not disable while another active grant remains.
	if err := svc.RevokeAddonGrant(ctx, first.ID); err != nil {
		t.Fatalf("RevokeAddonGrant first: %v", err)
	}
	assertAddonEnabled(t, repo, addonGrantTestBusinessID, "loyalty", true)

	// Revoking the last active grant must disable even though that grant never
	// flipped the add-on on itself.
	if err := svc.RevokeAddonGrant(ctx, second.ID); err != nil {
		t.Fatalf("RevokeAddonGrant second: %v", err)
	}
	assertAddonEnabled(t, repo, addonGrantTestBusinessID, "loyalty", false)
}

func TestDeleteAddonGrant_LastGrantDisablesGrantDrivenAddon(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	svc := application.NewService(ctx, infrastructure.NewRepository(db))
	repo := infrastructure.NewRepository(db)

	seedOwnerBusiness(t, db, addonGrantTestBusinessID, addonGrantTestOwnerID)
	resetAddonGrantState(t, db, addonGrantTestBusinessID, "loyalty")

	first := createAddonGrant(t, svc, addonGrantTestBusinessID, addonGrantTestOwnerID, "loyalty", "enabler")
	second := createAddonGrant(t, svc, addonGrantTestBusinessID, addonGrantTestOwnerID, "loyalty", "follower")

	// Hard-deleting the enabler loses its row, so the marker must move to the
	// surviving grant or the follow-up delete can no longer turn the add-on off.
	if err := svc.DeleteAddonGrant(ctx, first.ID); err != nil {
		t.Fatalf("DeleteAddonGrant first: %v", err)
	}
	assertAddonEnabled(t, repo, addonGrantTestBusinessID, "loyalty", true)

	grants, err := svc.ListAddonGrants(ctx, 100, 0)
	if err != nil {
		t.Fatalf("ListAddonGrants: %v", err)
	}
	var remaining *application.AddonGrantDTO
	for i := range grants.Grants {
		if grants.Grants[i].ID == second.ID {
			remaining = &grants.Grants[i]
		}
	}
	if remaining == nil {
		t.Fatalf("grant %d disappeared from the list", second.ID)
	}
	if !remaining.EnabledByGrant {
		t.Error("enablement marker was lost when the enabling grant was deleted")
	}

	if err := svc.DeleteAddonGrant(ctx, second.ID); err != nil {
		t.Fatalf("DeleteAddonGrant second: %v", err)
	}
	assertAddonEnabled(t, repo, addonGrantTestBusinessID, "loyalty", false)
}

func TestRevokeAddonGrant_KeepsIndependentlyEnabledAddon(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	svc := application.NewService(ctx, infrastructure.NewRepository(db))
	repo := infrastructure.NewRepository(db)

	seedOwnerBusiness(t, db, addonGrantTestBusinessID, addonGrantTestOwnerID)
	resetAddonGrantState(t, db, addonGrantTestBusinessID, "menu")

	// The business enabled the free add-on itself before any grant existed.
	if _, err := db.ExecContext(ctx, `
INSERT INTO business_addons (business_id, addon_id, is_enabled, enabled_at)
VALUES (?, 'menu', TRUE, NOW())
ON DUPLICATE KEY UPDATE is_enabled = TRUE, enabled_at = NOW(), disabled_at = NULL`,
		addonGrantTestBusinessID,
	); err != nil {
		t.Fatalf("enable addon: %v", err)
	}

	grant := createAddonGrant(t, svc, addonGrantTestBusinessID, addonGrantTestOwnerID, "menu", "pre-enabled")
	if grant.EnabledByGrant {
		t.Fatal("grant claims it enabled an add-on the business had already enabled")
	}

	if err := svc.RevokeAddonGrant(ctx, grant.ID); err != nil {
		t.Fatalf("RevokeAddonGrant: %v", err)
	}
	assertAddonEnabled(t, repo, addonGrantTestBusinessID, "menu", true)
}

func TestRevokeAndDeleteAddonGrant_UnknownGrantReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	svc := application.NewService(ctx, infrastructure.NewRepository(db))

	const unknownID int64 = 9_999_999_999
	if err := svc.RevokeAddonGrant(ctx, unknownID); err == nil {
		t.Error("RevokeAddonGrant reported success for an unknown grant")
	}
	if err := svc.DeleteAddonGrant(ctx, unknownID); err == nil {
		t.Error("DeleteAddonGrant reported success for an unknown grant")
	}
}

func resetAccessGrantState(t *testing.T, db *sql.DB, businessIDs ...int64) {
	t.Helper()
	ctx := context.Background()

	reset := func() {
		for _, businessID := range businessIDs {
			_, _ = db.ExecContext(ctx, `DELETE FROM manual_access_grants WHERE business_id = ?`, businessID)
			_, _ = db.ExecContext(ctx, `DELETE FROM subscriptions WHERE business_id = ?`, businessID)
		}
	}
	reset()
	t.Cleanup(reset)
}

// blockSubscriptionWritesFor makes every subscription write for businessID
// fail, standing in for the transient database error CreateGrant must not
// swallow.
func blockSubscriptionWritesFor(t *testing.T, db *sql.DB, businessID int64) {
	t.Helper()
	ctx := context.Background()

	for _, timing := range []string{"INSERT", "UPDATE"} {
		name := fmt.Sprintf("triage_block_sub_%s_%d", timing, businessID)
		_, _ = db.ExecContext(ctx, fmt.Sprintf("DROP TRIGGER IF EXISTS %s", name))
		stmt := fmt.Sprintf(`CREATE TRIGGER %s BEFORE %s ON subscriptions FOR EACH ROW
BEGIN
	IF NEW.business_id = %d THEN
		SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'subscription write blocked by test';
	END IF;
END`, name, timing, businessID)
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("create %s trigger: %v", timing, err)
		}
	}
	t.Cleanup(func() {
		for _, timing := range []string{"INSERT", "UPDATE"} {
			name := fmt.Sprintf("triage_block_sub_%s_%d", timing, businessID)
			_, _ = db.ExecContext(context.Background(), fmt.Sprintf("DROP TRIGGER IF EXISTS %s", name))
		}
	})
}

func firstPlanID(t *testing.T, repo *infrastructure.Repository) int64 {
	t.Helper()
	plans, err := repo.ListPlans(context.Background())
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	if len(plans) == 0 {
		t.Skip("no plan fixtures available")
	}
	return plans[0].ID
}

func TestCreateGrant_ActivatesSubscription(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	svc := application.NewService(ctx, infrastructure.NewRepository(db))
	repo := infrastructure.NewRepository(db)

	seedOwnerBusiness(t, db, accessGrantTestBusinessID, accessGrantTestOwnerID)
	resetAccessGrantState(t, db, accessGrantTestBusinessID)
	planID := firstPlanID(t, repo)

	// Partials still names an owner with a single business by user id.
	if _, err := svc.CreateGrant(ctx, accessGrantTestOwnerID, 0, planID, accessGrantTestOwnerID, "mirror test", nil); err != nil {
		t.Fatalf("CreateGrant: %v", err)
	}

	sub, err := repo.GetSubscriptionByBusiness(ctx, accessGrantTestBusinessID)
	if err != nil {
		t.Fatalf("subscription was not mirrored: %v", err)
	}
	if sub.PlanID != planID || sub.Status != "active" {
		t.Errorf("subscription = plan %d status %q, want plan %d status active", sub.PlanID, sub.Status, planID)
	}
}

func TestCreateGrant_SubscriptionWriteFailureRollsBackGrant(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	svc := application.NewService(ctx, infrastructure.NewRepository(db))
	repo := infrastructure.NewRepository(db)

	seedOwnerBusiness(t, db, accessGrantTestBusinessID, accessGrantTestOwnerID)
	resetAccessGrantState(t, db, accessGrantTestBusinessID)
	planID := firstPlanID(t, repo)
	blockSubscriptionWritesFor(t, db, accessGrantTestBusinessID)

	if _, err := svc.CreateGrant(ctx, 0, accessGrantTestBusinessID, planID, accessGrantTestOwnerID, "must fail", nil); err == nil {
		t.Fatal("CreateGrant reported success although the subscription write failed")
	}

	var count int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM manual_access_grants WHERE business_id = ?`, accessGrantTestBusinessID,
	).Scan(&count); err != nil {
		t.Fatalf("count grants: %v", err)
	}
	if count != 0 {
		t.Errorf("grant row persisted after the transaction failed: %d", count)
	}
}

// A plan belongs to one business, so an owner with several must say which.
func TestCreateGrant_OwnerWithSeveralBusinessesMustNameOne(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	svc := application.NewService(ctx, infrastructure.NewRepository(db))
	repo := infrastructure.NewRepository(db)

	secondBusinessID := accessGrantTestBusinessID + 100
	seedOwnerBusiness(t, db, accessGrantTestBusinessID, accessGrantTestOwnerID)
	seedOwnerBusiness(t, db, secondBusinessID, accessGrantTestOwnerID)
	resetAccessGrantState(t, db, accessGrantTestBusinessID, secondBusinessID)
	planID := firstPlanID(t, repo)

	if _, err := svc.CreateGrant(ctx, accessGrantTestOwnerID, 0, planID, accessGrantTestOwnerID, "ambiguous", nil); !errors.Is(err, application.ErrOwnerHasSeveralBusinesses) {
		t.Fatalf("CreateGrant by owner = %v, want ErrOwnerHasSeveralBusinesses", err)
	}
	if _, err := svc.CreateGrant(ctx, accessGrantTestOwnerID, secondBusinessID, planID, accessGrantTestOwnerID, "named", nil); err != nil {
		t.Fatalf("CreateGrant for the named business: %v", err)
	}
	if _, err := repo.GetSubscriptionByBusiness(ctx, secondBusinessID); err != nil {
		t.Fatalf("named business has no subscription: %v", err)
	}
	if _, err := repo.GetSubscriptionByBusiness(ctx, accessGrantTestBusinessID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("the other business got a subscription too: %v", err)
	}

	if _, err := svc.GetSubscriptionForUser(ctx, accessGrantTestOwnerID, 0); !errors.Is(err, application.ErrOwnerHasSeveralBusinesses) {
		t.Fatalf("GetSubscriptionForUser by owner = %v, want ErrOwnerHasSeveralBusinesses", err)
	}
	if sub, err := svc.GetSubscriptionForUser(ctx, accessGrantTestOwnerID, secondBusinessID); err != nil || sub == nil {
		t.Fatalf("GetSubscriptionForUser for the named business = %v, %v", sub, err)
	}
	if _, err := svc.GetSubscriptionForUser(ctx, accessGrantTestOwnerID, accessGrantTestBusinessID); !errors.Is(err, application.ErrRecordNotFound) {
		t.Fatalf("GetSubscriptionForUser for the business without a plan = %v, want ErrRecordNotFound", err)
	}
}
