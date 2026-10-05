package bankaccounts

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/finance/bank_accounts/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/finance/bank_accounts/internal/api"
	"github.com/voxire/onion/internal/modules/finance/bank_accounts/internal/application"
	financebankaccountv1 "github.com/voxire/voxie/gen/go/finance/bank_account/v1"
)

type Module struct {
	service    *application.Service
	repository *moduleinfra.Repository
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repo := moduleinfra.NewRepository(db)
	return &Module{service: application.NewService(ctx, repo), repository: repo}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financebankaccountv1.RegisterBankAccountServiceServer(
		api.NewBankAccountAPI(m.service),
		financebankaccountv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance bank account service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) WithReportingCurrency(read func(context.Context, int64) (string, error)) *Module {
	m.repository.WithReportingCurrency(read)
	return m
}
