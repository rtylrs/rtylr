package aiquota

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/aiquota/internal/application"
	"github.com/voxire/onion/internal/modules/aiquota/internal/routes"
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

func (m *Module) GetAnonymousRoutes() []string { return []string{} }

// Authorize / RecordUsage expose the quota checks to the echo module with
// primitive signatures, so echo depends on a small local interface rather than
// this module's internal types. Authorize fails open (allows) on error.
func (m *Module) Authorize(ctx context.Context, businessID, userID int64) (allowed bool, planSlug string, allowanceTokens, usedTokens, remainingTokens int64) {
	ok, st, err := m.service.Authorize(ctx, businessID, userID)
	if err != nil {
		return true, "", 0, 0, 0
	}
	return ok, st.PlanSlug, st.AllowanceTokens, st.UsedTokens, st.RemainingTokens
}

func (m *Module) RecordUsage(ctx context.Context, businessID, userID int64, tokens int) {
	m.service.Record(ctx, businessID, userID, tokens)
}
