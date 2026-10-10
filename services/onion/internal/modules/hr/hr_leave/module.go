package hr_leave

import (
	"context"
	"database/sql"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_leave/internal/infrastructure"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/api"
	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
	internaljobs "github.com/voxire/onion/internal/modules/hr/hr_leave/internal/jobs"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	platformjobs "github.com/voxire/onion/internal/platform/jobs"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/tenant"
	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
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

// ApprovedLeaveEmployeeIDs returns the set of employee ids with an approved
// leave request covering the given date (YYYY-MM-DD) - used by Flow's
// business-wide "who's in today" presence view. Passing the same date as both
// bounds relies on ListLeaveCalendar's overlap query to mean "covers this one
// day." Authenticated callers are re-resolved to their verified business id;
// the ownerID argument remains for trusted internal callers without auth data.
func (m *Module) ApprovedLeaveEmployeeIDs(ctx context.Context, ownerID int64, date string) (map[int64]bool, error) {
	businessID := ownerID
	if authUser, ok := authPkg.GetAuthUser(ctx); ok && authUser != nil {
		resolved, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
		if err != nil {
			return nil, err
		}
		businessID = resolved
	}
	if businessID <= 0 {
		return nil, domain.ErrLeaveAccessDenied
	}
	requests, err := m.service.GetLeaveCalendar(ctx, businessID, domain.LeaveCalendarParams{DateFrom: date, DateTo: date})
	if err != nil {
		return nil, err
	}
	onLeave := make(map[int64]bool, len(requests))
	for _, r := range requests {
		if r.Status == domain.LeaveStatusApproved {
			onLeave[r.EmployeeID] = true
		}
	}
	return onLeave, nil
}

func (m *Module) RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, _ ...interface{}) {
	serverHttp.RegisterGET[api.GetLeaveCalendarRequest, api.GetLeaveCalendarResponse](
		ctx, router, api.GetLeaveCalendarRoutes(), &api.GetLeaveCalendarHandler{Service: m.service},
	)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	err := hrleavev1.RegisterLeaveServiceServer(
		api.NewLeaveService(m.service),
		hrleavev1.WithMux(mux),
	)
	if err != nil {
		panic(err)
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) RegisterJobs(ctx context.Context, registry *platformjobs.Registry) {
	internaljobs.Register(ctx, registry, m.service)
}

func (m *Module) RunMonthlyAccrual(ctx context.Context) error {
	return m.service.RunMonthlyAccrual(ctx)
}
