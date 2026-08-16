package handlers

import (
	"context"
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

type WorkSchedulePayload struct {
	ID           int64  `json:"id"`
	OwnerID      int64  `json:"owner_id"`
	Name         string `json:"name"`
	ScheduleData string `json:"schedule_data"`
	IsDefault    bool   `json:"is_default"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func wsToPayload(ws domain.WorkScheduleRecord) *WorkSchedulePayload {
	return &WorkSchedulePayload{
		ID: ws.ID, OwnerID: ws.OwnerID, Name: ws.Name,
		ScheduleData: ws.ScheduleData, IsDefault: ws.IsDefault,
		CreatedAt: ws.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: ws.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- List Work Schedules ---

func ListWorkSchedulesRoutes() []string { return []string{"/hr/work-schedules"} }

type (
	ListWorkSchedulesHandler  struct{ Service *application.Service }
	ListWorkSchedulesRequest  struct{}
	ListWorkSchedulesResponse struct {
		WorkSchedules []*WorkSchedulePayload `json:"work_schedules"`
	}
)

func (h *ListWorkSchedulesHandler) Serve(ctx context.Context, _ *ListWorkSchedulesRequest) (ListWorkSchedulesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListWorkSchedulesResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListWorkSchedulesResponse{}, onekiterrors.PermissionDenied()
	}

	schedules, err := h.Service.ListWorkSchedules(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list work schedules", logger.Any("business_id", businessID), logger.Error(err))
		return ListWorkSchedulesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list work schedules")
	}

	payloads := make([]*WorkSchedulePayload, 0, len(schedules))
	for _, ws := range schedules {
		payloads = append(payloads, wsToPayload(ws))
	}
	return ListWorkSchedulesResponse{WorkSchedules: payloads}, nil
}

// --- Create Work Schedule ---

func CreateWorkScheduleRoutes() []string { return []string{"/hr/work-schedules"} }

type (
	CreateWorkScheduleHandler struct{ Service *application.Service }
	CreateWorkScheduleRequest struct {
		Name         string `json:"name"`
		ScheduleData string `json:"schedule_data"`
		IsDefault    bool   `json:"is_default"`
	}
)

type CreateWorkScheduleResponse struct {
	WorkSchedule *WorkSchedulePayload `json:"work_schedule"`
}

func (h *CreateWorkScheduleHandler) Validate(_ context.Context, req *CreateWorkScheduleRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	return nil
}

func (h *CreateWorkScheduleHandler) Serve(ctx context.Context, req *CreateWorkScheduleRequest) (CreateWorkScheduleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateWorkScheduleResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateWorkScheduleResponse{}, onekiterrors.PermissionDenied()
	}

	ws := domain.WorkScheduleRecord{Name: req.Name, ScheduleData: req.ScheduleData, IsDefault: req.IsDefault}
	created, err := h.Service.CreateWorkSchedule(ctx, businessID, ws)
	if err != nil {
		logger.Ctx(ctx).Error("failed to create work schedule", logger.Any("business_id", businessID), logger.Error(err))
		return CreateWorkScheduleResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create work schedule")
	}

	return CreateWorkScheduleResponse{WorkSchedule: wsToPayload(*created)}, nil
}

// --- Update Work Schedule ---

func UpdateWorkScheduleRoutes() []string { return []string{"/hr/work-schedules/{id}"} }

type (
	UpdateWorkScheduleHandler struct{ Service *application.Service }
	UpdateWorkScheduleRequest struct {
		Name         string `json:"name"`
		ScheduleData string `json:"schedule_data"`
		IsDefault    bool   `json:"is_default"`
	}
)

type UpdateWorkScheduleResponse struct {
	WorkSchedule *WorkSchedulePayload `json:"work_schedule"`
}

func (h *UpdateWorkScheduleHandler) Validate(_ context.Context, req *UpdateWorkScheduleRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	return nil
}

func (h *UpdateWorkScheduleHandler) Serve(ctx context.Context, req *UpdateWorkScheduleRequest) (UpdateWorkScheduleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateWorkScheduleResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateWorkScheduleResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateWorkScheduleResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid work schedule id")
	}

	ws := domain.WorkScheduleRecord{ID: id, Name: req.Name, ScheduleData: req.ScheduleData, IsDefault: req.IsDefault}
	updated, err := h.Service.UpdateWorkSchedule(ctx, businessID, ws)
	if err != nil {
		logger.Ctx(ctx).Error("failed to update work schedule", logger.Any("business_id", businessID), logger.Any("schedule_id", id), logger.Error(err))
		return UpdateWorkScheduleResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update work schedule")
	}

	return UpdateWorkScheduleResponse{WorkSchedule: wsToPayload(*updated)}, nil
}
