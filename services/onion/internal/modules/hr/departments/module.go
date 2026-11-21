package departments

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/departments/internal/api"
	"github.com/voxire/onion/internal/modules/hr/departments/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/hr/departments/internal/infrastructure"
	hrdepartmentv1 "github.com/voxire/voxie/gen/go/hr/department/v1"
)

type hrCorePermissionAdapter struct {
	core *hrcore.Module
}

func (a *hrCorePermissionAdapter) GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error) {
	return a.core.GetEmployeeIDByUserID(ctx, businessID, userID)
}

func (a *hrCorePermissionAdapter) HasPermission(ctx context.Context, businessID, employeeID int64, permissionKey string) (bool, error) {
	return a.core.HasPermission(ctx, businessID, employeeID, permissionKey)
}

type Module struct {
	service     *application.Service
	permissions api.PermissionProvider
}

func NewModule(ctx context.Context, db *sql.DB, hrCore ...*hrcore.Module) *Module {
	var permissions api.PermissionProvider
	if len(hrCore) > 0 && hrCore[0] != nil {
		permissions = &hrCorePermissionAdapter{core: hrCore[0]}
	}
	return &Module{
		service:     application.NewService(ctx, moduleinfra.NewRepository(db)),
		permissions: permissions,
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrdepartmentv1.RegisterDepartmentServiceServer(
		api.NewDepartmentAPI(m.service, m.permissions),
		hrdepartmentv1.WithMux(mux),
		hrdepartmentv1.WithErrorWriter(api.WriteError),
	); err != nil {
		panic(fmt.Errorf("register hr department service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
