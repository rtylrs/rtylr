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
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
)

func deptToProto(d domain.DepartmentRecord) *hrv1.DepartmentPayload {
	return &hrv1.DepartmentPayload{
		Id:                    d.ID,
		OwnerId:               d.OwnerID,
		LocationId:            d.LocationID,
		ParentDepartmentId:    &d.ParentDepartmentID,
		Name:                  d.Name,
		DefaultWorkScheduleId: &d.DefaultWorkScheduleID,
		CreatedAt:             d.CreatedAt.Unix(),
		UpdatedAt:             d.UpdatedAt.Unix(),
	}
}

// --- List Departments ---

func ListHRCoreDepartmentsRoutes() []string { return []string{"/hr/core/departments"} }

type (
	ListHRCoreDepartmentsHandler struct{ Service *application.Service }
	ListHRCoreDepartmentsRequest struct {
		LocationID int64 `json:"location_id"`
	}
)

type ListHRCoreDepartmentsResponse struct {
	Departments []*hrv1.DepartmentPayload `json:"departments"`
}

func (h *ListHRCoreDepartmentsHandler) Serve(ctx context.Context, req *ListHRCoreDepartmentsRequest) (ListHRCoreDepartmentsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListHRCoreDepartmentsResponse{}, errorModels.PermissionDeniedServerError()
	}

	departments, err := h.Service.ListDepartments(ctx, device.User.UserID, req.LocationID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list departments", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListHRCoreDepartmentsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list departments"}
	}

	payloads := make([]*hrv1.DepartmentPayload, 0, len(departments))
	for _, d := range departments {
		payloads = append(payloads, deptToProto(d))
	}
	return ListHRCoreDepartmentsResponse{Departments: payloads}, nil
}

// --- Create Department ---

func CreateHRCoreDepartmentRoutes() []string { return []string{"/hr/core/departments"} }

type (
	CreateHRCoreDepartmentHandler struct{ Service *application.Service }
	CreateHRCoreDepartmentRequest struct {
		Name                  string `json:"name"`
		LocationID            int64  `json:"location_id"`
		ParentDepartmentID    int64  `json:"parent_department_id"`
		DefaultWorkScheduleID int64  `json:"default_work_schedule_id"`
	}
)

type CreateHRCoreDepartmentResponse struct {
	Department *hrv1.DepartmentPayload `json:"department"`
}

func (h *CreateHRCoreDepartmentHandler) Validate(_ context.Context, req *CreateHRCoreDepartmentRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *CreateHRCoreDepartmentHandler) Serve(ctx context.Context, req *CreateHRCoreDepartmentRequest) (CreateHRCoreDepartmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateHRCoreDepartmentResponse{}, errorModels.PermissionDeniedServerError()
	}

	dept := domain.DepartmentRecord{Name: req.Name, LocationID: req.LocationID, ParentDepartmentID: req.ParentDepartmentID, DefaultWorkScheduleID: req.DefaultWorkScheduleID}
	created, err := h.Service.CreateDepartment(ctx, device.User.UserID, dept)
	if err != nil {
		logger.Ctx(ctx).Error("failed to create department", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateHRCoreDepartmentResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create department"}
	}

	return CreateHRCoreDepartmentResponse{Department: deptToProto(*created)}, nil
}

// --- Update Department ---

func UpdateHRCoreDepartmentRoutes() []string { return []string{"/hr/core/departments/{id}"} }

type (
	UpdateHRCoreDepartmentHandler struct{ Service *application.Service }
	UpdateHRCoreDepartmentRequest struct {
		Name                  string `json:"name"`
		LocationID            int64  `json:"location_id"`
		ParentDepartmentID    int64  `json:"parent_department_id"`
		DefaultWorkScheduleID int64  `json:"default_work_schedule_id"`
	}
)

type UpdateHRCoreDepartmentResponse struct {
	Department *hrv1.DepartmentPayload `json:"department"`
}

func (h *UpdateHRCoreDepartmentHandler) Validate(_ context.Context, req *UpdateHRCoreDepartmentRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *UpdateHRCoreDepartmentHandler) Serve(ctx context.Context, req *UpdateHRCoreDepartmentRequest) (UpdateHRCoreDepartmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateHRCoreDepartmentResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateHRCoreDepartmentResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid department id"}
	}

	dept := domain.DepartmentRecord{ID: id, Name: req.Name, LocationID: req.LocationID, ParentDepartmentID: req.ParentDepartmentID, DefaultWorkScheduleID: req.DefaultWorkScheduleID}
	updated, err := h.Service.UpdateDepartment(ctx, device.User.UserID, dept)
	if err != nil {
		logger.Ctx(ctx).Error("failed to update department", logger.Any("owner_id", device.User.UserID), logger.Any("department_id", id), logger.Error(err))
		return UpdateHRCoreDepartmentResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update department"}
	}

	return UpdateHRCoreDepartmentResponse{Department: deptToProto(*updated)}, nil
}
