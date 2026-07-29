package pos_manager

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/api"
	"github.com/voxire/onion/internal/modules/pos_manager/internal/application"
	posmanagerv1 "github.com/voxire/voxie/gen/go/pos/manager/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posmanagerv1.RegisterManagerServiceServer(
		api.NewManagerAPI(m.service),
		posmanagerv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS manager service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
