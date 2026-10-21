package api

import (
	"database/sql"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/finance/invoices/internal/infrastructure"
)

func newInvoicesServiceForTest(db *sql.DB) *application.Service {
	return application.NewService(moduleinfra.NewRepository(db), nil)
}
