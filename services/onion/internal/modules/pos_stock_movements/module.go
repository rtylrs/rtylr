package pos_stock_movements

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_stock_movements/internal/api"
	"github.com/voxire/onion/internal/modules/pos_stock_movements/internal/application"
	posstockv1 "github.com/voxire/voxie/gen/go/pos/stock/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posstockv1.RegisterStockServiceServer(
		api.NewStockAPI(m.service),
		posstockv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS stock service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
