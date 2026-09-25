package automations

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/automations/internal/api"
	"github.com/voxire/onion/internal/modules/automations/internal/application"
	"github.com/voxire/onion/internal/modules/automations/internal/jobs"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

type Module struct {
	service  *application.Service
	executor *application.Executor
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service:  application.NewService(ctx, db),
		executor: application.NewExecutor(db),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

// RegisterJobs is the hook platformjobs.AutoRegister looks for. Execution is a
// worker concern: the API process serves automation CRUD and never runs one.
func (m *Module) RegisterJobs(ctx context.Context, registry *platformjobs.Registry) {
	jobs.Register(ctx, registry, m.executor)
}
