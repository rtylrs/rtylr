package pos_menu

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_menu/internal/api"
	"github.com/voxire/onion/internal/modules/pos_menu/internal/application"
	posmenuitemv1 "github.com/voxire/voxie/gen/go/pos/menu_item/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
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
