package invoices

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/api"
	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	financedashboardv1 "github.com/voxire/voxie/gen/go/finance/dashboard/v1"
	financeinvoicespaymentsv1 "github.com/voxire/voxie/gen/go/finance/invoices/payments/v1"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
	financereportv1 "github.com/voxire/voxie/gen/go/finance/report/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
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

// SetCustomerSnapshotProvider connects invoice creation to the customers
// module through the port invoices owns (#840). customers implements it; no
// package imports the other.
func (m *Module) SetCustomerSnapshotProvider(p application.CustomerSnapshotProvider) {
	m.service.SetCustomerSnapshotProvider(p)
}
