package infrastructure_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// A resource_key longer than the VARCHAR(64) column fails in the middle of the
// permission rewrite (after the row update and the permission DELETE), which is
// exactly the failure the transaction must roll back.
func overLongResourceKey() string {
	return strings.Repeat("k", 65)
}

func TestTeam_UpdateGroup_FailedPermissionRewriteRollsBack(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	name := fmt.Sprintf("tx-rollback-group-%d", time.Now().UnixNano())
	groupID, err := repo.CreateGroup(ctx, testutil.OwnerID, domain.CreateGroupInput{
		Name: name,
		ResourcePermissions: []domain.ResourcePermission{
			{ResourceKey: "crm.customers", PermissionLevel: domain.PermLevelEditor},
		},
	})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteGroup(ctx, testutil.OwnerID, groupID) })

	err = repo.UpdateGroup(ctx, testutil.OwnerID, groupID, domain.UpdateGroupInput{
		Name: name + "-renamed",
		ResourcePermissions: []domain.ResourcePermission{
			{ResourceKey: "crm.orders", PermissionLevel: domain.PermLevelViewer},
			{ResourceKey: overLongResourceKey(), PermissionLevel: domain.PermLevelViewer},
		},
	})
	if err == nil {
		t.Fatal("UpdateGroup with an over-long resource_key: expected error, got nil")
	}

	group, err := repo.GetGroup(ctx, testutil.OwnerID, groupID)
	if err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	if group.Name != name {
		t.Errorf("group name = %q, want the pre-update %q after rollback", group.Name, name)
	}
	if len(group.ResourcePermissions) != 1 ||
		group.ResourcePermissions[0].ResourceKey != "crm.customers" ||
		group.ResourcePermissions[0].PermissionLevel != domain.PermLevelEditor {
		t.Errorf("resource permissions = %+v, want the original crm.customers/editor after rollback", group.ResourcePermissions)
	}
}

func TestTeam_CreateGroup_FailedPermissionRewriteLeavesNoGroup(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	name := fmt.Sprintf("tx-rollback-create-%d", time.Now().UnixNano())
	if _, err := repo.CreateGroup(ctx, testutil.OwnerID, domain.CreateGroupInput{
		Name: name,
		ResourcePermissions: []domain.ResourcePermission{
			{ResourceKey: overLongResourceKey(), PermissionLevel: domain.PermLevelViewer},
		},
	}); err == nil {
		t.Fatal("CreateGroup with an over-long resource_key: expected error, got nil")
	}

	// The rolled-back insert must not hold the (owner_id, name) unique key: a
	// retry under the same name has to succeed.
	groupID, err := repo.CreateGroup(ctx, testutil.OwnerID, domain.CreateGroupInput{
		Name: name,
		ResourcePermissions: []domain.ResourcePermission{
			{ResourceKey: "crm.customers", PermissionLevel: domain.PermLevelViewer},
		},
	})
	if err != nil {
		t.Fatalf("retry CreateGroup with the same name: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteGroup(ctx, testutil.OwnerID, groupID) })
}

func TestTeam_GroupWrites_RejectInvalidPermissionLevel(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	name := fmt.Sprintf("invalid-level-%d", time.Now().UnixNano())
	if _, err := repo.CreateGroup(ctx, testutil.OwnerID, domain.CreateGroupInput{
		Name: name,
		ResourcePermissions: []domain.ResourcePermission{
			{ResourceKey: "crm.customers", PermissionLevel: domain.PermissionLevel("admin")},
		},
	}); err != domain.ErrInvalidPermissionLevel {
		t.Fatalf("CreateGroup with an unknown permission level: err = %v, want ErrInvalidPermissionLevel", err)
	}

	groupID, err := repo.CreateGroup(ctx, testutil.OwnerID, domain.CreateGroupInput{
		Name: name,
		ResourcePermissions: []domain.ResourcePermission{
			{ResourceKey: "crm.customers", PermissionLevel: domain.PermLevelEditor},
		},
	})
	if err != nil {
		t.Fatalf("CreateGroup after rejected attempt: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteGroup(ctx, testutil.OwnerID, groupID) })

	if err := repo.UpdateGroup(ctx, testutil.OwnerID, groupID, domain.UpdateGroupInput{
		Name: name + "-renamed",
		ResourcePermissions: []domain.ResourcePermission{
			{ResourceKey: "crm.customers", PermissionLevel: domain.PermissionLevel("admin")},
		},
	}); err != domain.ErrInvalidPermissionLevel {
		t.Fatalf("UpdateGroup with an unknown permission level: err = %v, want ErrInvalidPermissionLevel", err)
	}

	group, err := repo.GetGroup(ctx, testutil.OwnerID, groupID)
	if err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	if group.Name != name {
		t.Errorf("group name = %q, want the pre-update %q after rollback", group.Name, name)
	}
	if len(group.ResourcePermissions) != 1 || group.ResourcePermissions[0].PermissionLevel != domain.PermLevelEditor {
		t.Errorf("resource permissions = %+v, want the stored crm.customers/editor", group.ResourcePermissions)
	}
}
