package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/platform/logger"
	hrtimeentryv1 "github.com/voxire/voxie/gen/go/hr/time_entry/v1"
)

func CreateTimeEntryRoutes() []string { return []string{"/hr/time-entries"} }

type CreateTimeEntryHandler struct{ Service *application.Service }

type CreateTimeEntryRequest struct {
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

type CreateTimeEntryResponse struct {
	Entry *hrtimeentryv1.TimeEntryPayload `json:"entry"`
}

func (h *CreateTimeEntryHandler) Validate(_ context.Context, req *CreateTimeEntryRequest) *errorModels.ServerError {
	if req.EmployeeID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "employee_id is required"}
	}
	if req.DurationMinutes <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "duration_minutes must be greater than 0"}
	}
	if req.Date == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "date is required"}
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "date must be in YYYY-MM-DD format"}
	}
	return nil
}

func (h *CreateTimeEntryHandler) Serve(ctx context.Context, req *CreateTimeEntryRequest) (CreateTimeEntryResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateTimeEntryResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateTimeEntryResponse{}, errorModels.PermissionDeniedServerError()
	}

	entry := domain.TimeEntryRecord{
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

	created, err := h.Service.CreateTimeEntry(ctx, businessID, entry)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidDuration) {
			return CreateTimeEntryResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Duration must be greater than 0"}
		}
		if errors.Is(err, domain.ErrDateInLockedPeriod) {
			return CreateTimeEntryResponse{}, &errorModels.ServerError{HTTPCode: http.StatusConflict, Code: http.StatusConflict, Message: "Cannot add entry to a locked period"}
		}
		logger.Ctx(ctx).Error("failed to create time entry", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateTimeEntryResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create time entry"}
	}

	return CreateTimeEntryResponse{Entry: timeEntryToProto(*created)}, nil
}
