package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/team/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// Group memberships written through the Groups API must carry the caller's
// business_id: the authorization snapshot readers resolve group licenses and
// role assignments with `user_group_members.business_id = <member's business>`,
// so a row left at the column default 0 silently drops every group grant.
func TestTeam_AddGroupMembers_WritesBusinessIDResolvableBySnapshot(t *testing.T) {
	db := testutil.OpenDB(t)
	ensureFixtureBusiness(t, db)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	businessID := testutil.BusinessID
	ownerID := testutil.OwnerID

	groupName := fmt.Sprintf("snapshot-group-%d", time.Now().UnixNano())
	groupID, err := repo.CreateGroup(ctx, ownerID, domain.CreateGroupInput{BusinessID: businessID, Name: groupName})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}

	email := fmt.Sprintf("snapshot-member-%d@test.local", time.Now().UnixNano())
	res, err := db.ExecContext(ctx,
		`INSERT INTO users (email, username, password_hash, business_id) VALUES (?, ?, '', ?)`,
		email, fmt.Sprintf("snapshot-member-%d", time.Now().UnixNano()), businessID)
	if err != nil {
		t.Fatalf("insert member: %v", err)
	}
	userID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_license_assignments WHERE group_id = ?", groupID)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_role_assignments WHERE group_id = ?", groupID)
		_, _ = db.ExecContext(ctx, "DELETE FROM user_group_members WHERE group_id = ?", groupID)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_business_members WHERE business_id = ? AND user_id = ?", businessID, userID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
		_ = repo.DeleteGroup(ctx, ownerID, businessID, groupID)
	})

	if _, err := db.ExecContext(ctx,
		`INSERT INTO authorization_license_assignments (business_id, group_id, app_id, source, status)
		 VALUES (?, ?, 'pos', 'group', 'active')`, businessID, groupID); err != nil {
		t.Fatalf("insert group license: %v", err)
	}

	var roleID int64
	if err := db.QueryRowContext(ctx,
		`SELECT id FROM authorization_roles WHERE business_id = 0 AND app_id = 'pos' ORDER BY id LIMIT 1`).Scan(&roleID); err != nil {
		t.Fatalf("find seeded pos role: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO authorization_role_assignments (business_id, group_id, app_id, role_id, source, status)
		 VALUES (?, ?, 'pos', ?, 'group', 'active')`, businessID, groupID, roleID); err != nil {
		t.Fatalf("insert group role assignment: %v", err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO authorization_business_members (business_id, user_id, status) VALUES (?, ?, 'active')`,
		businessID, userID); err != nil {
		t.Fatalf("insert active membership: %v", err)
	}

	if err := repo.AddGroupMembers(ctx, ownerID, groupID, businessID, ownerID, []int64{userID}); err != nil {
		t.Fatalf("AddGroupMembers: %v", err)
	}

	var memberBusinessID int64
	if err := db.QueryRowContext(ctx,
		`SELECT business_id FROM user_group_members WHERE group_id = ? AND user_id = ?`,
		groupID, userID).Scan(&memberBusinessID); err != nil {
		t.Fatalf("read membership: %v", err)
	}
	if memberBusinessID != businessID {
		t.Fatalf("membership business_id = %d, want %d", memberBusinessID, businessID)
	}

	snapshot, found, err := repo.GetAuthorizationSnapshot(ctx, businessID, userID, []string{"pos"})
	if err != nil {
		t.Fatalf("GetAuthorizationSnapshot: %v", err)
	}
	if !found {
		t.Fatal("authorization snapshot not found for member")
	}

	licenseFound := false
	for _, license := range snapshot.Licenses {
		if license.AppID == "pos" && license.GroupID != nil && *license.GroupID == groupID {
			licenseFound = true
		}
	}
	if !licenseFound {
		t.Fatalf("group license not resolved for member: licenses=%+v", snapshot.Licenses)
	}

	assignmentFound := false
	for _, assignment := range snapshot.Assignments {
		if assignment.GroupID != nil && *assignment.GroupID == groupID {
			assignmentFound = true
		}
	}
	if !assignmentFound {
		t.Fatalf("group role assignment not resolved for member: assignments=%+v", snapshot.Assignments)
	}
}
