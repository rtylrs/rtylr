package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/erp/products/internal/catalog/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/erp/products/internal/catalog/internal/api"
	"github.com/voxire/onion/internal/modules/erp/products/internal/catalog/internal/application"
	posmenuitemv1 "github.com/voxire/voxie/gen/go/pos/menu_item/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) SetOnlineOrdering(onlineOrdering func(context.Context, int64) (bool, error)) {
	m.service.SetOnlineOrdering(onlineOrdering)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posmenuitemv1.RegisterMenuItemServiceServer(
		api.NewMenuItemAPI(m.service),
		posmenuitemv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS menu item service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/pos/menu-item/v1/menu-items/public",
	}
}
