package staff

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/staff/internal/infrastructure"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/staff/internal/api"
	"github.com/voxire/onion/internal/modules/hr/staff/internal/application"
	hrstaffv1 "github.com/voxire/voxie/gen/go/hr/staff/v1"
)

type Module struct {
	service     *application.Service
	permissions api.PermissionProvider
}

type hrCorePermissionAdapter struct {
	core *hrcore.Module
}

func (a *hrCorePermissionAdapter) GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error) {
	return a.core.GetEmployeeIDByUserID(ctx, businessID, userID)
}

func (a *hrCorePermissionAdapter) HasPermission(ctx context.Context, businessID, employeeID int64, permissionKey string) (bool, error) {
	return a.core.HasPermission(ctx, businessID, employeeID, permissionKey)
}

func (a *hrCorePermissionAdapter) IsDirectoryEnabled(ctx context.Context, businessID int64) (bool, error) {
	return a.core.IsDirectoryEnabled(ctx, businessID)
}

func NewModule(ctx context.Context, db *sql.DB, hrCoreModule *hrcore.Module) *Module {
	var permissions api.PermissionProvider
	if hrCoreModule != nil {
		permissions = &hrCorePermissionAdapter{core: hrCoreModule}
	}

	return &Module{
		service:     application.NewService(ctx, moduleinfra.NewRepository(db)),
		permissions: permissions,
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrstaffv1.RegisterStaffServiceServer(
		api.NewStaffAPI(m.service, m.permissions),
		hrstaffv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register HR staff service: %w", err))
	}
}

// LoadDashboardStats returns headline HR dashboard metrics for cross-module OneKit wiring.
func (m *Module) LoadDashboardStats(ctx context.Context, businessID int64) (totalStaff, onShiftToday, pendingTimeOff, payrollDraft int, err error) {
	stats, err := m.service.GetDashboardStats(ctx, businessID)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return stats.TotalStaff, stats.OnShiftToday, stats.PendingTimeOff, stats.PayrollDraft, nil
}

func (m *Module) EnsureStaffForUser(ctx context.Context, businessID, userID int64, role string) error {
	return m.service.EnsureStaffForUser(ctx, businessID, userID, role)
}

func (m *Module) DeleteStaffForUser(ctx context.Context, businessID, userID int64) error {
	return m.service.DeleteStaffForUser(ctx, businessID, userID)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
