package testsupport

import (
	"database/sql"

	identityadapter "github.com/voxire/onion/internal/modules/flow/workspaces/adapters/identity"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/organization/scopes"
)

func NewRepository(db *sql.DB) *infrastructure.Repository {
	reader := scopes.NewReader(db)
	return infrastructure.NewRepository(db).WithIdentityReaders(identityadapter.NewIdentityReader(reader), reader.AccountByEmail, reader.BusinessUserByUsername)
}
