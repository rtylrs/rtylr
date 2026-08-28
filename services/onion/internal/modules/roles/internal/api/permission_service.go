package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/roles/internal/application"
	"github.com/voxire/onion/internal/modules/roles/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubpermissionv1 "github.com/voxire/voxie/gen/go/hub/permission/v1"
)

type PermissionAPI struct {
	Service *application.Service
}

func NewPermissionAPI(service *application.Service) *PermissionAPI {
	return &PermissionAPI{Service: service}
}

func (a *PermissionAPI) GetPermissions(ctx context.Context, _ *hubpermissionv1.GetPermissionsRequest) (*hubpermissionv1.GetPermissionsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !canViewRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	roles, err := a.Service.ListRoles(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list roles")
	}

	return &hubpermissionv1.GetPermissionsResponse{
		Matrix: buildPermissionMatrix(a.Service.ListPermissions(), roles),
	}, nil
}

func (a *PermissionAPI) ListRoles(ctx context.Context, req *hubpermissionv1.ListRolesRequest) (*hubpermissionv1.ListRolesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !canViewRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	rolesPage, err := a.Service.ListManagedRoles(ctx, businessID, domain.RoleFilter{
		AppID:    req.GetAppId(),
		Page:     int64(req.GetPage()),
		PageSize: int64(req.GetPageSize()),
	})
	if err != nil {
		return nil, roleAPIError(err, "Failed to list roles")
	}

	payloads := make([]*hubpermissionv1.CustomRole, 0, len(rolesPage.Roles))
	for _, role := range rolesPage.Roles {
		payloads = append(payloads, toProtoRole(role))
	}

	return &hubpermissionv1.ListRolesResponse{
		Roles:    payloads,
		Page:     int32(rolesPage.Page),     // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PageSize: int32(rolesPage.PageSize), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Total:    rolesPage.Total,
		NextPage: nextPage(rolesPage.Page, rolesPage.PageSize, rolesPage.Total),
	}, nil
}

func (a *PermissionAPI) CreateRole(ctx context.Context, req *hubpermissionv1.CreateRoleRequest) (*hubpermissionv1.CreateRoleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !canEditRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	name := req.GetName()
	if name == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	input := domain.CreateRoleInput{
		Name:         name,
		Description:  req.GetDescription(),
		AppID:        req.GetAppId(),
		RoleKey:      req.GetRoleKey(),
		Level:        req.GetLevel(),
		ParentRoleID: req.ParentRoleId,
		ScopeRules:   req.GetScopeRules(),
		Permissions:  req.GetPermissions(),
	}
	if !canDelegateRoleMutation(device.User, input.AppID, input.Permissions) {
		return nil, onekiterrors.PermissionDenied()
	}
	role, err := a.Service.CreateManagedRole(ctx, businessID, device.User.UserID, input)
	if err != nil {
		return nil, roleAPIError(err, "Failed to create role")
	}

	return &hubpermissionv1.CreateRoleResponse{
		Role: toProtoRole(role),
	}, nil
}

func (a *PermissionAPI) UpdatePermissionMatrix(ctx context.Context, req *hubpermissionv1.UpdatePermissionMatrixRequest) (*hubpermissionv1.UpdatePermissionMatrixResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !canEditRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	rolePage, err := a.Service.ListManagedRoles(ctx, businessID, domain.RoleFilter{})
	if err != nil {
		return nil, roleAPIError(err, "Failed to list roles")
	}

	roleName := req.GetRole()
	permKey := req.GetPermissionKey()
	var target *domain.Role
	for i := range rolePage.Roles {
		if rolePage.Roles[i].Name == roleName {
			target = &rolePage.Roles[i]
			break
		}
	}
	if target == nil {
		return nil, onekiterrors.New(http.StatusNotFound, "Role not found")
	}

	perms := target.Permissions
	if perms == nil {
		perms = []string{}
	}
	if req.GetEnabled() {
		found := false
		for _, p := range perms {
			if p == permKey {
				found = true
				break
			}
		}
		if !found {
			perms = append(perms, permKey)
		}
	} else {
		filtered := make([]string, 0, len(perms))
		for _, p := range perms {
			if p != permKey {
				filtered = append(filtered, p)
			}
		}
		perms = filtered
	}
	if !canDelegateRoleMutation(device.User, target.AppID, perms) {
		return nil, onekiterrors.PermissionDenied()
	}

	if _, err := a.Service.UpdateManagedRole(ctx, businessID, device.User.UserID, domain.UpdateRoleInput{
		ID:           target.ID,
		Name:         target.Name,
		Description:  target.Description,
		AppID:        target.AppID,
		Level:        target.Level,
		ParentRoleID: target.ParentRoleID,
		ScopeRules:   target.ScopeRules,
		Permissions:  perms,
	}); err != nil {
		return nil, roleAPIError(err, "Failed to update permission matrix")
	}

	return &hubpermissionv1.UpdatePermissionMatrixResponse{Success: true}, nil
}

func (a *PermissionAPI) UpdateRole(ctx context.Context, req *hubpermissionv1.UpdateRoleRequest) (*hubpermissionv1.UpdateRoleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if !canEditRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	id := req.GetId()
	if id <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Role id is required")
	}
	body := req.GetBody()
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Request body is required")
	}
	name := body.GetName()
	if name == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	role, roleErr := a.Service.GetManagedRole(ctx, businessID, id)
	if roleErr != nil {
		return nil, roleAPIError(roleErr, "Failed to get role")
	}

	perms := body.GetPermissions()
	if perms == nil {
		perms = []string{}
	}
	if !canDelegateRoleMutation(device.User, role.AppID, perms) {
		return nil, onekiterrors.PermissionDenied()
	}

	updatedRole, err := a.Service.UpdateManagedRole(ctx, businessID, device.User.UserID, domain.UpdateRoleInput{
		ID:           id,
		Name:         name,
		Description:  body.GetDescription(),
		AppID:        body.GetAppId(),
		Level:        body.GetLevel(),
		ParentRoleID: body.ParentRoleId,
		ScopeRules:   body.GetScopeRules(),
		Permissions:  perms,
	})
	if err != nil {
		return nil, roleAPIError(err, "Failed to update role")
	}

	return &hubpermissionv1.UpdateRoleResponse{Message: "Role updated", Role: toProtoRole(updatedRole)}, nil
}

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

func canViewRoles(user *authPkg.AuthUser) bool {
	if user == nil {
		return false
	}
	if authPkg.LegacyAuthorizationMode() {
		return authPkg.IsBusinessAdmin(user)
	}
	if user.EffectiveAccess != nil {
		return authPkg.HasDashboardPermission(user, "hub:roles:view")
	}
	return authPkg.IsBusinessAdmin(user)
}

func canEditRoles(user *authPkg.AuthUser) bool {
	if user == nil {
		return false
	}
	if authPkg.LegacyAuthorizationMode() {
		return authPkg.IsBusinessAdmin(user)
	}
	if user.EffectiveAccess != nil {
		return authPkg.HasDashboardPermission(user, "hub:roles:create") ||
			authPkg.HasDashboardPermission(user, "hub:roles:edit")
	}
	return authPkg.IsBusinessAdmin(user)
}

func canArchiveRoles(user *authPkg.AuthUser) bool {
	if user == nil {
		return false
	}
	if authPkg.LegacyAuthorizationMode() {
		return authPkg.IsBusinessAdmin(user)
	}
	if user.EffectiveAccess != nil {
		return authPkg.HasDashboardPermission(user, "hub:roles:archive")
	}
	return authPkg.IsBusinessAdmin(user)
}

func canDelegateRoleMutation(user *authPkg.AuthUser, appID string, permissions []string) bool {
	if user == nil {
		return false
	}
	if user.EffectiveAccess == nil {
		return authPkg.IsBusinessAdmin(user)
	}
	return authorization.CanDelegatePermissions(user.EffectiveAccess, authorization.AppID(appID), permissions)
}

func roleAPIError(err error, fallback string) error {
	switch {
	case errors.Is(err, domain.ErrRoleNotFound):
		return onekiterrors.New(http.StatusNotFound, "Role not found")
	case errors.Is(err, domain.ErrSystemRoleMutation), errors.Is(err, domain.ErrRoleArchived):
		return onekiterrors.New(http.StatusConflict, "Role cannot be changed")
	case errors.Is(err, domain.ErrInvalidPermission), errors.Is(err, domain.ErrInvalidAssignmentTarget), errors.Is(err, domain.ErrInvalidScope):
		return onekiterrors.New(http.StatusBadRequest, "Invalid role authorization input")
	case errors.Is(err, domain.ErrDelegationCeiling):
		return onekiterrors.PermissionDenied()
	case errors.Is(err, domain.ErrAuthorizationSchemaUnavailable):
		return onekiterrors.New(http.StatusInternalServerError, "Role authorization is not available")
	default:
		return onekiterrors.New(http.StatusInternalServerError, fallback)
	}
}

func nextPage(page, pageSize, total int64) *int32 {
	if page <= 0 || pageSize <= 0 || page*pageSize >= total {
		return nil
	}
	next := int32(page + 1) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	return &next
}

func buildPermissionMatrix(entries []domain.PermissionEntry, roles []domain.Role) *hubpermissionv1.PermissionMatrix {
	roleNames := make([]string, 0, len(roles))
	rolePermSets := make(map[string]map[string]struct{}, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
		perms := make(map[string]struct{}, len(role.Permissions))
		for _, p := range role.Permissions {
			perms[p] = struct{}{}
		}
		rolePermSets[role.Name] = perms
	}

	rows := make([]*hubpermissionv1.PermissionRow, 0, len(entries))
	for _, entry := range entries {
		roleValues := make(map[string]bool, len(roleNames))
		for _, roleName := range roleNames {
			_, enabled := rolePermSets[roleName][entry.Key]
			roleValues[roleName] = enabled
		}
		rows = append(rows, &hubpermissionv1.PermissionRow{
			Key:        entry.Key,
			Category:   entry.Category,
			Label:      entry.Label,
			RoleValues: roleValues,
		})
	}

	return &hubpermissionv1.PermissionMatrix{
		Rows:  rows,
		Roles: roleNames,
	}
}

func toProtoRole(role domain.Role) *hubpermissionv1.CustomRole {
	perms := role.Permissions
	if perms == nil {
		perms = []string{}
	}
	return &hubpermissionv1.CustomRole{
		Id:           role.ID,
		Name:         role.Name,
		Description:  role.Description,
		Permissions:  perms,
		AppId:        role.AppID,
		RoleKey:      role.RoleKey,
		Level:        role.Level,
		ParentRoleId: role.ParentRoleID,
		ScopeRules:   role.ScopeRules,
		IsSystem:     role.IsSystem,
		IsArchived:   role.IsArchived,
		Lifecycle:    role.Lifecycle,
		Version:      role.Version,
		MemberCount:  role.MemberCount,
		CreatedAt:    role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toProtoAssignment(assignment domain.RoleAssignment) *hubpermissionv1.RoleAssignment {
	return &hubpermissionv1.RoleAssignment{
		Id:        assignment.ID,
		RoleId:    assignment.RoleID,
		AppId:     assignment.AppID,
		UserId:    assignment.UserID,
		GroupId:   assignment.GroupID,
		Status:    assignment.Status,
		Source:    assignment.Source,
		IsPrimary: assignment.IsPrimary,
		Scopes:    assignment.Scopes,
		GrantedAt: assignment.GrantedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toProtoAuditEvent(event domain.AuditEvent) *hubpermissionv1.RoleAuditEvent {
	return &hubpermissionv1.RoleAuditEvent{
		Id:            event.ID,
		Action:        event.Action,
		AppId:         event.AppID,
		RoleId:        event.RoleID,
		ActorUserId:   event.ActorUserID,
		TargetUserId:  event.TargetUserID,
		TargetGroupId: event.TargetGroupID,
		BeforeJson:    event.BeforeJSON,
		AfterJson:     event.AfterJSON,
		Reason:        event.Reason,
		CreatedAt:     event.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
