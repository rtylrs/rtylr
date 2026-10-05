package crmsettings

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/crm/crm_settings/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/crm/crm_settings/internal/api"
	"github.com/voxire/onion/internal/modules/crm/crm_settings/internal/application"
	crmsettingsv1 "github.com/voxire/voxie/gen/go/crm/settings/v1"
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
	if err := crmsettingsv1.RegisterSettingsServiceServer(
		api.NewSettingsAPI(m.service),
		crmsettingsv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register crm settings service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
