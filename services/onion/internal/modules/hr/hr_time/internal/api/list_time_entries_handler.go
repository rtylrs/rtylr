package api

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/platform/utils/times"
	hrtimeentryv1 "github.com/voxire/voxie/gen/go/hr/time_entry/v1"
)

func ListTimeEntriesRoutes() []string { return []string{"/hr/time-entries"} }

type ListTimeEntriesHandler struct{ Service *application.Service }

// Field names and semantics follow hr.time_entry.v1 ListTimeEntriesRequest.
// This route is hand-registered rather than served by the generated service, so
// the contract and this struct agree only because they are kept in step by hand
// (tryrtylr/rtylr#873 is the general form of that).
//
// Optional filters are pointers so "absent" is distinguishable from a zero the
// caller meant. is_billable is the clearest case: false is a real billability,
// not a missing filter.
type ListTimeEntriesRequest struct {
	EmployeeID *int64  `json:"employee_id"`
	Date       string  `json:"date"`
	FromDate   *string `json:"from_date"`
	ToDate     *string `json:"to_date"`
	ProjectID  *int64  `json:"project_id"`
	IsBillable *bool   `json:"is_billable"`
	Status     *string `json:"status"`
	Limit      int32   `json:"limit"`
	Offset     int32   `json:"offset"`

	// period_start and period_end predate the contract's from_date/to_date and
	// are accepted as aliases so any caller still sending them keeps working.
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
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

	// Range precedence, as the contract's doc comment states it: from_date and
	// to_date win, then the period_* aliases, then date. date expands to the week
	// beginning on it, which is what this endpoint has always done and what
	// apps/hr's timesheet view relies on.
	fromDate, toDate := req.FromDate, req.ToDate
	if fromDate == nil && toDate == nil {
		switch {
		case req.PeriodStart != "" || req.PeriodEnd != "":
			if req.PeriodStart != "" {
				fromDate = &req.PeriodStart
			}
			if req.PeriodEnd != "" {
				toDate = &req.PeriodEnd
			}
		case req.Date != "":
			d, err := time.Parse(times.FormatWithDate, req.Date)
			if err != nil {
				return ListTimeEntriesResponse{}, onekiterrors.New(http.StatusBadRequest, "invalid date format")
			}
			start := d.Format(times.FormatWithDate)
			end := d.AddDate(0, 0, times.DaysPerWeek-1).Format(times.FormatWithDate)
			fromDate, toDate = &start, &end
		}
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	entries, total, err := h.Service.ListTimeEntries(ctx, businessID, domain.ListTimeEntriesParams{
		EmployeeID: req.EmployeeID,
		FromDate:   fromDate,
		ToDate:     toDate,
		ProjectID:  req.ProjectID,
		IsBillable: req.IsBillable,
		Status:     req.Status,
		Limit:      req.Limit,
		Offset:     req.Offset,
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
