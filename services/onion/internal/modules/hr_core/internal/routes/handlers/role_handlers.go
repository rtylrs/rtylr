package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

type RolePayload struct {
	ID          int64  `json:"id"`
	OwnerID     int64  `json:"owner_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Permissions string `json:"permissions"`
	IsSystem    bool   `json:"is_system"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func roleToPayload(r domain.RoleRecord) *RolePayload {
	return &RolePayload{
		ID: r.ID, OwnerID: r.OwnerID, Name: r.Name, Description: r.Description,
		Permissions: r.Permissions, IsSystem: r.IsSystem,
		CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: r.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- List Roles ---

func ListRolesRoutes() []string { return []string{"/hr/roles"} }

type (
	ListRolesHandler  struct{ Service *application.Service }
	ListRolesRequest  struct{}
	ListRolesResponse struct {
		Roles []*RolePayload `json:"roles"`
	}
)

func (h *ListRolesHandler) Serve(ctx context.Context, _ *ListRolesRequest) (ListRolesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListRolesResponse{}, errorModels.PermissionDeniedServerError()
	}

	roles, err := h.Service.ListRoles(ctx, device.User.UserID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list roles", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListRolesResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list roles"}
	}

	payloads := make([]*RolePayload, 0, len(roles))
	for _, r := range roles {
		payloads = append(payloads, roleToPayload(r))
	}
	return ListRolesResponse{Roles: payloads}, nil
}

// --- Create Role ---

func CreateRoleRoutes() []string { return []string{"/hr/roles"} }

type (
	CreateRoleHandler struct{ Service *application.Service }
	CreateRoleRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Permissions string `json:"permissions"`
	}
)

type CreateRoleResponse struct {
	Role *RolePayload `json:"role"`
}

func (h *CreateRoleHandler) Validate(_ context.Context, req *CreateRoleRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *CreateRoleHandler) Serve(ctx context.Context, req *CreateRoleRequest) (CreateRoleResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateRoleResponse{}, errorModels.PermissionDeniedServerError()
	}

	role := domain.RoleRecord{Name: req.Name, Description: req.Description, Permissions: req.Permissions}
	created, err := h.Service.CreateRole(ctx, device.User.UserID, role)
	if err != nil {
		logger.Ctx(ctx).Error("failed to create role", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateRoleResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create role"}
	}

	return CreateRoleResponse{Role: roleToPayload(*created)}, nil
}

// --- Update Role ---

func UpdateRoleRoutes() []string { return []string{"/hr/roles/{id}"} }

type (
	UpdateRoleHandler struct{ Service *application.Service }
	UpdateRoleRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Permissions string `json:"permissions"`
	}
)

type UpdateRoleResponse struct {
	Role *RolePayload `json:"role"`
}

func (h *UpdateRoleHandler) Validate(_ context.Context, req *UpdateRoleRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *UpdateRoleHandler) Serve(ctx context.Context, req *UpdateRoleRequest) (UpdateRoleResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateRoleResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateRoleResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid role id"}
	}

	role := domain.RoleRecord{ID: id, Name: req.Name, Description: req.Description, Permissions: req.Permissions}
	updated, err := h.Service.UpdateRole(ctx, device.User.UserID, role)
	if err != nil {
		logger.Ctx(ctx).Error("failed to update role", logger.Any("owner_id", device.User.UserID), logger.Any("role_id", id), logger.Error(err))
		return UpdateRoleResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update role"}
	}

	return UpdateRoleResponse{Role: roleToPayload(*updated)}, nil
}

// --- Assign Role ---

func AssignRoleRoutes() []string { return []string{"/hr/employees/{id}/role"} }

type (
	AssignRoleHandler struct{ Service *application.Service }
	AssignRoleRequest struct {
		RoleID int64 `json:"role_id"`
	}
)

type AssignRoleResponse struct {
	Success bool `json:"success"`
}

func (h *AssignRoleHandler) Validate(_ context.Context, req *AssignRoleRequest) *errorModels.ServerError {
	if req.RoleID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "role_id is required"}
	}
	return nil
}

func (h *AssignRoleHandler) Serve(ctx context.Context, req *AssignRoleRequest) (AssignRoleResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return AssignRoleResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	employeeID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return AssignRoleResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid employee id"}
	}

	err := h.Service.AssignRole(ctx, device.User.UserID, employeeID, req.RoleID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to assign role", logger.Any("owner_id", device.User.UserID), logger.Any("employee_id", employeeID), logger.Error(err))
		return AssignRoleResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to assign role"}
	}

	return AssignRoleResponse{Success: true}, nil
}
