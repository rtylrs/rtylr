package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func UpdateTimeEntryRoutes() []string { return []string{"/hr/time-entries/{id}"} }

type UpdateTimeEntryHandler struct{ Service *application.Service }

type UpdateTimeEntryRequest struct {
	EmployeeID      int64  `json:"employee_id"`
	ProjectID       int64  `json:"project_id"`
	TaskID          int64  `json:"task_id"`
	CategoryID      int64  `json:"category_id"`
	Date            string `json:"date"`
	DurationMinutes int32  `json:"duration_minutes"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	Description     string `json:"description"`
	IsBillable      bool   `json:"is_billable"`
	Source          string `json:"source"`
}

type UpdateTimeEntryResponse struct {
	Entry *hrv1.TimeEntryPayload `json:"entry"`
}

func (h *UpdateTimeEntryHandler) Validate(_ context.Context, req *UpdateTimeEntryRequest) *errorModels.ServerError {
	if req.DurationMinutes <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "duration_minutes must be greater than 0"}
	}
	if req.Date == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "date is required"}
	}
	return nil
}

func (h *UpdateTimeEntryHandler) Serve(ctx context.Context, req *UpdateTimeEntryRequest) (UpdateTimeEntryResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateTimeEntryResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return UpdateTimeEntryResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid time entry id"}
	}

	entry := domain.TimeEntryRecord{
		ID:              id,
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

	updated, err := h.Service.UpdateTimeEntry(ctx, device.User.UserID, entry)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidDuration) {
			return UpdateTimeEntryResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Duration must be greater than 0"}
		}
		logger.Ctx(ctx).Error("failed to update time entry", logger.Any("owner_id", device.User.UserID), logger.Any("entry_id", id), logger.Error(err))
		return UpdateTimeEntryResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update time entry"}
	}

	return UpdateTimeEntryResponse{Entry: timeEntryToProto(*updated)}, nil
}
