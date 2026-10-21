package templates

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/crm/templates/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/crm/templates/internal/api"
	"github.com/voxire/onion/internal/modules/crm/templates/internal/application"
	crmtemplatev1 "github.com/voxire/voxie/gen/go/crm/template/v1"
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
	if err := crmtemplatev1.RegisterTemplateServiceServer(
		api.NewTemplateAPI(m.service),
		crmtemplatev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register crm template service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
