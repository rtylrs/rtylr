package api

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/crm/customers"
	"github.com/voxire/onion/internal/modules/erp/products"
	financereports "github.com/voxire/onion/internal/modules/finance/finance_reports"
	"github.com/voxire/onion/internal/modules/hr/hr_reports"
	"github.com/voxire/onion/internal/modules/insights/internal/application"
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
		Finance: financereports.NewModule(ctx, db),
		POS:     posorders.NewModule(ctx, db),
		CRM:     customers.NewModule(ctx, db),
	}
}
