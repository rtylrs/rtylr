package application

import (
	"database/sql"

	moduleinfra "github.com/voxire/onion/internal/modules/organization/user/internal/infrastructure"
)

func newTestService(db *sql.DB, googleClientID string) *Service {
	return NewService(moduleinfra.NewRepository(db), googleClientID)
}
