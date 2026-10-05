package api

import (
	"database/sql"

	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/application"
	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/infrastructure"
)

func newPurchaseOrderTestService(db *sql.DB) *application.Service {
	return application.NewService(infrastructure.NewRepository(db))
}
