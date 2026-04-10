package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/platform/database/dbclient"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}
