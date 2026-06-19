package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/timesheets/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type TimesheetEntryPayload struct {
	StaffID       int64   `json:"staff_id"`
	StaffName     string  `json:"staff_name"`
	WeekStart     string  `json:"week_start"`
	TotalHours    float64 `json:"total_hours"`
	RegularHours  float64 `json:"regular_hours"`
	OvertimeHours float64 `json:"overtime_hours"`
}

func ListRoutes() []string { return []string{"/hr/timesheets"} }

type (
	ListHandler struct{ Service *application.Service }
	ListRequest struct {
		WeekStart string `json:"week_start"`
	}
)

type ListResponse struct {
	Timesheets []TimesheetEntryPayload `json:"timesheets"`
}

func (h *ListHandler) Validate(_ context.Context, req *ListRequest) *errorModels.ServerError {
	if req.WeekStart != "" {
		if _, err := time.Parse("2006-01-02", req.WeekStart); err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "week_start must be a date in YYYY-MM-DD format"}
		}
	}
	return nil
}

func (h *ListHandler) Serve(ctx context.Context, req *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}

	entries, err := h.Service.GetTimesheets(ctx, businessID, req.WeekStart)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get timesheets"}
	}

	payloads := make([]TimesheetEntryPayload, 0, len(entries))
	for _, e := range entries {
		payloads = append(payloads, TimesheetEntryPayload{
			StaffID: e.StaffID, StaffName: e.StaffName, WeekStart: e.WeekStart,
			TotalHours: e.TotalHours, RegularHours: e.RegularHours, OvertimeHours: e.OvertimeHours,
		})
	}

	return ListResponse{Timesheets: payloads}, nil
}
