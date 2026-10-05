package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrtimeentryv1 "github.com/voxire/voxie/gen/go/hr/time_entry/v1"
)

func UpdateTimeEntryRoutes() []string { return []string{"/hr/time-entries/{id}"} }

type UpdateTimeEntryHandler struct{ Service *application.Service }

type UpdateTimeEntryRequest struct {
	EmployeeID      *int64  `json:"employee_id"`
	ProjectID       *int64  `json:"project_id"`
	TaskID          *int64  `json:"task_id"`
	CategoryID      *int64  `json:"category_id"`
	Date            *string `json:"date"`
	DurationMinutes *int32  `json:"duration_minutes"`
	StartTime       *string `json:"start_time"`
	EndTime         *string `json:"end_time"`
	Description     *string `json:"description"`
	IsBillable      *bool   `json:"is_billable"`
	Source          *string `json:"source"`
}

type UpdateTimeEntryResponse struct {
	Entry *hrtimeentryv1.TimeEntryPayload `json:"entry"`
}

func (h *UpdateTimeEntryHandler) Validate(_ context.Context, req *UpdateTimeEntryRequest) error {
	if req == nil {
		return onekiterrors.New(http.StatusBadRequest, "request body is required")
	}
	if req.EmployeeID != nil && *req.EmployeeID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id must be greater than 0")
	}
	if req.DurationMinutes != nil && *req.DurationMinutes <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "duration_minutes must be greater than 0")
	}
	if req.Date != nil {
		if _, err := time.Parse(time.DateOnly, *req.Date); err != nil {
			return onekiterrors.New(http.StatusBadRequest, "date must be in YYYY-MM-DD format")
		}
	}
	return nil
}

func (h *UpdateTimeEntryHandler) Serve(ctx context.Context, req *UpdateTimeEntryRequest) (UpdateTimeEntryResponse, error) {
	if err := h.Validate(ctx, req); err != nil {
		return UpdateTimeEntryResponse{}, err
	}
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateTimeEntryResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateTimeEntryResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateTimeEntryResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid time entry id")
	}

	patch := domain.TimeEntryPatch{
		EmployeeID:      req.EmployeeID,
		ProjectID:       req.ProjectID,
		TaskID:          req.TaskID,
		CategoryID:      req.CategoryID,
		Date:            req.Date,
		DurationMinutes: req.DurationMinutes,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		Description:     req.Description,
		IsBillable:      req.IsBillable,
		Source:          req.Source,
	}
	updated, err := h.Service.UpdateTimeEntry(ctx, businessID, device.User.UserID, id, patch)
	if err != nil {
		if errors.Is(err, domain.ErrTimeEntryAccessDenied) {
			return UpdateTimeEntryResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrInvalidDuration) {
			return UpdateTimeEntryResponse{}, onekiterrors.New(http.StatusBadRequest, "Duration must be greater than 0")
		}
		if errors.Is(err, domain.ErrDateInLockedPeriod) {
			return UpdateTimeEntryResponse{}, onekiterrors.New(http.StatusConflict, "Cannot add entry to a locked period")
		}
		if errors.Is(err, domain.ErrTimeEntryNotFound) {
			return UpdateTimeEntryResponse{}, onekiterrors.New(http.StatusNotFound, "Time entry not found")
		}
		logger.Ctx(ctx).Error("failed to update time entry", logger.Any("owner_id", device.User.UserID), logger.Any("entry_id", id), logger.Error(err))
		return UpdateTimeEntryResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update time entry")
	}

	return UpdateTimeEntryResponse{Entry: timeEntryToProto(*updated)}, nil
}
