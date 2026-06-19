package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/departments/internal/application"
	"github.com/voxire/onion/internal/modules/departments/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrdepartmentv1 "github.com/voxire/voxie/gen/go/hr/department/v1"
)

func ListDepartmentsRoutes() []string  { return []string{"/hr/departments"} }
func CreateDepartmentRoutes() []string { return []string{"/hr/departments"} }
func UpdateDepartmentRoutes() []string { return []string{"/hr/departments"} }
func DeleteDepartmentRoutes() []string { return []string{"/hr/departments"} }

func newDepartmentPayload(d domain.Department) *hrdepartmentv1.DepartmentPayload {
	return &hrdepartmentv1.DepartmentPayload{
		Id:        d.ID,
		OwnerId:   d.OwnerID,
		Name:      d.Name,
		CreatedAt: d.CreatedAt.Unix(),
		UpdatedAt: d.UpdatedAt.Unix(),
	}
}

type (
	ListDepartmentsHandler  struct{ Service *application.Service }
	ListDepartmentsRequest  struct{}
	ListDepartmentsResponse struct {
		Departments []*hrdepartmentv1.DepartmentPayload `json:"departments"`
	}
)

func (h *ListDepartmentsHandler) Serve(ctx context.Context, _ *ListDepartmentsRequest) (ListDepartmentsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListDepartmentsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListDepartmentsResponse{}, errorModels.PermissionDeniedServerError()
	}
	departments, err := h.Service.ListDepartments(ctx, businessID)
	if err != nil {
		return ListDepartmentsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list departments"}
	}
	payloads := make([]*hrdepartmentv1.DepartmentPayload, 0, len(departments))
	for _, d := range departments {
		payloads = append(payloads, newDepartmentPayload(d))
	}
	return ListDepartmentsResponse{Departments: payloads}, nil
}

type (
	CreateDepartmentHandler struct{ Service *application.Service }
	CreateDepartmentRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
)

type CreateDepartmentResponse struct {
	ID int64 `json:"id"`
}

func (h *CreateDepartmentHandler) Validate(_ context.Context, req *CreateDepartmentRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *CreateDepartmentHandler) Serve(ctx context.Context, req *CreateDepartmentRequest) (CreateDepartmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateDepartmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateDepartmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.DepartmentInput{OwnerID: device.User.UserID, BusinessID: businessID, Name: req.Name, Description: req.Description}
	id, err := h.Service.CreateDepartment(ctx, input)
	if err != nil {
		return CreateDepartmentResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create department"}
	}
	return CreateDepartmentResponse{ID: id}, nil
}

type (
	UpdateDepartmentHandler struct{ Service *application.Service }
	UpdateDepartmentRequest struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
)

type UpdateDepartmentResponse struct {
	Success bool `json:"success"`
}

func (h *UpdateDepartmentHandler) Validate(_ context.Context, req *UpdateDepartmentRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *UpdateDepartmentHandler) Serve(ctx context.Context, req *UpdateDepartmentRequest) (UpdateDepartmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateDepartmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateDepartmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.DepartmentInput{ID: req.ID, Name: req.Name, Description: req.Description}
	if err := h.Service.UpdateDepartment(ctx, businessID, input); err != nil {
		return UpdateDepartmentResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update department"}
	}
	return UpdateDepartmentResponse{Success: true}, nil
}

type (
	DeleteDepartmentHandler struct{ Service *application.Service }
	DeleteDepartmentRequest struct {
		ID int64 `json:"id"`
	}
)

type DeleteDepartmentResponse struct {
	Success bool `json:"success"`
}

func (h *DeleteDepartmentHandler) Validate(_ context.Context, req *DeleteDepartmentRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	return nil
}

func (h *DeleteDepartmentHandler) Serve(ctx context.Context, req *DeleteDepartmentRequest) (DeleteDepartmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteDepartmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteDepartmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.DeleteDepartment(ctx, businessID, req.ID); err != nil {
		return DeleteDepartmentResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete department"}
	}
	return DeleteDepartmentResponse{Success: true}, nil
}
