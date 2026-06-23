package hr_overtime

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	hrcore "github.com/voxire/onion/internal/modules/hr_core"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/application"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	internaljobs "github.com/voxire/onion/internal/modules/hr_overtime/internal/jobs"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/routes"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
	"github.com/voxire/onion/internal/platform/logger"
)

type hrCoreApprovalAdapter struct {
	core *hrcore.Module
}

func (a *hrCoreApprovalAdapter) GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*domain.ApprovalChainInfo, error) {
	chain, err := a.core.GetApprovalChain(ctx, ownerID, requestType)
	if err != nil {
		return nil, err
	}
	if chain == nil {
		return nil, nil
	}
	levels := make([]domain.ApprovalLevelInfo, 0, len(chain.Levels))
	for _, lv := range chain.Levels {
		levels = append(levels, domain.ApprovalLevelInfo{
			LevelOrder:         lv.LevelOrder,
			ApproverType:       lv.ApproverType,
			ApproverRoleID:     lv.ApproverRoleID,
			ApproverEmployeeID: lv.ApproverEmployeeID,
		})
	}
	return &domain.ApprovalChainInfo{
		TotalLevels: chain.TotalLevels,
		Levels:      levels,
	}, nil
}

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB, hrCoreModule *hrcore.Module) *Module {
	adapter := &hrCoreApprovalAdapter{core: hrCoreModule}
	return &Module{
		service: application.NewService(ctx, db, adapter),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) RegisterScheduledJobs(ctx context.Context) {
	logger.Ctx(ctx).Info("hr_overtime: scheduled jobs registered (daily overtime detection: midnight)")
}

func (m *Module) RegisterJobs(ctx context.Context, registry *platformjobs.Registry) {
	internaljobs.Register(ctx, registry, m.service)
}

func (m *Module) RunDailyDetection(ctx context.Context) error {
	return m.service.RunDailyDetection(ctx)
}
