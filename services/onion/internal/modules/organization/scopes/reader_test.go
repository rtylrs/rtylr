package scopes_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/scopes"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/testutil"
)

func TestScopeReaderUsesCurrentBusinessMembershipAndDefault(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID, ownerID, memberID = int64(9967101), int64(9967201), int64(9967202)
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM users WHERE id IN (?, ?)", ownerID, memberID) })
	testutil.GrantBusinessAccess(t, db, businessID, ownerID, memberID)
	reader := scopes.NewReader(db)
	restore := tenant.SetScopeResolver(reader)
	defer restore()
	for _, userID := range []int64{ownerID, memberID} {
		id, err := tenant.ResolveBusinessID(ctx, businessID, userID)
		if err != nil || id != businessID {
			t.Fatalf("business access for %d: id=%d, err=%v", userID, id, err)
		}
	}
	owner, err := reader.ActiveBusinessOwner(ctx, businessID)
	if err != nil || owner != ownerID {
		t.Fatalf("active business owner=%d, err=%v", owner, err)
	}
	if _, err := db.Exec("UPDATE users SET business_id = ? WHERE id = ?", businessID, memberID); err != nil {
		t.Fatal(err)
	}
	id, err := tenant.ResolveBusinessID(ctx, 0, memberID)
	if err != nil || id != businessID {
		t.Fatalf("default business=%d, err=%v", id, err)
	}
	if _, err := reader.UserBusiness(ctx, businessID, testutil.OwnerID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("outsider access: %v", err)
	}
	if _, err := db.Exec("UPDATE authorization_business_members SET status = 'revoked' WHERE business_id = ? AND user_id = ?", businessID, memberID); err != nil {
		t.Fatal(err)
	}
	if id, err := tenant.ResolveBusinessID(ctx, businessID, memberID); id != 0 || !errors.Is(err, tenant.ErrNoBusinessScope) {
		t.Fatalf("revoked member: id=%d, err=%v", id, err)
	}
	if _, err := db.Exec("UPDATE businesses SET is_active = FALSE WHERE id = ?", businessID); err != nil {
		t.Fatal(err)
	}
	if id, err := tenant.ResolveBusinessID(ctx, businessID, ownerID); id != 0 || !errors.Is(err, tenant.ErrNoBusinessScope) {
		t.Fatalf("inactive business: id=%d, err=%v", id, err)
	}
	if id, err := tenant.ActiveBusinessID(ctx, businessID); id != 0 || !errors.Is(err, tenant.ErrNoBusinessScope) {
		t.Fatalf("inactive business lookup: id=%d, err=%v", id, err)
	}
}

func TestSubscriptionAccessUsesCurrentGrantAndTrial(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID = int64(9968301)
	reader := scopes.NewReader(db)
	if _, _, _, err := reader.SubscriptionAccess(ctx, businessID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing subscription: %v", err)
	}
	trialEnd := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	res, err := db.Exec("INSERT INTO subscriptions (business_id, plan_id, status, trial_ends_at) VALUES (?, 1, 'trial', ?)", businessID, trialEnd)
	if err != nil {
		t.Fatal(err)
	}
	subscriptionID, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM subscriptions WHERE id = ?", subscriptionID) })
	status, granted, expires, err := reader.SubscriptionAccess(ctx, businessID)
	if err != nil || status != "trial" || granted || !expires.Equal(trialEnd) {
		t.Fatalf("trial: status=%s, grant=%v, end=%v, error=%v", status, granted, expires, err)
	}
	res, err = db.Exec("INSERT INTO manual_access_grants (business_id, plan_id, granted_by) VALUES (?, 1, 1)", businessID)
	if err != nil {
		t.Fatal(err)
	}
	grantID, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM manual_access_grants WHERE id = ?", grantID) })
	if _, granted, _, err := reader.SubscriptionAccess(ctx, businessID); err != nil || !granted {
		t.Fatalf("grant: granted=%v, err=%v", granted, err)
	}
	if _, err := db.Exec("UPDATE manual_access_grants SET revoked_at = CURRENT_TIMESTAMP WHERE id = ?", grantID); err != nil {
		t.Fatal(err)
	}
	if _, granted, _, err := reader.SubscriptionAccess(ctx, businessID); err != nil || granted {
		t.Fatalf("revoked grant: granted=%v, err=%v", granted, err)
	}
}
