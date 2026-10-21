package integrations

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/integrations/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/integrations/internal/api"
	"github.com/voxire/onion/internal/modules/integrations/internal/application"
	hubintegrationv1 "github.com/voxire/voxie/gen/go/hub/integration/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, moduleinfra.NewRepository(db)),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubintegrationv1.RegisterIntegrationServiceServer(
		api.NewIntegrationAPI(m.service),
		hubintegrationv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub integration service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
