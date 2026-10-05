package hr_shifts

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/api"
	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/domain"
	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/infrastructure"
)

// hrCorePermissionAdapter exposes hr_core's role lookups to this module. It
// mirrors the adapter hr_payroll uses, so availability writes are gated the same
// way payroll approvals are rather than through a second gate style.
type hrCorePermissionAdapter struct {
	core *hrcore.Module
}

func (a *hrCorePermissionAdapter) GetEmployeeIDByUserID(ctx context.Context, ownerID, userID int64) (int64, error) {
	return a.core.GetEmployeeIDByUserID(ctx, ownerID, userID)
}

func (a *hrCorePermissionAdapter) HasPermission(ctx context.Context, ownerID, employeeID int64, permissionKey string) (bool, error) {
	return a.core.HasPermission(ctx, ownerID, employeeID, permissionKey)
}

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB, hrCoreModule *hrcore.Module) *Module {
	// Declared as the interface, not the concrete pointer: a nil *adapter stored
	// in an interface is not == nil, so the service's nil guard would pass and
	// then dereference it.
	var permissions domain.PermissionProvider
	if hrCoreModule != nil {
		permissions = &hrCorePermissionAdapter{core: hrCoreModule}
	}
	repo := moduleinfra.NewRepository(db)
	return &Module{
		service: application.NewService(ctx, repo, nil, permissions, repo),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
