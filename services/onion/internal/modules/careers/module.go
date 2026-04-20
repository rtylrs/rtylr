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
	// /careers is the legacy voxire-corporate list. The /recruit/public/*
	// routes back the candidate-facing careers sites — they resolve the
	// tenant from the URL slug so auth isn't needed.
	anon := []string{}
	anon = append(anon, handlers.CareersRoutes()...)
	anon = append(anon, handlers.GetPublicBrandingRoutes()...)
	anon = append(anon, handlers.ListPublicCareersRoutes()...)
	anon = append(anon, handlers.GetPublicCareerRoutes()...)
	anon = append(anon, handlers.SubmitApplicationRoutes()...)
	return anon
}
