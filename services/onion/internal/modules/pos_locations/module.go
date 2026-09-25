package pos_locations

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_locations/internal/api"
	"github.com/voxire/onion/internal/modules/pos_locations/internal/application"
	poslocationv1 "github.com/voxire/voxie/gen/go/pos/location/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := poslocationv1.RegisterLocationServiceServer(
		api.NewLocationAPI(m.service),
		poslocationv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS location service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
