package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
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

func (h *ListRolesHandler) Serve(ctx context.Context, _ *ListRolesRequest) (ListRolesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListRolesResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireHRPermission(ctx, "hr:roles:manage"); err != nil {
		return ListRolesResponse{}, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListRolesResponse{}, onekiterrors.PermissionDenied()
	}

	roles, err := h.Service.ListRoles(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list roles", logger.Any("business_id", businessID), logger.Error(err))
		return ListRolesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list roles")
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

func (h *CreateRoleHandler) Validate(_ context.Context, req *CreateRoleRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	return nil
}

func (h *CreateRoleHandler) Serve(ctx context.Context, req *CreateRoleRequest) (CreateRoleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateRoleResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireHRPermission(ctx, "hr:roles:manage"); err != nil {
		return CreateRoleResponse{}, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateRoleResponse{}, onekiterrors.PermissionDenied()
	}

	role := domain.RoleRecord{Name: req.Name, Description: req.Description, Permissions: req.Permissions}
	created, err := h.Service.CreateRoleForWriter(ctx, businessID, device.User.UserID, role)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return CreateRoleResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to create role", logger.Any("business_id", businessID), logger.Error(err))
		return CreateRoleResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create role")
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

func (h *UpdateRoleHandler) Validate(_ context.Context, req *UpdateRoleRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	return nil
}

func (h *UpdateRoleHandler) Serve(ctx context.Context, req *UpdateRoleRequest) (UpdateRoleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateRoleResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireHRPermission(ctx, "hr:roles:manage"); err != nil {
		return UpdateRoleResponse{}, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateRoleResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateRoleResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid role id")
	}

	role := domain.RoleRecord{ID: id, Name: req.Name, Description: req.Description, Permissions: req.Permissions}
	updated, err := h.Service.UpdateRoleForWriter(ctx, businessID, device.User.UserID, role)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return UpdateRoleResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to update role", logger.Any("business_id", businessID), logger.Any("role_id", id), logger.Error(err))
		return UpdateRoleResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update role")
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

func (h *AssignRoleHandler) Validate(_ context.Context, req *AssignRoleRequest) error {
	if req.RoleID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "role_id is required")
	}
	return nil
}

func (h *AssignRoleHandler) Serve(ctx context.Context, req *AssignRoleRequest) (AssignRoleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return AssignRoleResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireHRPermission(ctx, "hr:roles:manage"); err != nil {
		return AssignRoleResponse{}, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return AssignRoleResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	employeeID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return AssignRoleResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid employee id")
	}

	err := h.Service.AssignRoleForWriter(ctx, businessID, device.User.UserID, employeeID, req.RoleID)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return AssignRoleResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to assign role", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return AssignRoleResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to assign role")
	}

	return AssignRoleResponse{Success: true}, nil
}

// --- List Permission Catalog ---

func ListPermissionCatalogRoutes() []string { return []string{"/hr/permissions"} }

type (
	ListPermissionCatalogHandler  struct{ Service *application.Service }
	ListPermissionCatalogRequest  struct{}
	ListPermissionCatalogResponse struct {
		Permissions []*PermissionCatalogPayload `json:"permissions"`
	}
)

type PermissionCatalogPayload struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

func (h *ListPermissionCatalogHandler) Serve(ctx context.Context, _ *ListPermissionCatalogRequest) (ListPermissionCatalogResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListPermissionCatalogResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireHRPermission(ctx, "hr:roles:manage"); err != nil {
		return ListPermissionCatalogResponse{}, err
	}

	catalog := h.Service.ListPermissionCatalog()
	payloads := make([]*PermissionCatalogPayload, 0, len(catalog))
	for _, p := range catalog {
		payloads = append(payloads, &PermissionCatalogPayload{
			Key: p.Key, Label: p.Label, Description: p.Description, Category: p.Category,
		})
	}
	return ListPermissionCatalogResponse{Permissions: payloads}, nil
}
