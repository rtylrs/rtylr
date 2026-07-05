package financereconciliation

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	financereconciliationv1 "github.com/voxire/voxie/gen/go/finance/reconciliation/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/api"
	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financereconciliationv1.RegisterReconciliationServiceServer(
		api.NewReconciliationAPI(m.service),
		financereconciliationv1.WithMux(mux),
		financereconciliationv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register finance reconciliation service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
