package infrastructure_test

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestTeam_UpdateGroup_InvalidProfileRollsBack(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.EnsureBusiness(t, db, testutil.BusinessID, testutil.OwnerID)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	name := fmt.Sprintf("tx-rollback-group-%d", time.Now().UnixNano())
	groupID, err := repo.CreateGroup(ctx, testutil.OwnerID, domain.CreateGroupInput{
		BusinessID:     testutil.BusinessID,
		Name:           name,
		LicensedAppIDs: []string{"pos"},
	})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteGroup(ctx, testutil.OwnerID, testutil.BusinessID, groupID) })

	err = repo.UpdateGroup(ctx, testutil.OwnerID, testutil.BusinessID, groupID, domain.UpdateGroupInput{
		BusinessID:     testutil.BusinessID,
		Name:           name + "-renamed",
		LicensedAppIDs: []string{"crm"},
		PermissionProfileAssignments: []domain.AppRoleAssignmentInput{{
			AppID: "crm", RoleIDs: []int64{math.MaxInt64},
		}},
	})
	if err == nil {
		t.Fatal("UpdateGroup with a nonexistent permission profile: expected error, got nil")
	}

	group, err := repo.GetGroup(ctx, testutil.OwnerID, testutil.BusinessID, groupID)
	if err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	if group.Name != name {
		t.Errorf("group name = %q, want %q after rollback", group.Name, name)
	}
	if len(group.LicensedAppIDs) != 1 || group.LicensedAppIDs[0] != "pos" {
		t.Errorf("licensed apps = %v, want [pos] after rollback", group.LicensedAppIDs)
	}
}

func TestTeam_CreateGroup_InvalidProfileLeavesNoGroup(t *testing.T) {
	db := testutil.OpenDB(t)
	testutil.EnsureBusiness(t, db, testutil.BusinessID, testutil.OwnerID)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	name := fmt.Sprintf("tx-rollback-create-%d", time.Now().UnixNano())
	if _, err := repo.CreateGroup(ctx, testutil.OwnerID, domain.CreateGroupInput{
		BusinessID: testutil.BusinessID,
		Name:       name,
		PermissionProfileAssignments: []domain.AppRoleAssignmentInput{{
			AppID: "crm", RoleIDs: []int64{math.MaxInt64},
		}},
	}); err == nil {
		t.Fatal("CreateGroup with a nonexistent permission profile: expected error, got nil")
	}

	groupID, err := repo.CreateGroup(ctx, testutil.OwnerID, domain.CreateGroupInput{
		BusinessID: testutil.BusinessID,
		Name:       name,
	})
	if err != nil {
		t.Fatalf("retry CreateGroup with the same name: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteGroup(ctx, testutil.OwnerID, testutil.BusinessID, groupID) })
}
