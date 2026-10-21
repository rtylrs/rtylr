package infrastructure_test

import (
	"database/sql"

	"github.com/voxire/onion/internal/modules/organization/scopes"
	"github.com/voxire/onion/internal/modules/pos/memberlocations/internal/infrastructure"
)

func configuredRepository(db *sql.DB) *infrastructure.Repository {
	return infrastructure.NewRepository(db).WithIdentityReaders(scopes.NewReader(db).BusinessUserName, scopes.NewReader(db).IsBusinessUser)
}
