package featuresupport

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/featuresupport/internal/application"
	"github.com/voxire/onion/internal/modules/featuresupport/internal/routes"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}

// GetAnonymousRoutes exposes /features to the landing page without auth. Note
// the per-user reads use /features/me (authenticated, not listed here), and the
// /features/admin/* surface stays behind auth + an owner/admin role gate.
func (m *Module) GetAnonymousRoutes() []string {
	return []string{"/features"}
}
