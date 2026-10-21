package supportsettings

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/support/settings/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/support/settings/internal/api"
	"github.com/voxire/onion/internal/modules/support/settings/internal/application"
	supportsettingsv1 "github.com/voxire/voxie/gen/go/support/settings/v1"
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
	if err := supportsettingsv1.RegisterSettingsServiceServer(
		api.NewSettingsAPI(m.service),
		supportsettingsv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register support settings service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
