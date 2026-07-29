package bankaccounts

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/api"
	"github.com/voxire/onion/internal/modules/bank_accounts/internal/application"
	financebankaccountv1 "github.com/voxire/voxie/gen/go/finance/bank_account/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
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
