package oneonones

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/one_on_ones/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hr/one_on_ones/internal/api"
	"github.com/voxire/onion/internal/modules/hr/one_on_ones/internal/application"
	hroneononev1 "github.com/voxire/voxie/gen/go/hr/one_on_one/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hroneononev1.RegisterOneOnOneServiceServer(
		api.NewOneOnOneAPI(m.service),
		hroneononev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register one-on-one service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
