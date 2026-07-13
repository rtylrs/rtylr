package accounts

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	financeaccountv1 "github.com/voxire/voxie/gen/go/finance/account/v1"

	"github.com/voxire/onion/internal/modules/accounts/internal/api"
	"github.com/voxire/onion/internal/modules/accounts/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
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
