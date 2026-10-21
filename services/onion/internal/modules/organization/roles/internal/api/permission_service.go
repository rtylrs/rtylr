package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/organization/roles/internal/application"
	"github.com/voxire/onion/internal/modules/organization/roles/internal/domain"
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

func (a *PermissionAPI) ListPermissionCatalog(ctx context.Context, _ *hubpermissionv1.ListPermissionCatalogRequest) (*hubpermissionv1.ListPermissionCatalogResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil || !canViewRoles(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	if _, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	permissions, err := a.Service.ListPermissionCatalog(ctx)
	if err != nil {
		return nil, roleAPIError(err, "Failed to list permission catalog")
	}
	entries := make([]*hubpermissionv1.PermissionDefinition, 0, len(permissions))
	for _, permission := range permissions {
		entries = append(entries, &hubpermissionv1.PermissionDefinition{
			Key: permission.Key, AppId: permission.AppID, Resource: permission.Resource,
			Action: permission.Action, IsSensitive: permission.IsSensitive, IsHighRisk: permission.IsHighRisk,
		})
	}
	return &hubpermissionv1.ListPermissionCatalogResponse{Permissions: entries}, nil
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
	// Both ends of the change must be within the actor's ceiling: the permissions
	// being added, and the ones being taken away from the role's current holders.
	if !canDelegateRoleMutation(device.User, role.AppID, perms) ||
		!canDelegateRoleMutation(device.User, role.AppID, role.Permissions) {
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
