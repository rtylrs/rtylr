package hr_payroll

import (
	"context"
	"database/sql"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/infrastructure"

	"github.com/go-chi/chi/v5"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/api"
	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/domain"
	internaljobs "github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/jobs"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
	"github.com/voxire/onion/internal/platform/labourcost"
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

func (a *hrCoreApprovalAdapter) GetEmployeeIDByUserID(ctx context.Context, ownerID, userID int64) (int64, error) {
	return a.core.GetEmployeeIDByUserID(ctx, ownerID, userID)
}

func (a *hrCoreApprovalAdapter) HasPermission(ctx context.Context, ownerID, employeeID int64, permissionKey string) (bool, error) {
	return a.core.HasPermission(ctx, ownerID, employeeID, permissionKey)
}

func (a *hrCoreApprovalAdapter) HasRole(ctx context.Context, ownerID, employeeID, roleID int64) (bool, error) {
	return a.core.HasRole(ctx, ownerID, employeeID, roleID)
}

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB, hrCoreModule *hrcore.Module) *Module {
	adapter := &hrCoreApprovalAdapter{core: hrCoreModule}
	return &Module{
		service: application.NewService(ctx, moduleinfra.NewRepository(db), adapter, adapter),
	}
}

// SetFinancePoster connects payroll exports to the central finance journal.
// Keeping the dependency behind the domain interface prevents payroll from
// depending on journal persistence details.
func (m *Module) SetFinancePoster(p domain.FinanceJournalPoster) {
	m.service.SetFinancePoster(p)
}

// LabourCostProvider exposes the HR-owned aggregate port to consumers without
// exposing the payroll service or compensation details.
func (m *Module) LabourCostProvider() labourcost.Provider {
	return m.service
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) RegisterJobs(ctx context.Context, registry *platformjobs.Registry) {
	internaljobs.Register(ctx, registry, m.service)
}

func (m *Module) NotifyPayrollPeriodOpen(ctx context.Context) error {
	return m.service.NotifyPayrollPeriodOpen(ctx)
}
