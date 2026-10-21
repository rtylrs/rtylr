package api

import (
	"context"
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/customers"
	"github.com/voxire/onion/internal/modules/erp/products"
	"github.com/voxire/onion/internal/modules/hr/hr_reports"
	"github.com/voxire/onion/internal/modules/insights/kpis/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/insights/kpis/internal/infrastructure"
	posorders "github.com/voxire/onion/internal/modules/pos/orders"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
)

func TestDirectAuthorizationMatrix(t *testing.T) {
	testutil.RunDirectAuthorizationCases(t, []testutil.DirectAuthorizationCase{
		{Name: "hub dashboard view", App: authorization.AppHub, Permission: "hub:dashboard:view", Call: func(ctx context.Context) (any, error) {
			response, err := NewDashboardAPI(application.NewService(ctx, kpiRepository(testutil.DirectAuthorizationDB(ctx)))).GetKPIs(ctx, nil)
			return response, err
		}},
	})
}

func kpiRepository(db *sql.DB) *moduleinfra.Repository {
	ctx := context.Background()
	return moduleinfra.NewRepository(moduleinfra.Sources{
		Staff:     hr_reports.NewModule(ctx, db),
		Customers: customers.NewModule(ctx, db),
		Products:  products.NewModule(ctx, db),
		Orders:    posorders.NewModule(ctx, db),
	})
}
