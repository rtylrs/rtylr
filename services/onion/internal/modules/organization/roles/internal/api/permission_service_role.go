package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/organization/roles/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubpermissionv1 "github.com/voxire/voxie/gen/go/hub/permission/v1"
)

func (a *PermissionAPI) DeleteRole(ctx context.Context, req *hubpermissionv1.DeleteRoleRequest) (*hubpermissionv1.DeleteRoleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !canArchiveRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	id := req.GetId()
	if id <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Role id is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.requireRoleCeiling(ctx, device.User, businessID, id); err != nil {
		return nil, err
	}
	if _, err := a.Service.ArchiveManagedRole(ctx, businessID, device.User.UserID, id, ""); err != nil {
		return nil, roleAPIError(err, "Failed to archive role")
	}

	return &hubpermissionv1.DeleteRoleResponse{Message: "Role archived"}, nil
}

func (a *PermissionAPI) GetRole(ctx context.Context, req *hubpermissionv1.GetRoleRequest) (*hubpermissionv1.GetRoleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !canViewRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	role, err := a.Service.GetManagedRole(ctx, businessID, req.GetId())
	if err != nil {
		return nil, roleAPIError(err, "Failed to get role")
	}
	return &hubpermissionv1.GetRoleResponse{Role: toProtoRole(role)}, nil
}

func (a *PermissionAPI) CloneRole(ctx context.Context, req *hubpermissionv1.CloneRoleRequest) (*hubpermissionv1.CloneRoleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !canEditRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	sourceRole, sourceErr := a.Service.GetManagedRole(ctx, businessID, req.GetId())
	if sourceErr != nil {
		return nil, roleAPIError(sourceErr, "Failed to get role")
	}
	if !canDelegateRoleMutation(device.User, sourceRole.AppID, sourceRole.Permissions) {
		return nil, onekiterrors.PermissionDenied()
	}
	role, err := a.Service.CloneManagedRole(ctx, businessID, device.User.UserID, req.GetId(), req.GetName(), req.GetDescription())
	if err != nil {
		return nil, roleAPIError(err, "Failed to clone role")
	}
	return &hubpermissionv1.CloneRoleResponse{Role: toProtoRole(role)}, nil
}

func (a *PermissionAPI) ArchiveRole(ctx context.Context, req *hubpermissionv1.ArchiveRoleRequest) (*hubpermissionv1.ArchiveRoleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !canArchiveRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.requireRoleCeiling(ctx, device.User, businessID, req.GetId()); err != nil {
		return nil, err
	}
	role, err := a.Service.ArchiveManagedRole(ctx, businessID, device.User.UserID, req.GetId(), req.GetReason())
	if err != nil {
		return nil, roleAPIError(err, "Failed to archive role")
	}
	return &hubpermissionv1.ArchiveRoleResponse{Role: toProtoRole(role)}, nil
}

func (a *PermissionAPI) ListRoleAssignments(ctx context.Context, req *hubpermissionv1.ListRoleAssignmentsRequest) (*hubpermissionv1.ListRoleAssignmentsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !canViewRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	filter := domain.AssignmentFilter{AppID: req.GetAppId(), Page: int64(req.GetPage()), PageSize: int64(req.GetPageSize())}
	if req.RoleId != nil {
		filter.RoleID = req.RoleId
	}
	page, err := a.Service.ListManagedAssignments(ctx, businessID, filter)
	if err != nil {
		return nil, roleAPIError(err, "Failed to list role assignments")
	}
	assignments := make([]*hubpermissionv1.RoleAssignment, 0, len(page.Assignments))
	for _, assignment := range page.Assignments {
		assignments = append(assignments, toProtoAssignment(assignment))
	}
	return &hubpermissionv1.ListRoleAssignmentsResponse{
		Assignments: assignments,
		Page:        int32(page.Page),     // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PageSize:    int32(page.PageSize), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Total:       page.Total,
		NextPage:    nextPage(page.Page, page.PageSize, page.Total),
	}, nil
}

func (a *PermissionAPI) AssignRole(ctx context.Context, req *hubpermissionv1.AssignRoleRequest) (*hubpermissionv1.AssignRoleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !canEditRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	role, roleErr := a.Service.GetManagedRole(ctx, businessID, req.GetRoleId())
	if roleErr != nil {
		return nil, roleAPIError(roleErr, "Failed to get role")
	}
	if !canDelegateRoleMutation(device.User, role.AppID, role.Permissions) {
		return nil, onekiterrors.PermissionDenied()
	}
	assignment, err := a.Service.AssignManagedRole(ctx, businessID, device.User.UserID, domain.AssignRoleInput{
		RoleID:    req.GetRoleId(),
		AppID:     req.GetAppId(),
		UserID:    req.UserId,
		GroupID:   req.GroupId,
		IsPrimary: req.GetIsPrimary(),
		Scopes:    req.GetScopes(),
	})
	if err != nil {
		return nil, roleAPIError(err, "Failed to assign role")
	}
	return &hubpermissionv1.AssignRoleResponse{Assignment: toProtoAssignment(assignment)}, nil
}

func (a *PermissionAPI) RemoveRoleAssignment(ctx context.Context, req *hubpermissionv1.RemoveRoleAssignmentRequest) (*hubpermissionv1.RemoveRoleAssignmentResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !canEditRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	assignment, err := a.Service.GetManagedAssignment(ctx, businessID, req.GetId())
	if err != nil {
		return nil, roleAPIError(err, "Failed to get role assignment")
	}
	if err := a.requireRoleCeiling(ctx, device.User, businessID, assignment.RoleID); err != nil {
		return nil, err
	}
	if err := a.Service.RemoveManagedAssignment(ctx, businessID, device.User.UserID, req.GetId()); err != nil {
		return nil, roleAPIError(err, "Failed to revoke role assignment")
	}
	return &hubpermissionv1.RemoveRoleAssignmentResponse{Message: "Role assignment revoked"}, nil
}

func (a *PermissionAPI) PreviewPermissionImpact(ctx context.Context, req *hubpermissionv1.PreviewPermissionImpactRequest) (*hubpermissionv1.PreviewPermissionImpactResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !canEditRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	impact, err := a.Service.PreviewManagedPermissionImpact(ctx, businessID, req.GetId(), req.GetPermissions())
	if err != nil {
		return nil, roleAPIError(err, "Failed to preview permission impact")
	}
	return &hubpermissionv1.PreviewPermissionImpactResponse{
		AddedPermissions:    impact.AddedPermissions,
		RemovedPermissions:  impact.RemovedPermissions,
		AffectedMemberCount: impact.AffectedMemberCount,
	}, nil
}

func (a *PermissionAPI) ListRoleAudit(ctx context.Context, req *hubpermissionv1.ListRoleAuditRequest) (*hubpermissionv1.ListRoleAuditResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !canViewRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	filter := domain.AuditFilter{Page: int64(req.GetPage()), PageSize: int64(req.GetPageSize())}
	if req.RoleId != nil {
		filter.RoleID = req.RoleId
	}
	page, err := a.Service.ListManagedAudit(ctx, businessID, filter)
	if err != nil {
		return nil, roleAPIError(err, "Failed to list role audit history")
	}
	events := make([]*hubpermissionv1.RoleAuditEvent, 0, len(page.Events))
	for _, event := range page.Events {
		events = append(events, toProtoAuditEvent(event))
	}
	return &hubpermissionv1.ListRoleAuditResponse{
		Events:   events,
		Page:     int32(page.Page),     // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PageSize: int32(page.PageSize), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Total:    page.Total,
		NextPage: nextPage(page.Page, page.PageSize, page.Total),
	}, nil
}

// requireRoleCeiling applies the grant-time delegation ceiling to anything that
// takes a role's permissions away from its holders: revoking an assignment or
// archiving the role strips exactly what assigning it conferred.
func (a *PermissionAPI) requireRoleCeiling(ctx context.Context, user *authPkg.AuthUser, businessID, roleID int64) error {
	role, err := a.Service.GetManagedRole(ctx, businessID, roleID)
	if err != nil {
		return roleAPIError(err, "Failed to get role")
	}
	if !canDelegateRoleMutation(user, role.AppID, role.Permissions) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}
