package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
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

func (r *CreateTimeEntryRequest) UnmarshalJSON(data []byte) error {
	type alias CreateTimeEntryRequest
	aux := struct {
		*alias
		EmployeeID json.RawMessage `json:"employee_id"`
		ProjectID  json.RawMessage `json:"project_id"`
		TaskID     json.RawMessage `json:"task_id"`
		CategoryID json.RawMessage `json:"category_id"`
	}{alias: (*alias)(r)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	ids := []struct {
		name string
		raw  json.RawMessage
		dst  *int64
	}{
		{name: "employee_id", raw: aux.EmployeeID, dst: &r.EmployeeID},
		{name: "project_id", raw: aux.ProjectID, dst: &r.ProjectID},
		{name: "task_id", raw: aux.TaskID, dst: &r.TaskID},
		{name: "category_id", raw: aux.CategoryID, dst: &r.CategoryID},
	}
	for _, id := range ids {
		if len(id.raw) == 0 || string(id.raw) == "null" {
			continue
		}
		value, err := parseJSONInt64(id.raw)
		if err != nil {
			return fmt.Errorf("%s: %w", id.name, err)
		}
		*id.dst = value
	}
	return nil
}

func parseJSONInt64(raw json.RawMessage) (int64, error) {
	var stringValue string
	if err := json.Unmarshal(raw, &stringValue); err == nil {
		return strconv.ParseInt(strings.TrimSpace(stringValue), 10, 64)
	}

	var value int64
	if err := json.Unmarshal(raw, &value); err != nil {
		return 0, err
	}
	return value, nil
}

type CreateTimeEntryResponse struct {
	Entry *hrtimeentryv1.TimeEntryPayload `json:"entry"`
}

func (h *CreateTimeEntryHandler) Validate(_ context.Context, req *CreateTimeEntryRequest) error {
	if req == nil {
		return onekiterrors.New(http.StatusBadRequest, "request body is required")
	}
	if req.EmployeeID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id must be greater than 0")
	}
	if req.DurationMinutes <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "duration_minutes must be greater than 0")
	}
	if req.Date == "" {
		return onekiterrors.New(http.StatusBadRequest, "date is required")
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "date must be in YYYY-MM-DD format")
	}
	return nil
}

func (h *CreateTimeEntryHandler) Serve(ctx context.Context, req *CreateTimeEntryRequest) (CreateTimeEntryResponse, error) {
	if req == nil {
		return CreateTimeEntryResponse{}, onekiterrors.New(http.StatusBadRequest, "request body is required")
	}
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateTimeEntryResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateTimeEntryResponse{}, onekiterrors.PermissionDenied()
	}
	if req.EmployeeID <= 0 {
		return CreateTimeEntryResponse{}, onekiterrors.New(http.StatusBadRequest, "employee_id must be greater than 0")
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

	created, err := h.Service.CreateTimeEntry(ctx, businessID, device.User.UserID, entry)
	if err != nil {
		if errors.Is(err, domain.ErrTimeEntryAccessDenied) {
			return CreateTimeEntryResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrInvalidDuration) {
			return CreateTimeEntryResponse{}, onekiterrors.New(http.StatusBadRequest, "Duration must be greater than 0")
		}
		if errors.Is(err, domain.ErrDateInLockedPeriod) {
			return CreateTimeEntryResponse{}, onekiterrors.New(http.StatusConflict, "Cannot add entry to a locked period")
		}
		logger.Ctx(ctx).Error("failed to create time entry", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateTimeEntryResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create time entry")
	}

	return CreateTimeEntryResponse{Entry: timeEntryToProto(*created)}, nil
}
