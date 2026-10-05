package api

import (
	"database/sql"

	"github.com/voxire/onion/internal/modules/finance/finance_reconciliation/internal/application"
	"github.com/voxire/onion/internal/modules/finance/finance_reconciliation/internal/infrastructure"
)

func newReconciliationTestService(db *sql.DB) *application.Service {
	return application.NewService(infrastructure.NewRepository(db))
}
