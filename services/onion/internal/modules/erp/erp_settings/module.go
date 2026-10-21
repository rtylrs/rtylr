package erpsettings

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/erp/erp_settings/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/erp/erp_settings/internal/api"
	"github.com/voxire/onion/internal/modules/erp/erp_settings/internal/application"
	erpsettingsv1 "github.com/voxire/voxie/gen/go/erp/settings/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := erpsettingsv1.RegisterSettingsServiceServer(
		api.NewSettingsAPI(m.service),
		erpsettingsv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register ERP settings service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
