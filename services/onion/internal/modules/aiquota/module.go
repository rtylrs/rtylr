package aiquota

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/aiquota/internal/api"
	"github.com/voxire/onion/internal/modules/aiquota/internal/application"
	hubaiv1 "github.com/voxire/voxie/gen/go/hub/ai/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubaiv1.RegisterAiServiceServer(
		api.NewAiAPI(m.service),
		hubaiv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub ai service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string { return []string{} }

// ReserveQuota / ReconcileUsage expose the quota gate to the echo module with
// primitive signatures, so echo depends on a small local interface rather than
// this module's internal types. ReserveQuota atomically debits an estimate
// before the provider call; ReconcileUsage applies the actual-vs-estimate
// delta afterwards (with retries inside the service).
func (m *Module) ReserveQuota(ctx context.Context, businessID, userID int64) (allowed bool, planSlug string, allowanceTokens, usedTokens, remainingTokens, reservedTokens int64) {
	ok, st, err := m.service.Reserve(ctx, businessID, userID)
	if err != nil {
		// Deny rather than trust ok: a tenant scope that could not be resolved
		// must never read as an available allowance.
		return false, "", 0, 0, 0, 0
	}
	return ok, st.PlanSlug, st.AllowanceTokens, st.UsedTokens, st.RemainingTokens, application.ReservationEstimateTokens
}

func (m *Module) ReconcileUsage(ctx context.Context, businessID, userID, actualTokens, reservedTokens int64) {
	m.service.Reconcile(ctx, businessID, userID, actualTokens, reservedTokens)
}
