package hubmenu

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/api"
	"github.com/voxire/onion/internal/modules/hub_menu/internal/application"
	hubmenuv1 "github.com/voxire/voxie/gen/go/hub/menu/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubmenuv1.RegisterMenuServiceServer(
		api.NewMenuAPI(m.service),
		hubmenuv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub menu service: %w", err))
	}
}
