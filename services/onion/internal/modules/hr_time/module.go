package hr_time

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	hrcore "github.com/voxire/onion/internal/modules/hr_core"
	"github.com/voxire/onion/internal/modules/hr_time/internal/api"
	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	internaljobs "github.com/voxire/onion/internal/modules/hr_time/internal/jobs"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
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

func (a *hrCoreApprovalAdapter) GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error) {
	return a.core.GetEmployeeIDByUserID(ctx, businessID, userID)
}

func (a *hrCoreApprovalAdapter) GetEmployeeBusinessID(ctx context.Context, businessID, employeeID int64) (int64, error) {
	return a.core.GetEmployeeBusinessID(ctx, businessID, employeeID)
}

func (a *hrCoreApprovalAdapter) HasPermission(ctx context.Context, businessID, employeeID int64, permissionKey string) (bool, error) {
	return a.core.HasPermission(ctx, businessID, employeeID, permissionKey)
}

func (a *hrCoreApprovalAdapter) HasRole(ctx context.Context, businessID, employeeID, roleID int64) (bool, error) {
	return a.core.HasRole(ctx, businessID, employeeID, roleID)
}

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB, hrCoreModule *hrcore.Module) *Module {
	var approvalProvider domain.ApprovalChainProvider
	var permissions domain.PermissionProvider
	if hrCoreModule != nil {
		adapter := &hrCoreApprovalAdapter{core: hrCoreModule}
		approvalProvider = adapter
		permissions = adapter
	}
	return &Module{
		service: application.NewService(ctx, db, approvalProvider, permissions),
	}
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

func (m *Module) RunUnsubmittedTimesheetReminder(ctx context.Context) error {
	return m.service.RunUnsubmittedTimesheetReminder(ctx)
}
