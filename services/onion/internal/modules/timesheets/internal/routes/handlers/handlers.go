package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/timesheets/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type TimesheetEntryPayload struct {
	StaffID       int64   `json:"staffId"`
	StaffName     string  `json:"staffName"`
	WeekStart     string  `json:"weekStart"`
	TotalHours    float64 `json:"totalHours"`
	RegularHours  float64 `json:"regularHours"`
	OvertimeHours float64 `json:"overtimeHours"`
}

func ListRoutes() []string { return []string{"/hr/timesheets"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct {
	WeekStart string `json:"weekStart"`
}
type ListResponse struct {
	Timesheets []TimesheetEntryPayload `json:"timesheets"`
}

func (h *ListHandler) Serve(ctx context.Context, req *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}

	entries, err := h.Service.GetTimesheets(ctx, device.User.UserID, req.WeekStart)
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
