package api

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr/timesheets/internal/application"
	"github.com/voxire/onion/internal/modules/hr/timesheets/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrtimesheetv1 "github.com/voxire/voxie/gen/go/hr/timesheet/v1"
)

type TimesheetAPI struct {
	service *application.Service
}

func NewTimesheetAPI(service *application.Service) *TimesheetAPI {
	return &TimesheetAPI{service: service}
}

func (a *TimesheetAPI) ListTimesheets(ctx context.Context, req *hrtimesheetv1.ListTimesheetsRequest) (*hrtimesheetv1.ListTimesheetsResponse, error) {
	if req.GetWeekStart() != "" {
		if _, err := time.Parse("2006-01-02", req.GetWeekStart()); err != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "week_start must be a date in YYYY-MM-DD format")
		}
	}

	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	entries, err := a.service.GetTimesheets(ctx, businessID, req.GetWeekStart())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get timesheets")
	}

	return &hrtimesheetv1.ListTimesheetsResponse{Timesheets: newTimesheetEntries(entries)}, nil
}

func newTimesheetEntries(entries []domain.TimesheetEntry) []*hrtimesheetv1.TimesheetEntry {
	out := make([]*hrtimesheetv1.TimesheetEntry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, &hrtimesheetv1.TimesheetEntry{
			StaffId:       entry.StaffID,
			StaffName:     entry.StaffName,
			WeekStart:     entry.WeekStart,
			TotalHours:    entry.TotalHours,
			RegularHours:  entry.RegularHours,
			OvertimeHours: entry.OvertimeHours,
		})
	}
	return out
}
