package handlers

import (
	"context"
	"net/http"

	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"

	"github.com/voxire/onion/internal/modules/departments/internal/application"
	"github.com/voxire/onion/internal/modules/departments/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func ListDepartmentsRoutes() []string   { return []string{"/hr/departments"} }
func CreateDepartmentRoutes() []string  { return []string{"/hr/departments"} }
func UpdateDepartmentRoutes() []string  { return []string{"/hr/departments"} }
func DeleteDepartmentRoutes() []string  { return []string{"/hr/departments"} }

func newDepartmentPayload(d domain.Department) *hrv1.Department {
	return &hrv1.Department{
		Id: d.ID, OwnerId: d.OwnerID, Name: d.Name, Description: d.Description,
		CreatedAt: d.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: d.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type ListDepartmentsHandler struct{ Service *application.Service }
type ListDepartmentsRequest struct{}
type ListDepartmentsResponse struct {
	Departments []*hrv1.Department `json:"departments"`
}

func (h *ListDepartmentsHandler) Serve(ctx context.Context, _ *ListDepartmentsRequest) (ListDepartmentsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListDepartmentsResponse{}, errorModels.PermissionDeniedServerError()
	}
	departments, err := h.Service.ListDepartments(ctx, device.User.UserID)
	if err != nil {
		return ListDepartmentsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list departments"}
	}
	payloads := make([]*hrv1.Department, 0, len(departments))
	for _, d := range departments {
		payloads = append(payloads, newDepartmentPayload(d))
	}
	return ListDepartmentsResponse{Departments: payloads}, nil
}

type CreateDepartmentHandler struct{ Service *application.Service }
type CreateDepartmentRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
type CreateDepartmentResponse struct {
	ID int64 `json:"id"`
}

func (h *CreateDepartmentHandler) Validate(_ context.Context, req *CreateDepartmentRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *CreateDepartmentHandler) Serve(ctx context.Context, req *CreateDepartmentRequest) (CreateDepartmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateDepartmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.DepartmentInput{OwnerID: device.User.UserID, Name: req.Name, Description: req.Description}
	id, err := h.Service.CreateDepartment(ctx, input)
	if err != nil {
		return CreateDepartmentResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create department"}
	}
	return CreateDepartmentResponse{ID: id}, nil
}

type UpdateDepartmentHandler struct{ Service *application.Service }
type UpdateDepartmentRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
type UpdateDepartmentResponse struct {
	Success bool `json:"success"`
}

func (h *UpdateDepartmentHandler) Validate(_ context.Context, req *UpdateDepartmentRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *UpdateDepartmentHandler) Serve(ctx context.Context, req *UpdateDepartmentRequest) (UpdateDepartmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateDepartmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.DepartmentInput{ID: req.ID, Name: req.Name, Description: req.Description}
	if err := h.Service.UpdateDepartment(ctx, device.User.UserID, input); err != nil {
		return UpdateDepartmentResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update department"}
	}
	return UpdateDepartmentResponse{Success: true}, nil
}

type DeleteDepartmentHandler struct{ Service *application.Service }
type DeleteDepartmentRequest struct {
	ID int64 `json:"id"`
}
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
	if err := h.Service.DeleteDepartment(ctx, device.User.UserID, req.ID); err != nil {
		return DeleteDepartmentResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete department"}
	}
	return DeleteDepartmentResponse{Success: true}, nil
}
