package staff

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr/staff/internal/application"
	"github.com/voxire/onion/internal/platform/orgchart"
)

// orgchartResolver adapts the staff service to the platform contract so Flow can
// ask who reports to a caller without importing this module's infrastructure,
// which docs/sqlc-query-ownership.md forbids.
type orgchartResolver struct {
	service *application.Service
}

func (r *orgchartResolver) ManagedUserIDs(ctx context.Context, managerUserID, businessID int64) ([]int64, error) {
	return r.service.ListManagedUserIDs(ctx, businessID, managerUserID)
}

// NewOrgchartResolver returns the reporting-line provider backed by HR staff and
// department records.
func NewOrgchartResolver(service *application.Service) orgchart.ManagerResolver {
	return &orgchartResolver{service: service}
}

// OrgchartResolver exposes the reporting-line provider to cross-module wiring.
// Only the provider crosses the boundary, not the service, so a consumer cannot
// reach the rest of the staff API through it.
func (m *Module) OrgchartResolver() orgchart.ManagerResolver {
	return NewOrgchartResolver(m.service)
}
