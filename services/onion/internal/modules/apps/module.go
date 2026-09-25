package apps

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/apps/internal/api"
	"github.com/voxire/onion/internal/modules/apps/internal/application"
	"github.com/voxire/onion/internal/platform/hubapp"
	hubappv1 "github.com/voxire/voxie/gen/go/hub/app/v1"
)

type Module struct {
	service       *application.Service
	addonsService hubapp.AddonProvider
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubappv1.RegisterAppServiceServer(
		api.NewAppAPI(m.service, m.addonsService),
		hubappv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub app service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) SetAddonProvider(provider hubapp.AddonProvider) {
	m.addonsService = provider
}
