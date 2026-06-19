package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/roles/internal/application"
	"github.com/voxire/onion/internal/modules/roles/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func RoleRoutes() []string       { return []string{"/roles"} }
func PermissionRoutes() []string { return []string{"/permissions"} }

// -- List Roles --

type (
	ListRolesHandler struct{ Service *application.Service }
	ListRolesRequest struct{}
)

type RolePayload struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListRolesResponse struct {
	Roles []RolePayload `json:"roles"`
}

func (h *ListRolesHandler) Serve(ctx context.Context, _ *ListRolesRequest) (ListRolesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListRolesResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListRolesResponse{}, errorModels.PermissionDeniedServerError()
	}

	roles, err := h.Service.ListRoles(ctx, businessID)
	if err != nil {
		return ListRolesResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list roles",
		}
	}

	payloads := make([]RolePayload, 0, len(roles))
	for _, r := range roles {
		perms := r.Permissions
		if perms == nil {
			perms = []string{}
		}
		payloads = append(payloads, RolePayload{
			ID:          r.ID,
			Name:        r.Name,
			Permissions: perms,
			IsSystem:    r.IsSystem,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		})
	}

	return ListRolesResponse{Roles: payloads}, nil
}

// -- Create Role --

type CreateRoleHandler struct{ Service *application.Service }

type CreateRoleRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type CreateRoleResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateRoleHandler) Validate(_ context.Context, req *CreateRoleRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Name is required",
		}
	}
	return nil
}

func (h *CreateRoleHandler) Serve(ctx context.Context, req *CreateRoleRequest) (CreateRoleResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateRoleResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateRoleResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := h.Service.CreateRole(ctx, businessID, domain.CreateRoleInput{
		Name:        req.Name,
		Permissions: req.Permissions,
	})
	if err != nil {
		return CreateRoleResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to create role",
		}
	}

	return CreateRoleResponse{ID: id, Message: "Role created"}, nil
}

// -- Update Role --

type UpdateRoleHandler struct{ Service *application.Service }

type UpdateRoleRequest struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type UpdateRoleResponse struct {
	Message string `json:"message"`
}

func (h *UpdateRoleHandler) Validate(_ context.Context, req *UpdateRoleRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Role id is required",
		}
	}
	if req.Name == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Name is required",
		}
	}
	return nil
}

func (h *UpdateRoleHandler) Serve(ctx context.Context, req *UpdateRoleRequest) (UpdateRoleResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateRoleResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateRoleResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.UpdateRole(ctx, businessID, domain.UpdateRoleInput{
		ID:          req.ID,
		Name:        req.Name,
		Permissions: req.Permissions,
	}); err != nil {
		return UpdateRoleResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to update role",
		}
	}

	return UpdateRoleResponse{Message: "Role updated"}, nil
}

// -- Delete Role --

type DeleteRoleHandler struct{ Service *application.Service }

type DeleteRoleRequest struct {
	ID int64 `json:"id"`
}

type DeleteRoleResponse struct {
	Message string `json:"message"`
}

func (h *DeleteRoleHandler) Validate(_ context.Context, req *DeleteRoleRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Role id is required",
		}
	}
	return nil
}

func (h *DeleteRoleHandler) Serve(ctx context.Context, req *DeleteRoleRequest) (DeleteRoleResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteRoleResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteRoleResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.DeleteRole(ctx, businessID, req.ID); err != nil {
		return DeleteRoleResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to delete role",
		}
	}

	return DeleteRoleResponse{Message: "Role deleted"}, nil
}

// -- List Permissions --

type (
	ListPermissionsHandler struct{ Service *application.Service }
	ListPermissionsRequest struct{}
)

type PermissionPayload struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type ListPermissionsResponse struct {
	Permissions []PermissionPayload `json:"permissions"`
}

func (h *ListPermissionsHandler) Serve(_ context.Context, _ *ListPermissionsRequest) (ListPermissionsResponse, *errorModels.ServerError) {
	entries := h.Service.ListPermissions()

	payloads := make([]PermissionPayload, 0, len(entries))
	for _, e := range entries {
		payloads = append(payloads, PermissionPayload{
			Key:         e.Key,
			Label:       e.Label,
			Description: e.Description,
			Category:    e.Category,
		})
	}

	return ListPermissionsResponse{Permissions: payloads}, nil
}
