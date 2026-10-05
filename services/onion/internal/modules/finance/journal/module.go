package journal

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/finance/journal/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/finance/journal/internal/api"
	"github.com/voxire/onion/internal/modules/finance/journal/internal/application"
	"github.com/voxire/onion/internal/platform/financeaccounts"
	"github.com/voxire/onion/internal/platform/financejournal"
	financejournalv1 "github.com/voxire/voxie/gen/go/finance/journal/v1"
)

type Module struct {
	service *application.Service
}

func (m *Module) SetPayrollAccountProvider(provider financeaccounts.PayrollAccountProvider) {
	m.service.SetPayrollAccountProvider(provider)
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

// PayrollPoster adapts the central journal service to the payroll module's
// finance boundary. Keeping the adapter here ensures payroll cannot bypass
// journal validation, posting, or idempotency rules.
func (m *Module) PayrollPoster() financejournal.PayrollJournalPoster {
	return payrollPoster{service: m.service}
}

type payrollPoster struct {
	service *application.Service
}

func (p payrollPoster) PostPayrollJournal(ctx context.Context, entry financejournal.PayrollJournalEntry) error {
	return p.service.PostPayrollJournal(ctx, entry)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financejournalv1.RegisterJournalServiceServer(
		api.NewJournalAPI(m.service),
		financejournalv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance journal service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
