package websitelead

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/websitelead/internal/api"
	"github.com/voxire/onion/internal/modules/websitelead/internal/application"
)

// Notify re-exports the recipient configuration so cmd/api does not import the
// module's internal packages.
type Notify = application.Notify

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB, notify Notify) *Module {
	return &Module{
		service: application.NewService(ctx, db, notify),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{"/quote", "/contact", "/newsletter", "/analytics/event"}
}
