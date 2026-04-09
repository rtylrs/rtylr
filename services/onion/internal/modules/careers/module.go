package careers

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/careers/internal/application"
	"github.com/voxire/onion/internal/modules/careers/internal/routes"
	"github.com/voxire/onion/internal/modules/careers/internal/routes/handlers"
)

type Module struct {
	CareersService *application.Service
}

func NewModule(ctx context.Context) *Module {
	careersService := application.NewService(ctx)

	return &Module{
		CareersService: careersService,
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.CareersService)
}

func (m *Module) GetAnonymousRoutes() []string {
	return handlers.CareersRoutes()
}
