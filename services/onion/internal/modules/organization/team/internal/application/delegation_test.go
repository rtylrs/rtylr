package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
	"github.com/voxire/onion/internal/platform/authorization"
)

type fakeDelegationStore struct {
	roles  map[int64]domain.AuthorizationRoleGrant
	groups map[int64]*domain.Group
}

func (f fakeDelegationStore) GetAuthorizationRoleGrant(_ context.Context, _, roleID int64) (domain.AuthorizationRoleGrant, error) {
	grant, ok := f.roles[roleID]
	if !ok {
		return domain.AuthorizationRoleGrant{}, errors.New("unknown role")
	}
	return grant, nil
}

func (f fakeDelegationStore) GetGroup(_ context.Context, _, groupID int64) (*domain.Group, error) {
	group, ok := f.groups[groupID]
	if !ok {
		return nil, errors.New("unknown group")
	}
	return group, nil
}

const (
	cashierRole = int64(1)
	managerRole = int64(2)
)

func delegationFixture() (fakeDelegationStore, *authorization.EffectiveAccess) {
	store := fakeDelegationStore{
		roles: map[int64]domain.AuthorizationRoleGrant{
			cashierRole: {AppID: "pos", Permissions: []string{"pos:orders:view"}},
			managerRole: {AppID: "pos", Permissions: []string{"pos:orders:view", "pos:refunds:approve"}},
		},
		groups: map[int64]*domain.Group{
			10: {ID: 10, PermissionProfileAssignments: []authorization.RoleAssignment{{AppID: "pos", RoleIDs: []int64{managerRole}}}},
			11: {ID: 11, PermissionProfileAssignments: []authorization.RoleAssignment{{AppID: "pos", RoleIDs: []int64{cashierRole}}}},
		},
	}
	// A delegate who may edit roles and teams but only holds cashier-level POS access.
	access := &authorization.EffectiveAccess{
		DashPermissions: []string{"hub:roles:edit", "hub:team:edit"},
		Apps:            []authorization.EffectiveAppAccess{{AppID: "pos", HasAccess: true, Permissions: []string{"pos:orders:view"}}},
	}
	return store, access
}

func roleInput(roleIDs ...int64) *[]domain.AppRoleAssignmentInput {
	return &[]domain.AppRoleAssignmentInput{{AppID: "pos", RoleIDs: roleIDs}}
}

func currentWith(roleIDs ...int64) *delegationCurrent {
	var grants []roleGrant
	for _, id := range roleIDs {
		grants = append(grants, roleGrant{AppID: "pos", RoleID: id})
	}
	return &delegationCurrent{MemberType: "member", Grants: grants}
}

func TestDelegateCannotGrantARoleAboveTheirCeiling(t *testing.T) {
	store, access := delegationFixture()
	err := checkDelegation(context.Background(), store, access, 1, domain.UpdatePermissionsInput{PermissionProfileAssignments: roleInput(managerRole)}, currentWith())
	if !errors.Is(err, domain.ErrDelegationCeiling) {
		t.Fatalf("err = %v, want ErrDelegationCeiling", err)
	}
}

func TestDelegateCannotRemoveARoleAboveTheirCeiling(t *testing.T) {
	store, access := delegationFixture()
	err := checkDelegation(context.Background(), store, access, 1, domain.UpdatePermissionsInput{PermissionProfileAssignments: roleInput(cashierRole)}, currentWith(cashierRole, managerRole))
	if !errors.Is(err, domain.ErrDelegationCeiling) {
		t.Fatalf("stripping the manager role: err = %v, want ErrDelegationCeiling", err)
	}
}

func TestDelegateCannotClearAllRolesWithoutRoleEditPermission(t *testing.T) {
	store, access := delegationFixture()
	access.DashPermissions = []string{"hub:team:edit"}
	err := checkDelegation(context.Background(), store, access, 1, domain.UpdatePermissionsInput{PermissionProfileAssignments: &[]domain.AppRoleAssignmentInput{}}, currentWith(cashierRole))
	if !errors.Is(err, domain.ErrDelegationCeiling) {
		t.Fatalf("err = %v, want ErrDelegationCeiling", err)
	}
}

func TestDelegateMayRemoveARoleWithinTheirCeiling(t *testing.T) {
	store, access := delegationFixture()
	err := checkDelegation(context.Background(), store, access, 1, domain.UpdatePermissionsInput{PermissionProfileAssignments: &[]domain.AppRoleAssignmentInput{}}, currentWith(cashierRole))
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
}

func TestResendingUnchangedRolesNeedsNoRolePermission(t *testing.T) {
	store, access := delegationFixture()
	access.DashPermissions = []string{"hub:team:edit"}
	err := checkDelegation(context.Background(), store, access, 1, domain.UpdatePermissionsInput{PermissionProfileAssignments: roleInput(cashierRole, managerRole)}, currentWith(cashierRole, managerRole))
	if err != nil {
		t.Fatalf("a no-op role list was refused: %v", err)
	}
}

func TestDelegateCannotAddOrRemoveGroupsThatCarryHigherRoles(t *testing.T) {
	store, access := delegationFixture()
	add := checkDelegation(context.Background(), store, access, 1, domain.UpdatePermissionsInput{GroupIDs: &[]int64{10}}, &delegationCurrent{})
	if !errors.Is(add, domain.ErrDelegationCeiling) {
		t.Fatalf("joining the manager group: err = %v", add)
	}
	remove := checkDelegation(context.Background(), store, access, 1, domain.UpdatePermissionsInput{GroupIDs: &[]int64{}}, &delegationCurrent{GroupIDs: []int64{10}})
	if !errors.Is(remove, domain.ErrDelegationCeiling) {
		t.Fatalf("leaving the manager group: err = %v", remove)
	}
	if err := checkDelegation(context.Background(), store, access, 1, domain.UpdatePermissionsInput{GroupIDs: &[]int64{11}}, &delegationCurrent{}); err != nil {
		t.Fatalf("joining a cashier group: %v", err)
	}
}

func TestDelegateCannotDemoteAnAdmin(t *testing.T) {
	store, access := delegationFixture()
	err := checkDelegation(context.Background(), store, access, 1, domain.UpdatePermissionsInput{MemberType: "member"}, &delegationCurrent{MemberType: "admin"})
	if !errors.Is(err, domain.ErrDelegationCeiling) {
		t.Fatalf("err = %v, want ErrDelegationCeiling", err)
	}
}

func TestGroupMembershipChangesFollowTheGroupsRoles(t *testing.T) {
	store, access := delegationFixture()
	if grantsWithinCeiling(context.Background(), store, 1, access, grantsOfGroupAssignments(store.groups[10].PermissionProfileAssignments)) {
		t.Fatal("a delegate was allowed to change membership of a group carrying a role above their ceiling")
	}
	if !grantsWithinCeiling(context.Background(), store, 1, access, grantsOfGroupAssignments(store.groups[11].PermissionProfileAssignments)) {
		t.Fatal("membership of a cashier group was refused")
	}
}

func TestScopeChangeCountsAsGrantAndRevoke(t *testing.T) {
	scoped := []authorization.Scope{{Type: "location", ID: "5"}}
	added, removed := diffGrants(
		[]roleGrant{{AppID: "pos", RoleID: cashierRole, Scopes: scopeSignature(nil)}},
		[]roleGrant{{AppID: "pos", RoleID: cashierRole, Scopes: scopeSignature(scoped)}},
	)
	if len(added) != 1 || len(removed) != 1 {
		t.Fatalf("added=%v removed=%v", added, removed)
	}
}
