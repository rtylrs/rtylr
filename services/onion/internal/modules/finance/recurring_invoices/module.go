package recurringinvoices

import (
	"context"
	"database/sql"
	"fmt"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/finance/recurring_invoices/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/finance/recurring_invoices/internal/api"
	"github.com/voxire/onion/internal/modules/finance/recurring_invoices/internal/application"
	internaljobs "github.com/voxire/onion/internal/modules/finance/recurring_invoices/internal/jobs"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
	financerecurringinvoicev1 "github.com/voxire/voxie/gen/go/finance/recurring_invoice/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, _ ...interface{}) {
	_ = ctx
	_ = router
}

func (m *Module) RegisterJobs(ctx context.Context, registry *platformjobs.Registry) {
	internaljobs.Register(ctx, registry, m.service)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financerecurringinvoicev1.RegisterRecurringInvoiceServiceServer(
		api.NewRecurringInvoiceAPI(m.service),
		financerecurringinvoicev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance recurring invoice service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
