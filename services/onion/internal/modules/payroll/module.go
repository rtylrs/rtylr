package payroll

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	hrcore "github.com/voxire/onion/internal/modules/hr_core"
	"github.com/voxire/onion/internal/modules/payroll/internal/api"
	"github.com/voxire/onion/internal/modules/payroll/internal/application"
	hrpayrollv1 "github.com/voxire/voxie/gen/go/hr/payroll/v1"
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

func NewModule(ctx context.Context, db *sql.DB, hrCoreModule *hrcore.Module) *Module {
	var permissions api.PermissionProvider
	if hrCoreModule != nil {
		permissions = &hrCorePermissionAdapter{core: hrCoreModule}
	}

	return &Module{
		service:     application.NewService(ctx, db),
		permissions: permissions,
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrpayrollv1.RegisterPayrollServiceServer(
		api.NewPayrollAPI(m.service, m.permissions),
		hrpayrollv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register payroll service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
