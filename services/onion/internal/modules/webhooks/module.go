package webhooks

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/webhooks/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) GetService() *application.Service {
	return m.service
}
