package api

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/erp/products"
	"github.com/voxire/onion/internal/modules/organization/addons"
	"github.com/voxire/onion/internal/modules/pos/orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos/orders/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/pos/settings"
)

func newOrderTestService(ctx context.Context, db *sql.DB) *application.Service {
	repository := infrastructure.NewRepository(db).
		WithSalesProductReader(products.NewModule(ctx, db).SalesProduct).
		WithOnlineOrderingReader(addons.NewModule(ctx, db).OnlineOrderingEnabled)
	return application.NewService(repository, settings.NewService(ctx, db))
}
