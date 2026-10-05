package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/infrastructure/persistence/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) GetActiveCareers(ctx context.Context) ([]domain.Career, error) {
	rows, err := r.queries.GetAllCareers(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "doesn't exist") {
			return []domain.Career{}, nil
		}
		return nil, err
	}

	careers := make([]domain.Career, 0, len(rows))
	for _, row := range rows {
		careers = append(careers, careerFromRow(row))
	}

	return careers, nil
}
