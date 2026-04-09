package routes

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, financeService *application.Service) {
	// Invoices
	http.RegisterGET[handlers.ListInvoicesRequest, handlers.ListInvoicesResponse](
		ctx,
		router,
		handlers.ListInvoicesRoutes(),
		&handlers.ListInvoicesHandler{FinanceService: financeService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateInvoiceRequest, handlers.CreateInvoiceResponse](
		ctx,
		router,
		handlers.CreateInvoiceRoutes(),
		&handlers.CreateInvoiceHandler{FinanceService: financeService},
	)

	http.RegisterGETWithValidation[handlers.GetInvoiceRequest, handlers.GetInvoiceResponse](
		ctx,
		router,
		handlers.GetInvoiceRoutes(),
		&handlers.GetInvoiceHandler{FinanceService: financeService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateInvoiceRequest, handlers.UpdateInvoiceResponse](
		ctx,
		router,
		handlers.UpdateInvoiceRoutes(),
		&handlers.UpdateInvoiceHandler{FinanceService: financeService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateInvoiceStatusRequest, handlers.UpdateInvoiceStatusResponse](
		ctx,
		router,
		handlers.UpdateInvoiceStatusRoutes(),
		&handlers.UpdateInvoiceStatusHandler{FinanceService: financeService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteInvoiceRequest, handlers.DeleteInvoiceResponse](
		ctx,
		router,
		handlers.DeleteInvoiceRoutes(),
		&handlers.DeleteInvoiceHandler{FinanceService: financeService},
	)

	http.RegisterPOSTWithValidation[handlers.RecordPaymentRequest, handlers.RecordPaymentResponse](
		ctx,
		router,
		handlers.RecordPaymentRoutes(),
		&handlers.RecordPaymentHandler{FinanceService: financeService},
	)

	// Expenses
	http.RegisterGET[handlers.ListExpensesRequest, handlers.ListExpensesResponse](
		ctx,
		router,
		handlers.ListExpensesRoutes(),
		&handlers.ListExpensesHandler{FinanceService: financeService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateExpenseRequest, handlers.CreateExpenseResponse](
		ctx,
		router,
		handlers.CreateExpenseRoutes(),
		&handlers.CreateExpenseHandler{FinanceService: financeService},
	)

	http.RegisterPUTWithValidation[handlers.UpdateExpenseRequest, handlers.UpdateExpenseResponse](
		ctx,
		router,
		handlers.UpdateExpenseRoutes(),
		&handlers.UpdateExpenseHandler{FinanceService: financeService},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteExpenseRequest, handlers.DeleteExpenseResponse](
		ctx,
		router,
		handlers.DeleteExpenseRoutes(),
		&handlers.DeleteExpenseHandler{FinanceService: financeService},
	)

	// Expense Categories
	http.RegisterGET[handlers.ListExpenseCategoriesRequest, handlers.ListExpenseCategoriesResponse](
		ctx,
		router,
		handlers.ListExpenseCategoriesRoutes(),
		&handlers.ListExpenseCategoriesHandler{FinanceService: financeService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateExpenseCategoryRequest, handlers.CreateExpenseCategoryResponse](
		ctx,
		router,
		handlers.CreateExpenseCategoryRoutes(),
		&handlers.CreateExpenseCategoryHandler{FinanceService: financeService},
	)

	// Reports
	http.RegisterGET[handlers.PLReportRequest, handlers.PLReportResponse](
		ctx,
		router,
		handlers.PLReportRoutes(),
		&handlers.PLReportHandler{FinanceService: financeService},
	)

	http.RegisterGET[handlers.DashboardRequest, handlers.DashboardResponse](
		ctx,
		router,
		handlers.DashboardRoutes(),
		&handlers.DashboardHandler{FinanceService: financeService},
	)

	// Tax
	http.RegisterGET[handlers.ListTaxConfigRequest, handlers.ListTaxConfigResponse](
		ctx,
		router,
		handlers.ListTaxConfigRoutes(),
		&handlers.ListTaxConfigHandler{FinanceService: financeService},
	)

	http.RegisterPOSTWithValidation[handlers.CreateTaxConfigRequest, handlers.CreateTaxConfigResponse](
		ctx,
		router,
		handlers.CreateTaxConfigRoutes(),
		&handlers.CreateTaxConfigHandler{FinanceService: financeService},
	)

	http.RegisterGET[handlers.TaxReportRequest, handlers.TaxReportResponse](
		ctx,
		router,
		handlers.TaxReportRoutes(),
		&handlers.TaxReportHandler{FinanceService: financeService},
	)
}
