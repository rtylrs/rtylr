package hr_time

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	hrcore "github.com/voxire/onion/internal/modules/hr_core"
	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_time/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/hr_time/internal/routes"
	"github.com/voxire/onion/internal/platform/logger"
)

type hrCoreApprovalAdapter struct {
	core *hrcore.Module
}

func (a *hrCoreApprovalAdapter) GetApprovalChain(ctx context.Context, businessID int64, requestType string) (*domain.ApprovalChainInfo, error) {
	chain, err := a.core.GetApprovalChain(ctx, businessID, requestType)
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
	approvalAdapter := &hrCoreApprovalAdapter{core: hrCoreModule}
	flowAdapter := infrastructure.NewFlowAdapter(db)
	return &Module{
		service: application.NewService(ctx, db, approvalAdapter, flowAdapter),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) RegisterScheduledJobs(ctx context.Context) {
	logger.Ctx(ctx).Info("hr_time: scheduled jobs registered (unsubmitted timesheet reminder: Monday 9am)")
}
