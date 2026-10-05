package floors

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/pos/floors/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/pos/floors/internal/api"
	"github.com/voxire/onion/internal/modules/pos/floors/internal/application"
	posfloorv1 "github.com/voxire/voxie/gen/go/pos/floor/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posfloorv1.RegisterFloorServiceServer(
		api.NewFloorAPI(m.service),
		posfloorv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS floor service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
