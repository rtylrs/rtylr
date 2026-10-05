package apikeys

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/organization/api_keys/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/organization/api_keys/internal/api"
	"github.com/voxire/onion/internal/modules/organization/api_keys/internal/application"
	"github.com/voxire/onion/internal/platform/hubdeveloper"
	hubdeveloperv1 "github.com/voxire/voxie/gen/go/hub/developer/v1"
)

type Module struct {
	service         *application.Service
	webhooksService hubdeveloper.WebhooksProvider
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, moduleinfra.NewRepository(db)),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubdeveloperv1.RegisterDeveloperServiceServer(
		api.NewDeveloperAPI(NewHubAPIKeysProvider(m.service), m.webhooksService),
		hubdeveloperv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub developer service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) SetWebhooksProvider(provider hubdeveloper.WebhooksProvider) {
	m.webhooksService = provider
}
