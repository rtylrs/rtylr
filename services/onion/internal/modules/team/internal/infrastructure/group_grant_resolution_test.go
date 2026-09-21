package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

// A group app grant must resolve to a 'group' license row: resolveApp rejects an
// explicit row that carries a group id, and the matched-group branch suppresses
// the per-user fallback license, so before the fix a group app grant denied the
// app to every member of the group.
func TestTeam_UpsertNormalizedAuthorization_GroupAppGrantResolves(t *testing.T) {
	db := testutil.OpenDB(t)
	ensureFixtureBusiness(t, db)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	businessID := testutil.BusinessID
	ownerID := testutil.OwnerID

	groupID, err := repo.CreateGroup(ctx, ownerID, domain.CreateGroupInput{
		Name: fmt.Sprintf("grant-group-%d", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	userID := insertTenantMember(t, db, businessID, "none")

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_license_assignments WHERE group_id = ?", groupID)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_role_assignments WHERE group_id = ?", groupID)
		_, _ = db.ExecContext(ctx, "DELETE FROM user_group_members WHERE group_id = ?", groupID)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_license_assignments WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_role_assignments WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_business_members WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_audit_events WHERE target_user_id = ?", userID)
		_ = repo.DeleteGroup(ctx, ownerID, groupID)
	})

	groupIDCopy := groupID
	if err := repo.UpsertNormalizedAuthorization(ctx, businessID, ownerID, userID,
		[]string{"pos"}, domain.RoleStaff, "member",
		[]domain.AppRoleAssignmentInput{{AppID: "pos", GroupID: &groupIDCopy}},
		[]int64{groupID},
	); err != nil {
		t.Fatalf("UpsertNormalizedAuthorization: %v", err)
	}

	var source string
	var licenseUserID *int64
	if err := db.QueryRowContext(ctx,
		`SELECT source, user_id FROM authorization_license_assignments
		 WHERE business_id = ? AND group_id = ? AND app_id = 'pos'`,
		businessID, groupID).Scan(&source, &licenseUserID); err != nil {
		t.Fatalf("read group license: %v", err)
	}
	if source != string(authorization.SourceGroup) {
		t.Fatalf("group license source = %q, want %q", source, authorization.SourceGroup)
	}
	if licenseUserID != nil {
		t.Fatalf("group license user_id = %d, want NULL", *licenseUserID)
	}

	snapshot, found, err := repo.GetAuthorizationSnapshot(ctx, businessID, userID, []string{"pos"})
	if err != nil {
		t.Fatalf("GetAuthorizationSnapshot: %v", err)
	}
	if !found {
		t.Fatal("authorization snapshot not found")
	}
	access := authorization.Resolve(snapshot)
	if len(access.Apps) != 1 {
		t.Fatalf("resolved apps = %d, want 1", len(access.Apps))
	}
	if app := access.Apps[0]; !app.Licensed || !app.HasAccess {
		t.Fatalf("pos access = licensed:%t hasAccess:%t, want both true", app.Licensed, app.HasAccess)
	}
}

// appRoleKey must only ask for catalog keys that exist. Insights has no manager
// role (the backfill maps the legacy manager roles to viewer) and pos maps
// menu_manager to location_manager, so lookups used to fail with sql.ErrNoRows
// and rolled the whole invite/update back.
func TestTeam_UpsertNormalizedAuthorization_LegacyManagerRoleKeysExist(t *testing.T) {
	db := testutil.OpenDB(t)
	ensureFixtureBusiness(t, db)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	businessID := testutil.BusinessID
	ownerID := testutil.OwnerID
	userID := insertTenantMember(t, db, businessID, "none")

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_license_assignments WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_role_assignments WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_business_members WHERE user_id = ?", userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_audit_events WHERE target_user_id = ?", userID)
	})

	roleKeyFor := func(appID string) (string, error) {
		var roleKey string
		err := db.QueryRowContext(ctx, `
			SELECT r.role_key
			FROM authorization_role_assignments a
			JOIN authorization_roles r ON r.id = a.role_id
			WHERE a.business_id = ? AND a.user_id = ? AND a.app_id = ? AND a.status = 'active'
			ORDER BY a.id DESC
			LIMIT 1`, businessID, userID, appID).Scan(&roleKey)
		return roleKey, err
	}

	if err := repo.UpsertNormalizedAuthorization(ctx, businessID, ownerID, userID,
		[]string{"insights"}, domain.RoleOperator, "", nil, nil); err != nil {
		t.Fatalf("UpsertNormalizedAuthorization(insights, operator): %v", err)
	}
	roleKey, err := roleKeyFor("insights")
	if err != nil {
		t.Fatalf("read insights role assignment: %v", err)
	}
	if roleKey != "insights.viewer" {
		t.Fatalf("insights role = %q, want insights.viewer", roleKey)
	}

	if err := repo.UpsertNormalizedAuthorization(ctx, businessID, ownerID, userID,
		[]string{"pos"}, domain.RoleMenuManager, "", nil, nil); err != nil {
		t.Fatalf("UpsertNormalizedAuthorization(pos, menu_manager): %v", err)
	}
	roleKey, err = roleKeyFor("pos")
	if err != nil {
		t.Fatalf("read pos role assignment: %v", err)
	}
	if roleKey != "pos.location_manager" {
		t.Fatalf("pos role = %q, want pos.location_manager", roleKey)
	}
}
