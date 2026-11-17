package api

import (
	"context"
	"database/sql"
	"time"

	crmsales "github.com/voxire/onion/internal/modules/crm/crm_sales"
	"github.com/voxire/onion/internal/modules/crm/customers"
	"github.com/voxire/onion/internal/modules/erp/products"
	financereports "github.com/voxire/onion/internal/modules/finance/finance_reports"
	flowinsights "github.com/voxire/onion/internal/modules/flow/insights"
	"github.com/voxire/onion/internal/modules/hr/hr_reports"
	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure"
	posorders "github.com/voxire/onion/internal/modules/pos/orders"
)

// testRepository wires the real owning modules, so these tests cover the
// owners' reads as well as how Insights combines them.
func testRepository(db *sql.DB) *infrastructure.Repository {
	ctx := context.Background()
	return infrastructure.NewRepository(db, infrastructure.Sources{
		Finance: financereports.NewModule(ctx, db),
		HR:      hr_reports.NewModule(ctx, db),
		ERP:     products.NewModule(ctx, db),
		POS:     posorders.NewModule(ctx, db),
		CRM:     customers.NewModule(ctx, db),
	})
}

// testDashboardReaders wires the real owners of the dashboard's headline
// figures.
func testDashboardReaders(db *sql.DB) application.DashboardReaders {
	ctx := context.Background()
	return application.DashboardReaders{
		Finance:  financereports.NewModule(ctx, db),
		POS:      posorders.NewModule(ctx, db),
		CRM:      customers.NewModule(ctx, db),
		CRMSales: crmsales.NewModule(ctx, db),
		Flow:     flowHours{flowinsights.NewModule(ctx, db)},
	}
}

// flowHours stands in for insights/adapters, which this package cannot import
// without a cycle.
type flowHours struct{ module *flowinsights.Module }

func (f flowHours) FlowLoggedTime(ctx context.Context, businessID int64, from, to time.Time) (domain.FlowLoggedTime, error) {
	result, err := f.module.FlowLoggedTime(ctx, businessID, from, to)
	if err != nil {
		return domain.FlowLoggedTime{}, err
	}
	projects := make([]domain.FlowProjectTime, 0, len(result.Projects))
	for _, p := range result.Projects {
		projects = append(projects, domain.FlowProjectTime{ProjectID: p.ProjectID, Name: p.Name, Visibility: p.Visibility, Status: p.Status, Minutes: p.Minutes})
	}
	return domain.FlowLoggedTime{TotalMinutes: result.TotalMinutes, Projects: projects}, nil
}
