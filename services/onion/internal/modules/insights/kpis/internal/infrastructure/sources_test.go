package infrastructure_test

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/crm/customers"
	"github.com/voxire/onion/internal/modules/erp/products"
	"github.com/voxire/onion/internal/modules/hr/hr_reports"
	"github.com/voxire/onion/internal/modules/insights/kpis/internal/infrastructure"
	posorders "github.com/voxire/onion/internal/modules/pos/orders"
)

// newRepository wires the real owning modules, so these tests cover the
// owners' reads as well as the summary built from them.
func newRepository(db *sql.DB) *infrastructure.Repository {
	ctx := context.Background()
	return infrastructure.NewRepository(infrastructure.Sources{
		Staff:     hr_reports.NewModule(ctx, db),
		Customers: customers.NewModule(ctx, db),
		Products:  products.NewModule(ctx, db),
		Orders:    posorders.NewModule(ctx, db),
	})
}
