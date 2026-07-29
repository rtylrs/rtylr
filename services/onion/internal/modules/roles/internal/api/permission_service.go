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
	if device.User == nil {
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

func (a *PermissionAPI) ListRoles(ctx context.Context, _ *hubpermissionv1.ListRolesRequest) (*hubpermissionv1.ListRolesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
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

	payloads := make([]*hubpermissionv1.CustomRole, 0, len(roles))
	for _, role := range roles {
		payloads = append(payloads, toProtoRole(role))
	}

	return &hubpermissionv1.ListRolesResponse{Roles: payloads}, nil
}

func (a *PermissionAPI) CreateRole(ctx context.Context, req *hubpermissionv1.CreateRoleRequest) (*hubpermissionv1.CreateRoleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
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

	id, err := a.Service.CreateRole(ctx, businessID, domain.CreateRoleInput{
		Name:        name,
		Permissions: req.GetPermissions(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create role")
	}

	perms := req.GetPermissions()
	if perms == nil {
		perms = []string{}
	}

	return &hubpermissionv1.CreateRoleResponse{
		Role: &hubpermissionv1.CustomRole{
			Id:          id,
			Name:        name,
			Description: req.GetDescription(),
			Permissions: perms,
		},
	}, nil
}

func (a *PermissionAPI) UpdatePermissionMatrix(ctx context.Context, req *hubpermissionv1.UpdatePermissionMatrixRequest) (*hubpermissionv1.UpdatePermissionMatrixResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
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

	roleName := req.GetRole()
	permKey := req.GetPermissionKey()
	var target *domain.Role
	for i := range roles {
		if roles[i].Name == roleName {
			target = &roles[i]
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

	if err := a.Service.UpdateRole(ctx, businessID, domain.UpdateRoleInput{
		ID:          target.ID,
		Name:        target.Name,
		Permissions: perms,
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update permission matrix")
	}

	return &hubpermissionv1.UpdatePermissionMatrixResponse{Success: true}, nil
}

func (a *PermissionAPI) UpdateRole(ctx context.Context, req *hubpermissionv1.UpdateRoleRequest) (*hubpermissionv1.UpdateRoleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
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

	perms := body.GetPermissions()
	if perms == nil {
		perms = []string{}
	}

	if err := a.Service.UpdateRole(ctx, businessID, domain.UpdateRoleInput{
		ID:          id,
		Name:        name,
		Permissions: perms,
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update role")
	}

	return &hubpermissionv1.UpdateRoleResponse{Message: "Role updated"}, nil
}

func (a *PermissionAPI) DeleteRole(ctx context.Context, req *hubpermissionv1.DeleteRoleRequest) (*hubpermissionv1.DeleteRoleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
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

	if err := a.Service.DeleteRole(ctx, businessID, id); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete role")
	}

	return &hubpermissionv1.DeleteRoleResponse{Message: "Role deleted"}, nil
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
		Id:          role.ID,
		Name:        role.Name,
		Permissions: perms,
		CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
