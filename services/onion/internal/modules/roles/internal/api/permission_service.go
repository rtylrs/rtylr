package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/roles/internal/application"
	"github.com/voxire/onion/internal/modules/roles/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
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
