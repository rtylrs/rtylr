package careers

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/careers/internal/application"
	"github.com/voxire/onion/internal/modules/careers/internal/routes"
	"github.com/voxire/onion/internal/modules/careers/internal/routes/handlers"
)

type Module struct {
	CareersService *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	careersService := application.NewService(ctx, db)

	return &Module{
		CareersService: careersService,
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.CareersService)
}

// InvalidateActiveCareers busts the cached public careers catalogue.
// Call this after writing to the shared careers table from outside this
// module (e.g. voxhire's admin panel writes into the same table).
func InvalidateActiveCareers(ctx context.Context) {
	application.InvalidateActiveCareersCache(ctx)
}

func (m *Module) GetAnonymousRoutes() []string {
	// /careers is the legacy voxire-corporate list. The /recruit/public/*
	// routes back the candidate-facing careers sites — they resolve the
	// tenant from the URL slug so auth isn't needed.
	anon := []string{}
	anon = append(anon, handlers.CareersRoutes()...)
	anon = append(anon, handlers.ApplyRoutes()...)
	anon = append(anon, handlers.GetPublicBrandingRoutes()...)
	anon = append(anon, handlers.ListPublicCareersRoutes()...)
	anon = append(anon, handlers.GetPublicCareerRoutes()...)
	anon = append(anon, handlers.SubmitApplicationRoutes()...)
	anon = append(anon, handlers.ResumeUploadURLRoutes()...)
	anon = append(anon, handlers.PublicOGRoutes()...)
	return anon
}
