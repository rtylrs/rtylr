package financereconciliation

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/finance/finance_reconciliation/internal/api"
	"github.com/voxire/onion/internal/modules/finance/finance_reconciliation/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/finance/finance_reconciliation/internal/infrastructure"
	financereconciliationv1 "github.com/voxire/voxie/gen/go/finance/reconciliation/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financereconciliationv1.RegisterReconciliationServiceServer(
		api.NewReconciliationAPI(m.service),
		financereconciliationv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance reconciliation service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
