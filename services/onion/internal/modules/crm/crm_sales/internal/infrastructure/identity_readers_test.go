package infrastructure_test

import (
	"database/sql"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/organization/scopes"
)

func configuredRepository(db *sql.DB) *infrastructure.Repository {
	return infrastructure.NewRepository(db).WithMembershipReader(scopes.NewReader(db).IsBusinessUser)
}
