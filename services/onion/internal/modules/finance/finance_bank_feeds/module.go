package financebankfeeds

import (
	"context"
	"database/sql"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"

	moduleinfra "github.com/voxire/onion/internal/modules/finance/finance_bank_feeds/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/finance/finance_bank_feeds/internal/api"
	"github.com/voxire/onion/internal/modules/finance/finance_bank_feeds/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
