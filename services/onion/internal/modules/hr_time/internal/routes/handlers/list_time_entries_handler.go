package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/platform/utils/times"
	hrtimeentryv1 "github.com/voxire/voxie/gen/go/hr/time_entry/v1"
)

func ListTimeEntriesRoutes() []string { return []string{"/hr/time-entries"} }

type ListTimeEntriesHandler struct{ Service *application.Service }

type ListTimeEntriesRequest struct {
	EmployeeID  int64  `json:"employee_id"`
	Date        string `json:"date"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
	Limit       int32  `json:"limit"`
	Offset      int32  `json:"offset"`
}

type ListTimeEntriesResponse struct {
	Entries []*hrtimeentryv1.TimeEntryPayload `json:"entries"`
	Total   int64                             `json:"total"`
	Limit   int32                             `json:"limit"`
	Offset  int32                             `json:"offset"`
}

func (h *ListTimeEntriesHandler) Serve(ctx context.Context, req *ListTimeEntriesRequest) (ListTimeEntriesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListTimeEntriesResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListTimeEntriesResponse{}, onekiterrors.PermissionDenied()
	}

	periodStart := req.PeriodStart
	periodEnd := req.PeriodEnd
	if req.Date != "" && periodStart == "" {
		d, err := time.Parse(times.FormatWithDate, req.Date)
		if err != nil {
			return ListTimeEntriesResponse{}, onekiterrors.New(http.StatusBadRequest, "invalid date format")
		}
		periodStart = d.Format(times.FormatWithDate)
		periodEnd = d.AddDate(0, 0, times.DaysPerWeek-1).Format(times.FormatWithDate)
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	entries, total, err := h.Service.ListTimeEntries(ctx, businessID, domain.ListTimeEntriesParams{
		EmployeeID:  req.EmployeeID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Limit:       req.Limit,
		Offset:      req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list time entries", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListTimeEntriesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list time entries")
	}

	payloads := make([]*hrtimeentryv1.TimeEntryPayload, 0, len(entries))
	for _, e := range entries {
		payloads = append(payloads, timeEntryToProto(e))
	}

	return ListTimeEntriesResponse{Entries: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
