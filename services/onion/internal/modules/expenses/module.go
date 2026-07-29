package expenses

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/expenses/internal/api"
	"github.com/voxire/onion/internal/modules/expenses/internal/application"
	financeexpensesv1 "github.com/voxire/voxie/gen/go/finance/expenses/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financeexpensesv1.RegisterExpensesServiceServer(
		api.NewExpensesAPI(m.service),
		financeexpensesv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance expenses service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
