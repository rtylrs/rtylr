package invoices

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/api"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	moduleinfra "github.com/voxire/onion/internal/modules/finance/invoices/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/email"
	financedashboardv1 "github.com/voxire/voxie/gen/go/finance/dashboard/v1"
	financeinvoicespaymentsv1 "github.com/voxire/voxie/gen/go/finance/invoices/payments/v1"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
	financereportv1 "github.com/voxire/voxie/gen/go/finance/report/v1"
)

type Module struct {
	service *application.Service
}

type BillableTimeEntry = domain.BillableTimeEntry
type BillableTimeReader = application.BillableTimeReader

func NewModule(ctx context.Context, db *sql.DB) *Module {
	sendInvoiceEmail := func(ctx context.Context, invoice domain.Invoice) error {
		subject := fmt.Sprintf("Invoice %s from rtylr", invoice.InvoiceNumber)
		return email.EnqueueTemplate(ctx, db, email.PriorityHigh, invoice.ClientEmail, subject, email.TemplateInvoice, email.InvoiceData{
			Subject:       subject,
			CustomerName:  invoice.ClientName,
			InvoiceNumber: invoice.InvoiceNumber,
			Amount:        fmt.Sprintf("%.2f %s", invoice.Total, invoice.Currency),
			DueDate:       invoice.DueDate.Format("2006-01-02"),
		})
	}
	return &Module{
		service: application.NewService(moduleinfra.NewRepository(db), sendInvoiceEmail),
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

func (m *Module) SetBillableTimeReader(reader application.BillableTimeReader) {
	m.service.SetBillableTimeReader(reader)
}

func (m *Module) InvoiceExists(ctx context.Context, businessID, invoiceID int64) (bool, error) {
	return m.service.InvoiceExists(ctx, businessID, invoiceID)
}
