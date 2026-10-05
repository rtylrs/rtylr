package purchase_orders

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/api"
	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/infrastructure"
	erppurchaseorderv1 "github.com/voxire/voxie/gen/go/erp/purchase_order/v1"
)

type Module struct {
	service *application.Service
	repo    *moduleinfra.Repository
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repo := moduleinfra.NewRepository(db)
	return &Module{repo: repo, service: application.NewService(repo)}
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

// Exists validates a current business-scoped reference without exposing ERP records or write authority.
func (m *Module) Exists(ctx context.Context, businessID, id int64) (bool, error) {
	return m.repo.Exists(ctx, businessID, id)
}
