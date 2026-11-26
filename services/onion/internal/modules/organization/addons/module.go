package addons

import (
	"context"
	"database/sql"

	moduleinfra "github.com/voxire/onion/internal/modules/organization/addons/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/organization/addons/internal/application"
)

type Module struct {
	service    *application.Service
	repository *moduleinfra.Repository
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repository := moduleinfra.NewRepository(db)
	return &Module{service: application.NewService(ctx, repository), repository: repository}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

// GetService exposes the service for other modules to check addon status.
func (m *Module) GetService() *application.Service {
	return m.service
}

func (m *Module) OnlineOrderingEnabled(ctx context.Context, businessID int64) (bool, error) {
	return m.repository.IsEnabled(ctx, businessID, "online_ordering")
}

// PublicMenuEnabled reports whether a business may serve its public digital menu catalog
// (true when either the Digital Menu addon or Online Ordering addon is enabled).
func (m *Module) PublicMenuEnabled(ctx context.Context, businessID int64) (bool, error) {
	menuOn, err := m.repository.IsEnabled(ctx, businessID, "menu")
	if err != nil {
		return false, err
	}
	if menuOn {
		return true, nil
	}
	return m.repository.IsEnabled(ctx, businessID, "online_ordering")
}
