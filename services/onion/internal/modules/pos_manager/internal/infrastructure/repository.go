package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_manager/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
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

func (r *Repository) VerifyPIN(ctx context.Context, ownerID int64, pin string) (*domain.VerifyResult, error) {
	staffID, err := r.queries.VerifyManagerPIN(ctx, sqlc.VerifyManagerPINParams{
		OwnerID: ownerID,
		Pin:     pin,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.VerifyResult{Valid: false}, nil
		}
		logger.Ctx(ctx).Error("Failed to verify manager PIN", logger.Error(err))
		return nil, err
	}

	return &domain.VerifyResult{
		Valid:   true,
		StaffID: staffID,
	}, nil
}
