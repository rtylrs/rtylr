package application

import (
	"context"
	"database/sql"

	moduleinfra "github.com/voxire/onion/internal/modules/organization/team/internal/infrastructure"
)

func newTestService(ctx context.Context, db *sql.DB) *Service {
	return NewService(ctx, moduleinfra.NewRepository(db))
}
