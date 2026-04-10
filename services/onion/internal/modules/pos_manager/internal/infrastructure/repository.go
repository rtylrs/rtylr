package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) VerifyPIN(ctx context.Context, ownerID int64, pin string) (*domain.VerifyResult, error) {
	var staffID int64
	err := r.pool.QueryRowContext(ctx,
		`SELECT staff_id FROM pos_manager_pins WHERE owner_id = ? AND pin = ?`,
		ownerID, pin,
	).Scan(&staffID)
	if err != nil {
		if err == sql.ErrNoRows {
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
