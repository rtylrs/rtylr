package invoices

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	financedashboardv1 "github.com/voxire/voxie/gen/go/finance/dashboard/v1"
	financeinvoicespaymentsv1 "github.com/voxire/voxie/gen/go/finance/invoices/payments/v1"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
	financereportv1 "github.com/voxire/voxie/gen/go/finance/report/v1"

	"github.com/voxire/onion/internal/modules/invoices/internal/api"
	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	serverHttp.RegisterGET[api.ListInvoicesRequest, api.ListInvoicesResponse](
		ctx,
		router,
		api.ListInvoicesRoutes(),
		&api.ListInvoicesHandler{FinanceService: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.CreateInvoiceRequest, api.CreateInvoiceResponse](
		ctx,
		router,
		api.CreateInvoiceRoutes(),
		&api.CreateInvoiceHandler{FinanceService: m.service},
	)
	serverHttp.RegisterGETWithValidation[api.GetInvoiceRequest, api.GetInvoiceResponse](
		ctx,
		router,
		api.GetInvoiceRoutes(),
		&api.GetInvoiceHandler{FinanceService: m.service},
	)
	serverHttp.RegisterPUTWithValidation[api.UpdateInvoiceRequest, api.UpdateInvoiceResponse](
		ctx,
		router,
		api.UpdateInvoiceRoutes(),
		&api.UpdateInvoiceHandler{FinanceService: m.service},
	)
	serverHttp.RegisterPUTWithValidation[api.UpdateInvoiceStatusRequest, api.UpdateInvoiceStatusResponse](
		ctx,
		router,
		api.UpdateInvoiceStatusRoutes(),
		&api.UpdateInvoiceStatusHandler{FinanceService: m.service},
	)
	serverHttp.RegisterDELETEWithValidation[api.DeleteInvoiceRequest, api.DeleteInvoiceResponse](
		ctx,
		router,
		api.DeleteInvoiceRoutes(),
		&api.DeleteInvoiceHandler{FinanceService: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.RecordPaymentRequest, api.RecordPaymentResponse](
		ctx,
		router,
		api.RecordPaymentRoutes(),
		&api.RecordPaymentHandler{FinanceService: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.SendInvoiceRequest, api.SendInvoiceResponse](
		ctx,
		router,
		api.SendInvoiceRoutes(),
		&api.SendInvoiceHandler{FinanceService: m.service},
	)
	serverHttp.RegisterGET[api.ListExpensesRequest, api.ListExpensesResponse](
		ctx,
		router,
		api.ListExpensesRoutes(),
		&api.ListExpensesHandler{FinanceService: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.CreateExpenseRequest, api.CreateExpenseResponse](
		ctx,
		router,
		api.CreateExpenseRoutes(),
		&api.CreateExpenseHandler{FinanceService: m.service},
	)
	serverHttp.RegisterPUTWithValidation[api.UpdateExpenseRequest, api.UpdateExpenseResponse](
		ctx,
		router,
		api.UpdateExpenseRoutes(),
		&api.UpdateExpenseHandler{FinanceService: m.service},
	)
	serverHttp.RegisterDELETEWithValidation[api.DeleteExpenseRequest, api.DeleteExpenseResponse](
		ctx,
		router,
		api.DeleteExpenseRoutes(),
		&api.DeleteExpenseHandler{FinanceService: m.service},
	)
	serverHttp.RegisterGET[api.ListExpenseCategoriesRequest, api.ListExpenseCategoriesResponse](
		ctx,
		router,
		api.ListExpenseCategoriesRoutes(),
		&api.ListExpenseCategoriesHandler{FinanceService: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.CreateExpenseCategoryRequest, api.CreateExpenseCategoryResponse](
		ctx,
		router,
		api.CreateExpenseCategoryRoutes(),
		&api.CreateExpenseCategoryHandler{FinanceService: m.service},
	)
	serverHttp.RegisterGET[api.DashboardRequest, api.DashboardResponse](
		ctx,
		router,
		api.DashboardRoutes(),
		&api.DashboardHandler{FinanceService: m.service},
	)
	serverHttp.RegisterGET[api.ListTaxConfigRequest, api.ListTaxConfigResponse](
		ctx,
		router,
		api.ListTaxConfigRoutes(),
		&api.ListTaxConfigHandler{FinanceService: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.CreateTaxConfigRequest, api.CreateTaxConfigResponse](
		ctx,
		router,
		api.CreateTaxConfigRoutes(),
		&api.CreateTaxConfigHandler{FinanceService: m.service},
	)
	serverHttp.RegisterGET[api.TaxReportRequest, api.TaxReportResponse](
		ctx,
		router,
		api.TaxReportRoutes(),
		&api.TaxReportHandler{FinanceService: m.service},
	)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financereportv1.RegisterReportServiceServer(
		api.NewReportAPI(m.service),
		financereportv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance report service: %w", err))
	}

	if err := financeinvoicesv1.RegisterInvoicesServiceServer(
		api.NewInvoicesAPI(m.service),
		financeinvoicesv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance invoices service: %w", err))
	}

	if err := financeinvoicespaymentsv1.RegisterInvoicePaymentServiceServer(
		api.NewInvoicePaymentAPI(m.service),
		financeinvoicespaymentsv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance invoice payment service: %w", err))
	}

	if err := financedashboardv1.RegisterDashboardServiceServer(
		api.NewDashboardAPI(m.service),
		financedashboardv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance dashboard service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
