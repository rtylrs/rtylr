package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/aiquota/internal/domain"
	"github.com/voxire/onion/internal/modules/aiquota/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{queries: sqlc.New(db), pool: db}
}

// ResolveOwnerID verifies the business owner id used to scope quota reads and
// token-usage writes. The platform resolver already fails closed; swallowing
// its error and returning fallbackUserID here would reintroduce that fallback
// one layer up, so the error is propagated unchanged and the caller must deny.
func (r *Repository) ResolveOwnerID(ctx context.Context, businessID, fallbackUserID int64) (int64, error) {
	return tenant.ResolveOwnerID(ctx, businessID, fallbackUserID)
}

func (r *Repository) GetMode(ctx context.Context, ownerID int64) (domain.Mode, error) {
	mode, err := r.queries.GetQuotaMode(ctx, ownerID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ModePool, nil
	}
	if err != nil {
		logger.Ctx(ctx).Error("GetMode failed", logger.Error(err))
		return domain.ModePool, err
	}
	return domain.Mode(mode).Normalize(), nil
}

func (r *Repository) SetMode(ctx context.Context, ownerID int64, mode domain.Mode) error {
	err := r.queries.UpsertQuotaMode(ctx, sqlc.UpsertQuotaModeParams{
		OwnerID: ownerID,
		Mode:    sqlc.AiQuotaSettingsMode(mode.Normalize()),
	})
	if err != nil {
		logger.Ctx(ctx).Error("SetMode failed", logger.Error(err))
	}
	return err
}

func (r *Repository) GetUsedTokens(ctx context.Context, ownerID, bucketUserID int64, month time.Time) (int64, error) {
	used, err := r.queries.GetMonthlyTokenUsage(ctx, sqlc.GetMonthlyTokenUsageParams{
		OwnerID:      ownerID,
		UserID:       bucketUserID,
		BillingMonth: month,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		logger.Ctx(ctx).Error("GetUsedTokens failed", logger.Error(err))
		return 0, err
	}
	return used, nil
}

func (r *Repository) AddUsage(ctx context.Context, ownerID, bucketUserID int64, month time.Time, tokens int64) error {
	err := r.queries.AddMonthlyTokenUsage(ctx, sqlc.AddMonthlyTokenUsageParams{
		OwnerID:      ownerID,
		UserID:       bucketUserID,
		BillingMonth: month,
		TokensUsed:   tokens,
	})
	if err != nil {
		logger.Ctx(ctx).Error("AddUsage failed", logger.Error(err))
	}
	return err
}

// Reserve atomically adds the estimate to the usage bucket only when the
// result stays within allowance, returning false when the bucket is exhausted.
func (r *Repository) Reserve(ctx context.Context, ownerID, bucketUserID int64, month time.Time, estimate, allowance int64) (bool, error) {
	if err := r.queries.EnsureUsageRow(ctx, sqlc.EnsureUsageRowParams{
		OwnerID: ownerID, UserID: bucketUserID, BillingMonth: month,
	}); err != nil {
		return false, err
	}
	res, err := r.queries.ReserveMonthlyTokens(ctx, sqlc.ReserveMonthlyTokensParams{
		TokensUsed: estimate,
		OwnerID:    ownerID, UserID: bucketUserID, BillingMonth: month,
		TokensUsed_2: estimate,
		TokensUsed_3: allowance,
	})
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows == 1, nil
}

// Adjust reconciles a prior reservation against actual usage. The delta may
// be negative when the response used fewer tokens than reserved.
func (r *Repository) Adjust(ctx context.Context, ownerID, bucketUserID int64, month time.Time, delta int64) error {
	return r.queries.AdjustMonthlyTokenUsage(ctx, sqlc.AdjustMonthlyTokenUsageParams{
		TokensUsed: delta,
		OwnerID:    ownerID, UserID: bucketUserID, BillingMonth: month,
	})
}
