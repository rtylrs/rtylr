package api

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// These hand-registered paths implement the HR compatibility contract. The
	// route-contract-check target compares this table with the .onk declarations.
	// Time Entries
	serverHttp.RegisterGET[ListTimeEntriesRequest, ListTimeEntriesResponse](
		ctx, router, ListTimeEntriesRoutes(), &ListTimeEntriesHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[CreateTimeEntryRequest, CreateTimeEntryResponse](
		ctx, router, CreateTimeEntryRoutes(), &CreateTimeEntryHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[UpdateTimeEntryRequest, UpdateTimeEntryResponse](
		ctx, router, UpdateTimeEntryRoutes(), &UpdateTimeEntryHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[DeleteTimeEntryRequest, DeleteTimeEntryResponse](
		ctx, router, DeleteTimeEntryRoutes(), &DeleteTimeEntryHandler{Service: service},
	)

	// Timesheets
	serverHttp.RegisterGET[GetTimesheetRequest, GetTimesheetResponse](
		ctx, router, GetTimesheetRoutes(), &GetTimesheetHandler{Service: service},
	)
	serverHttp.RegisterGET[GetTimesheetByIDRequest, GetTimesheetByIDResponse](
		ctx, router, GetTimesheetByIDRoutes(), &GetTimesheetByIDHandler{Service: service},
	)
	serverHttp.RegisterGET[ListTimesheetsRequest, ListTimesheetsResponse](
		ctx, router, ListTimesheetsRoutes(), &ListTimesheetsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[SubmitTimesheetRequest, SubmitTimesheetResponse](
		ctx, router, SubmitTimesheetRoutes(), &SubmitTimesheetHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[ApproveTimesheetRequest, ApproveTimesheetResponse](
		ctx, router, ApproveTimesheetRoutes(), &ApproveTimesheetHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[RejectTimesheetRequest, RejectTimesheetResponse](
		ctx, router, RejectTimesheetRoutes(), &RejectTimesheetHandler{Service: service},
	)

	// Mark entries invoiced (the billable loop's invoice-side write, #836/#791)
	serverHttp.RegisterPATCHWithValidation[MarkTimeEntriesInvoicedRequest, MarkTimeEntriesInvoicedResponse](
		ctx, router, MarkTimeEntriesInvoicedRoutes(), &MarkTimeEntriesInvoicedHandler{Service: service},
	)
}
