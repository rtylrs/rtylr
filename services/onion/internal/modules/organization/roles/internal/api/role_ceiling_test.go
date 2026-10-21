package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/roles/internal/application"
	"github.com/voxire/onion/internal/modules/organization/roles/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/tenant"
	hubpermissionv1 "github.com/voxire/voxie/gen/go/hub/permission/v1"
)

type businessScope struct{}

func (businessScope) ActiveBusinessOwner(context.Context, int64) (int64, error) { return 1, nil }
func (businessScope) UserBusiness(_ context.Context, businessID, _ int64) (int64, error) {
	return businessID, nil
}

type ceilingRepo struct {
	roles       map[int64]domain.Role
	assignments map[int64]domain.RoleAssignment
	removed     []int64
	archived    []int64
	updated     []int64
}

func (r *ceilingRepo) ListPermissionCatalog(context.Context) ([]domain.PermissionEntry, error) {
	return nil, nil
}
func (r *ceilingRepo) ListNormalized(context.Context, int64, domain.RoleFilter) (domain.RolePage, error) {
	return domain.RolePage{}, nil
}
func (r *ceilingRepo) GetNormalized(_ context.Context, _, roleID int64) (domain.Role, error) {
	role, ok := r.roles[roleID]
	if !ok {
		return domain.Role{}, domain.ErrRoleNotFound
	}
	return role, nil
}
func (r *ceilingRepo) CreateNormalized(context.Context, int64, int64, domain.CreateRoleInput) (domain.Role, error) {
	return domain.Role{}, nil
}
func (r *ceilingRepo) UpdateNormalized(_ context.Context, _, _ int64, input domain.UpdateRoleInput) (domain.Role, error) {
	r.updated = append(r.updated, input.ID)
	return r.roles[input.ID], nil
}
func (r *ceilingRepo) ArchiveNormalized(_ context.Context, _, _, roleID int64, _ string) (domain.Role, error) {
	r.archived = append(r.archived, roleID)
	return r.roles[roleID], nil
}
func (r *ceilingRepo) CloneNormalized(context.Context, int64, int64, int64, string, string) (domain.Role, error) {
	return domain.Role{}, nil
}
func (r *ceilingRepo) ListAssignments(context.Context, int64, domain.AssignmentFilter) (domain.RoleAssignmentPage, error) {
	return domain.RoleAssignmentPage{}, nil
}
func (r *ceilingRepo) AssignRole(context.Context, int64, int64, domain.AssignRoleInput) (domain.RoleAssignment, error) {
	return domain.RoleAssignment{}, nil
}
func (r *ceilingRepo) RemoveAssignment(_ context.Context, _, _, assignmentID int64) error {
	r.removed = append(r.removed, assignmentID)
	return nil
}
func (r *ceilingRepo) GetAssignment(_ context.Context, _, assignmentID int64) (domain.RoleAssignment, error) {
	assignment, ok := r.assignments[assignmentID]
	if !ok {
		return domain.RoleAssignment{}, domain.ErrRoleNotFound
	}
	return assignment, nil
}
func (r *ceilingRepo) PreviewPermissionImpact(context.Context, int64, int64, []string) (domain.PermissionImpact, error) {
	return domain.PermissionImpact{}, nil
}
func (r *ceilingRepo) ListAudit(context.Context, int64, domain.AuditFilter) (domain.AuditPage, error) {
	return domain.AuditPage{}, nil
}

const (
	cashierRole = int64(1)
	managerRole = int64(2)
)

func ceilingFixture(t *testing.T) (*PermissionAPI, *ceilingRepo, context.Context) {
	t.Helper()
	t.Cleanup(tenant.SetScopeResolver(businessScope{}))
	repo := &ceilingRepo{
		roles: map[int64]domain.Role{
			cashierRole: {ID: cashierRole, AppID: "pos", Name: "Cashier", Permissions: []string{"pos:orders:view"}},
			managerRole: {ID: managerRole, AppID: "pos", Name: "Manager", Permissions: []string{"pos:orders:view", "pos:refunds:approve"}},
		},
		assignments: map[int64]domain.RoleAssignment{
			100: {ID: 100, RoleID: cashierRole, AppID: "pos"},
			200: {ID: 200, RoleID: managerRole, AppID: "pos"},
		},
	}
	api := NewPermissionAPI(application.NewService(context.Background(), repo))
	// A delegate who can edit and archive roles but only holds cashier-level POS access.
	user := &authPkg.AuthUser{UserID: 7, BusinessID: 1, EffectiveAccess: &authorization.EffectiveAccess{
		DashPermissions: []string{"hub:roles:edit", "hub:roles:archive"},
		Apps:            []authorization.EffectiveAppAccess{{AppID: "pos", HasAccess: true, Permissions: []string{"pos:orders:view"}}},
	}}
	return api, repo, (&authPkg.Device{User: user}).Ctx(context.Background())
}

func TestRemoveRoleAssignmentEnforcesTheGrantingCeiling(t *testing.T) {
	api, repo, ctx := ceilingFixture(t)

	if _, err := api.RemoveRoleAssignment(ctx, &hubpermissionv1.RemoveRoleAssignmentRequest{Id: 200}); err == nil {
		t.Fatal("a delegate revoked a manager-level assignment they could not have granted")
	}
	if _, err := api.RemoveRoleAssignment(ctx, &hubpermissionv1.RemoveRoleAssignmentRequest{Id: 100}); err != nil {
		t.Fatalf("revoking a cashier-level assignment: %v", err)
	}
	if len(repo.removed) != 1 || repo.removed[0] != 100 {
		t.Fatalf("removed = %v, want only 100", repo.removed)
	}
}

func TestRemoveRoleAssignmentReportsAMissingAssignmentAsNotFound(t *testing.T) {
	api, _, ctx := ceilingFixture(t)
	if _, err := api.RemoveRoleAssignment(ctx, &hubpermissionv1.RemoveRoleAssignmentRequest{Id: 999}); err == nil {
		t.Fatal("removing a missing assignment succeeded")
	}
}

func TestArchivingARoleEnforcesTheGrantingCeiling(t *testing.T) {
	api, repo, ctx := ceilingFixture(t)

	if _, err := api.ArchiveRole(ctx, &hubpermissionv1.ArchiveRoleRequest{Id: managerRole}); err == nil {
		t.Fatal("a delegate archived a role above their ceiling")
	}
	if _, err := api.DeleteRole(ctx, &hubpermissionv1.DeleteRoleRequest{Id: managerRole}); err == nil {
		t.Fatal("a delegate deleted a role above their ceiling")
	}
	if _, err := api.ArchiveRole(ctx, &hubpermissionv1.ArchiveRoleRequest{Id: cashierRole}); err != nil {
		t.Fatalf("archiving a cashier role: %v", err)
	}
	if len(repo.archived) != 1 || repo.archived[0] != cashierRole {
		t.Fatalf("archived = %v", repo.archived)
	}
}

func TestUpdatingARoleCannotStripPermissionsAboveTheCeiling(t *testing.T) {
	api, repo, ctx := ceilingFixture(t)

	// The new permission set is within the ceiling, but it removes
	// pos:refunds:approve from every manager.
	_, err := api.UpdateRole(ctx, &hubpermissionv1.UpdateRoleRequest{
		Id:   managerRole,
		Body: &hubpermissionv1.UpdateRoleBody{Name: "Manager", Permissions: []string{"pos:orders:view"}},
	})
	if err == nil || len(repo.updated) != 0 {
		t.Fatalf("a delegate stripped permissions above their ceiling: err=%v updated=%v", err, repo.updated)
	}
}
