package infrastructure

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/voxire/onion/internal/platform/database"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platforms/voxire/careers/internal/domain"
	"github.com/voxire/onion/internal/platforms/voxire/careers/internal/infrastructure/persistence/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) GetActiveCareers(ctx context.Context) ([]domain.Career, error) {
	return db.NewReadExecutor[domain.Career, sqlc.GetAllCareersRow](db.MainDatabase).FetchOrdered(ctx, r.queries.GetAllCareers)
}
