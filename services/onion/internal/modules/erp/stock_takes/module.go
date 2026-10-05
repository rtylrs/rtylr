package stocktakes

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/erp/stock_takes/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/erp/stock_takes/internal/api"
	"github.com/voxire/onion/internal/modules/erp/stock_takes/internal/application"
	erpstocktakev1 "github.com/voxire/voxie/gen/go/erp/stock_take/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, moduleinfra.NewRepository(db)),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := erpstocktakev1.RegisterStockTakeServiceServer(
		api.NewStockTakeAPI(m.service),
		erpstocktakev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register ERP stock take service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
