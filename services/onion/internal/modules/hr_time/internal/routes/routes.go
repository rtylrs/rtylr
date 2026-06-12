package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr_time/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, service *application.Service) {
	// Time Entries
	serverHttp.RegisterGET[handlers.ListTimeEntriesRequest, handlers.ListTimeEntriesResponse](
		ctx, router, handlers.ListTimeEntriesRoutes(), &handlers.ListTimeEntriesHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.CreateTimeEntryRequest, handlers.CreateTimeEntryResponse](
		ctx, router, handlers.CreateTimeEntryRoutes(), &handlers.CreateTimeEntryHandler{Service: service},
	)
	serverHttp.RegisterPUTWithValidation[handlers.UpdateTimeEntryRequest, handlers.UpdateTimeEntryResponse](
		ctx, router, handlers.UpdateTimeEntryRoutes(), &handlers.UpdateTimeEntryHandler{Service: service},
	)
	serverHttp.RegisterDELETEWithValidation[handlers.DeleteTimeEntryRequest, handlers.DeleteTimeEntryResponse](
		ctx, router, handlers.DeleteTimeEntryRoutes(), &handlers.DeleteTimeEntryHandler{Service: service},
	)

	// Timesheets
	serverHttp.RegisterGET[handlers.GetTimesheetRequest, handlers.GetTimesheetResponse](
		ctx, router, handlers.GetTimesheetRoutes(), &handlers.GetTimesheetHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.GetTimesheetByIDRequest, handlers.GetTimesheetByIDResponse](
		ctx, router, handlers.GetTimesheetByIDRoutes(), &handlers.GetTimesheetByIDHandler{Service: service},
	)
	serverHttp.RegisterGET[handlers.ListTimesheetsRequest, handlers.ListTimesheetsResponse](
		ctx, router, handlers.ListTimesheetsRoutes(), &handlers.ListTimesheetsHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.SubmitTimesheetRequest, handlers.SubmitTimesheetResponse](
		ctx, router, handlers.SubmitTimesheetRoutes(), &handlers.SubmitTimesheetHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.ApproveTimesheetRequest, handlers.ApproveTimesheetResponse](
		ctx, router, handlers.ApproveTimesheetRoutes(), &handlers.ApproveTimesheetHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.RejectTimesheetRequest, handlers.RejectTimesheetResponse](
		ctx, router, handlers.RejectTimesheetRoutes(), &handlers.RejectTimesheetHandler{Service: service},
	)

	// Flow import & billable export
	serverHttp.RegisterPOSTWithValidation[handlers.ImportFlowRequest, handlers.ImportFlowResponse](
		ctx, router, handlers.ImportFlowRoutes(), &handlers.ImportFlowHandler{Service: service},
	)
	serverHttp.RegisterPOSTWithValidation[handlers.ExportBillableRequest, handlers.ExportBillableResponse](
		ctx, router, handlers.ExportBillableRoutes(), &handlers.ExportBillableHandler{Service: service},
	)
}
