// Package voxhire wires the public hire.voxire.com forms (company
// hiring inquiries and talent applications) into Onion. Both endpoints
// are anonymous (no auth) and available via CORS to *.voxire.com.
//
// Tables are auto-migrated by the repository on startup:
//
//   - voxhire_company_leads
//   - voxhire_talent_applications
//
// CV files land in R2 at the key prefix voxhire/applications/<track-slug>/...
package voxhire

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/voxhire/internal/api"
	"github.com/voxire/onion/internal/modules/voxhire/internal/application"
	internaljobs "github.com/voxire/onion/internal/modules/voxhire/internal/jobs"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

// GetAnonymousRoutes returns paths that bypass JWT auth. Both VoxHire
// public endpoints are anonymous because they accept submissions from
// hire.voxire.com visitors who don't have an Onion account.
func (m *Module) GetAnonymousRoutes() []string {
	anon := []string{}
	anon = append(anon, api.ContactRoutes()...)
	anon = append(anon, api.ApplicationsRoutes()...)
	// Recruitment service: public careers apply + the recruiter-admin
	// surface (self-guarded by the admin JWT, see AdminAuthMiddleware).
	anon = append(anon, api.RecruitAnonymousRoutes()...)
	return anon
}

func (m *Module) RegisterJobs(ctx context.Context, registry *platformjobs.Registry) {
	internaljobs.Register(ctx, registry, m.service)
}

func (m *Module) ProcessPendingCVs(ctx context.Context) error {
	return m.service.ProcessPendingCVs(ctx)
}
