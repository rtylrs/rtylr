package accounts

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/finance/accounts/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/finance/accounts/internal/api"
	"github.com/voxire/onion/internal/modules/finance/accounts/internal/application"
	"github.com/voxire/onion/internal/platform/financeaccounts"
	"github.com/voxire/onion/internal/platform/financejournal"
	financeaccountv1 "github.com/voxire/voxie/gen/go/finance/account/v1"
)

type Module struct {
	service *application.Service
}

func (m *Module) PayrollAccountProvider() financeaccounts.PayrollAccountProvider {
	return m.service
}

// LedgerAccountProvider gives journal the standard accounts that approved
// expenses and invoice payments post to.
func (m *Module) LedgerAccountProvider() financeaccounts.LedgerAccountProvider {
	return m.service
}

// SetLedgerSyncer makes the account list post outstanding operational entries
// before it reads balances.
func (m *Module) SetLedgerSyncer(ledger financejournal.LedgerSyncer) {
	m.service.SetLedgerSyncer(ledger)
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financeaccountv1.RegisterAccountServiceServer(
		api.NewAccountAPI(m.service),
		financeaccountv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance account service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
