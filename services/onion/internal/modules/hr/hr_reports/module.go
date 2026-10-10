package hr_reports

import (
	"context"
	"database/sql"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_reports/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/hr_reports/internal/api"
	"github.com/voxire/onion/internal/modules/hr/hr_reports/internal/application"
)

type Module struct {
	service     *application.Service
	permissions *hr_core.Module
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db)), permissions: hr_core.NewModule(ctx, db)}
}

func (m *Module) RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service, m.permissions)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
