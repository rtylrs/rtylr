package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrtimeentryv1 "github.com/voxire/voxie/gen/go/hr/time_entry/v1"
)

// TimeEntryAPI serves the generated hr.time_entry.v1 contract with the same
// handlers as the hand-registered /hr/time-entries and /hr/timesheets routes.
// The contract had clients but no server, so finance's billable-time picker
// and its mark-invoiced write-back were 404s.
type TimeEntryAPI struct {
	service *application.Service
}

func NewTimeEntryAPI(service *application.Service) *TimeEntryAPI {
	return &TimeEntryAPI{service: service}
}

func (a *TimeEntryAPI) ListTimeEntries(ctx context.Context, req *hrtimeentryv1.ListTimeEntriesRequest) (*hrtimeentryv1.ListTimeEntriesResponse, error) {
	legacy := &ListTimeEntriesRequest{
		EmployeeID: req.EmployeeId,
		FromDate:   req.FromDate,
		ToDate:     req.ToDate,
		ProjectID:  req.ProjectId,
		IsBillable: req.IsBillable,
		Status:     req.Status,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}
	if req.Date != nil {
		legacy.Date = *req.Date
	}
	res, err := (&ListTimeEntriesHandler{Service: a.service}).Serve(ctx, legacy)
	if err != nil {
		return nil, err
	}
	return &hrtimeentryv1.ListTimeEntriesResponse{Entries: res.Entries, Total: res.Total, Limit: res.Limit, Offset: res.Offset}, nil
}

func (a *TimeEntryAPI) CreateTimeEntry(ctx context.Context, req *hrtimeentryv1.CreateTimeEntryRequest) (*hrtimeentryv1.CreateTimeEntryResponse, error) {
	legacy := &CreateTimeEntryRequest{
		EmployeeID:      deref(req.EmployeeId),
		ProjectID:       deref(req.ProjectId),
		TaskID:          deref(req.TaskId),
		CategoryID:      deref(req.CategoryId),
		Date:            req.Date,
		DurationMinutes: deref(req.DurationMinutes),
		StartTime:       deref(req.StartTime),
		EndTime:         deref(req.EndTime),
		Description:     req.Description,
		IsBillable:      req.IsBillable,
		Source:          req.Source,
	}
	handler := &CreateTimeEntryHandler{Service: a.service}
	if err := handler.Validate(ctx, legacy); err != nil {
		return nil, err
	}
	res, err := handler.Serve(ctx, legacy)
	if err != nil {
		return nil, err
	}
	return &hrtimeentryv1.CreateTimeEntryResponse{Entry: res.Entry}, nil
}

func (a *TimeEntryAPI) ListTimesheets(ctx context.Context, req *hrtimeentryv1.ListTimesheetsRequest) (*hrtimeentryv1.ListTimesheetsResponse, error) {
	res, err := (&ListTimesheetsHandler{Service: a.service}).Serve(ctx, &ListTimesheetsRequest{
		EmployeeID: req.EmployeeId,
		Status:     req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &hrtimeentryv1.ListTimesheetsResponse{Timesheets: res.Timesheets, Total: res.Total, Limit: res.Limit, Offset: res.Offset}, nil
}

func (a *TimeEntryAPI) SubmitTimesheet(ctx context.Context, req *hrtimeentryv1.SubmitTimesheetRequest) (*hrtimeentryv1.SubmitTimesheetResponse, error) {
	if req.TimesheetId == nil || *req.TimesheetId <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "timesheet_id is required")
	}
	if _, err := (&SubmitTimesheetHandler{Service: a.service}).submit(ctx, *req.TimesheetId); err != nil {
		return nil, err
	}
	return &hrtimeentryv1.SubmitTimesheetResponse{Success: true}, nil
}

// ApproveTimesheet carries both outcomes: approved=false rejects with the
// given reason, as the contract documents.
func (a *TimeEntryAPI) ApproveTimesheet(ctx context.Context, req *hrtimeentryv1.ApproveTimesheetRequest) (*hrtimeentryv1.ApproveTimesheetResponse, error) {
	if req.Id <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid timesheet id")
	}
	if req.Body != nil && !req.Body.Approved {
		if _, err := (&RejectTimesheetHandler{Service: a.service}).reject(ctx, req.Id, deref(req.Body.RejectionReason)); err != nil {
			return nil, err
		}
		return &hrtimeentryv1.ApproveTimesheetResponse{Success: true}, nil
	}
	if _, err := (&ApproveTimesheetHandler{Service: a.service}).approve(ctx, req.Id); err != nil {
		return nil, err
	}
	return &hrtimeentryv1.ApproveTimesheetResponse{Success: true}, nil
}

func (a *TimeEntryAPI) MarkTimeEntriesInvoiced(ctx context.Context, req *hrtimeentryv1.MarkTimeEntriesInvoicedRequest) (*hrtimeentryv1.MarkTimeEntriesInvoicedResponse, error) {
	legacy := &MarkTimeEntriesInvoicedRequest{}
	if req.Body != nil {
		legacy.EntryIDs = req.Body.EntryIds
		legacy.InvoiceID = deref(req.Body.InvoiceId)
	}
	handler := &MarkTimeEntriesInvoicedHandler{Service: a.service}
	if err := handler.Validate(ctx, legacy); err != nil {
		return nil, err
	}
	res, err := handler.Serve(ctx, legacy)
	if err != nil {
		return nil, err
	}
	return &hrtimeentryv1.MarkTimeEntriesInvoicedResponse{MarkedCount: res.MarkedCount, SkippedCount: res.SkippedCount}, nil
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
