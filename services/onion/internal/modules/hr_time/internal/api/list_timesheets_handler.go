package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrtimeentryv1 "github.com/voxire/voxie/gen/go/hr/time_entry/v1"
)

func ListTimesheetsRoutes() []string { return []string{"/hr/timesheets"} }

type ListTimesheetsHandler struct{ Service *application.Service }

type ListTimesheetsRequest struct {
	EmployeeID *int64  `json:"employee_id"`
	Status     *string `json:"status"`
	Limit      int32   `json:"limit"`
	Offset     int32   `json:"offset"`
}

type ListTimesheetsResponse struct {
	Timesheets []*hrtimeentryv1.TimesheetPayload `json:"timesheets"`
	Total      int64                             `json:"total"`
	Limit      int32                             `json:"limit"`
	Offset     int32                             `json:"offset"`
}

const timesheetEntryPageSize int32 = 100

func (h *ListTimesheetsHandler) entriesForTimesheet(ctx context.Context, businessID int64, ts domain.TimesheetRecord) ([]*hrtimeentryv1.TimeEntryPayload, error) {
	employeeID := ts.EmployeeID
	fromDate := ts.PeriodStart
	toDate := ts.PeriodEnd
	records := make([]domain.TimeEntryRecord, 0)

	for offset := int32(0); ; {
		page, total, err := h.Service.ListTimeEntries(ctx, businessID, domain.ListTimeEntriesParams{
			EmployeeID: &employeeID,
			FromDate:   &fromDate,
			ToDate:     &toDate,
			Limit:      timesheetEntryPageSize,
			Offset:     offset,
		})
		if err != nil {
			return nil, err
		}

		// HR-created entries commonly have no timesheet_id, so their period is
		// the grouping key. An entry explicitly attached to another timesheet is
		// not part of this payload.
		for _, entry := range page {
			if entry.TimesheetID == ts.ID || entry.TimesheetID == 0 {
				records = append(records, entry)
			}
		}

		nextOffset := int64(offset) + int64(len(page))
		if len(page) == 0 || nextOffset >= total {
			break
		}
		if nextOffset > int64(1<<31-1) {
			return nil, fmt.Errorf("timesheet %d has too many entries to load", ts.ID)
		}
		offset = int32(nextOffset) // #nosec G115 -- the guard above keeps the conversion in int32 range.
	}

	payloads := make([]*hrtimeentryv1.TimeEntryPayload, 0, len(records))
	for _, entry := range records {
		payloads = append(payloads, timeEntryToProto(entry))
	}
	return payloads, nil
}

func (h *ListTimesheetsHandler) Serve(ctx context.Context, req *ListTimesheetsRequest) (ListTimesheetsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListTimesheetsResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListTimesheetsResponse{}, onekiterrors.PermissionDenied()
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

	timesheets, total, err := h.Service.ListTimesheetsForCaller(ctx, businessID, device.User.UserID, domain.ListTimesheetsParams{
		EmployeeID: req.EmployeeID,
		Status:     req.Status,
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		if errors.Is(err, domain.ErrTimesheetAccessDenied) {
			return ListTimesheetsResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to list timesheets", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListTimesheetsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list timesheets")
	}

	payloads := make([]*hrtimeentryv1.TimesheetPayload, 0, len(timesheets))
	for _, ts := range timesheets {
		entries, entryErr := h.entriesForTimesheet(ctx, businessID, ts)
		if entryErr != nil {
			logger.Ctx(ctx).Error("failed to load timesheet entries", logger.Any("owner_id", device.User.UserID), logger.Any("timesheet_id", ts.ID), logger.Error(entryErr))
			return ListTimesheetsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list timesheets")
		}
		payload := timesheetToProto(ts)
		payload.Entries = entries
		payloads = append(payloads, payload)
	}

	return ListTimesheetsResponse{Timesheets: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
