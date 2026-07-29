package stocktransfers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/stock_transfers/internal/api"
	"github.com/voxire/onion/internal/modules/stock_transfers/internal/application"
	erpstocktransferv1 "github.com/voxire/voxie/gen/go/erp/stock_transfer/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := erpstocktransferv1.RegisterStockTransferServiceServer(
		api.NewStockTransferAPI(m.service),
		erpstocktransferv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register ERP stock transfer service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
