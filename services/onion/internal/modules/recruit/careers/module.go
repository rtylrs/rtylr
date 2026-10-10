package careers

import (
	"context"
	"database/sql"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/api"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/application"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/infrastructure"
)

type Module struct {
	CareersService *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	careersService := application.NewService(infrastructure.NewRepository(db))

	return &Module{
		CareersService: careersService,
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.CareersService)
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
	anon = append(anon, api.CareersRoutes()...)
	anon = append(anon, api.ApplyRoutes()...)
	anon = append(anon, api.GetPublicBrandingRoutes()...)
	anon = append(anon, api.ListPublicCareersRoutes()...)
	anon = append(anon, api.GetPublicCareerRoutes()...)
	anon = append(anon, api.SubmitApplicationRoutes()...)
	anon = append(anon, api.ResumeUploadURLRoutes()...)
	anon = append(anon, api.PublicOGRoutes()...)
	return anon
}
