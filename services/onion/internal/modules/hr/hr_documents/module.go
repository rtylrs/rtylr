package hr_documents

import (
	"context"
	"database/sql"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"

	"github.com/voxire/onion/internal/modules/hr/hr_documents/internal/api"
	"github.com/voxire/onion/internal/modules/hr/hr_documents/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_documents/internal/infrastructure"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(infrastructure.NewRepository(db))}
}

func (m *Module) RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
