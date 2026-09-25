package purchase_orders

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/api"
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/application"
	erppurchaseorderv1 "github.com/voxire/voxie/gen/go/erp/purchase_order/v1"
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
	if err := erppurchaseorderv1.RegisterPurchaseOrderServiceServer(
		api.NewPurchaseOrderAPI(m.service),
		erppurchaseorderv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register ERP purchase order service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
