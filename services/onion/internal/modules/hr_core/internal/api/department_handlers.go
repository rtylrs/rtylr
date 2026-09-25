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
	hrdepartmentv1 "github.com/voxire/voxie/gen/go/hr/department/v1"
)

func deptToProto(d domain.DepartmentRecord) *hrdepartmentv1.DepartmentPayload {
	return &hrdepartmentv1.DepartmentPayload{
		Id:        d.ID,
		OwnerId:   d.OwnerID,
		Name:      d.Name,
		CreatedAt: d.CreatedAt.Unix(),
		UpdatedAt: d.UpdatedAt.Unix(),
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
	Departments []*hrdepartmentv1.DepartmentPayload `json:"departments"`
}

func (h *ListHRCoreDepartmentsHandler) Serve(ctx context.Context, req *ListHRCoreDepartmentsRequest) (ListHRCoreDepartmentsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListHRCoreDepartmentsResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListHRCoreDepartmentsResponse{}, onekiterrors.PermissionDenied()
	}

	departments, err := h.Service.ListDepartments(ctx, businessID, req.LocationID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list departments", logger.Any("business_id", businessID), logger.Error(err))
		return ListHRCoreDepartmentsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list departments")
	}

	payloads := make([]*hrdepartmentv1.DepartmentPayload, 0, len(departments))
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
	Department *hrdepartmentv1.DepartmentPayload `json:"department"`
}

func (h *CreateHRCoreDepartmentHandler) Validate(_ context.Context, req *CreateHRCoreDepartmentRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	return nil
}

func (h *CreateHRCoreDepartmentHandler) Serve(ctx context.Context, req *CreateHRCoreDepartmentRequest) (CreateHRCoreDepartmentResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateHRCoreDepartmentResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateHRCoreDepartmentResponse{}, onekiterrors.PermissionDenied()
	}

	dept := domain.DepartmentRecord{Name: req.Name, LocationID: req.LocationID, ParentDepartmentID: req.ParentDepartmentID, DefaultWorkScheduleID: req.DefaultWorkScheduleID}
	created, err := h.Service.CreateDepartment(ctx, businessID, dept)
	if err != nil {
		logger.Ctx(ctx).Error("failed to create department", logger.Any("business_id", businessID), logger.Error(err))
		return CreateHRCoreDepartmentResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create department")
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
	Department *hrdepartmentv1.DepartmentPayload `json:"department"`
}

func (h *UpdateHRCoreDepartmentHandler) Validate(_ context.Context, req *UpdateHRCoreDepartmentRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	return nil
}

func (h *UpdateHRCoreDepartmentHandler) Serve(ctx context.Context, req *UpdateHRCoreDepartmentRequest) (UpdateHRCoreDepartmentResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateHRCoreDepartmentResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateHRCoreDepartmentResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateHRCoreDepartmentResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid department id")
	}

	dept := domain.DepartmentRecord{ID: id, Name: req.Name, LocationID: req.LocationID, ParentDepartmentID: req.ParentDepartmentID, DefaultWorkScheduleID: req.DefaultWorkScheduleID}
	updated, err := h.Service.UpdateDepartment(ctx, businessID, dept)
	if err != nil {
		if errors.Is(err, domain.ErrDepartmentNotFound) {
			return UpdateHRCoreDepartmentResponse{}, onekiterrors.New(http.StatusNotFound, "Department not found")
		}
		logger.Ctx(ctx).Error("failed to update department", logger.Any("business_id", businessID), logger.Any("department_id", id), logger.Error(err))
		return UpdateHRCoreDepartmentResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update department")
	}

	return UpdateHRCoreDepartmentResponse{Department: deptToProto(*updated)}, nil
}
