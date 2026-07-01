package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/aiquota/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{pool: db}
}

// ResolveOwnerID maps the caller's business_id to the canonical tenant id
// (businesses.owner_id), which is what subscriptions and usage are keyed by, so
// every member of a team resolves to the same pool. Falls back to the caller's
// own user id (owner accounts, or if the lookup fails).
func (r *Repository) ResolveOwnerID(ctx context.Context, businessID, fallbackUserID int64) int64 {
	ownerID, err := tenant.ResolveOwnerID(ctx, businessID, fallbackUserID)
	if err != nil {
		logger.Ctx(ctx).Warn("failed to resolve owner for AI quota",
			logger.Any("business_id", businessID),
			logger.Error(err),
		)
		return fallbackUserID
	}
	return ownerID
}

// GetEntitlement reads the tenant's active plan. When the tenant has no
// subscription row it falls back to the Free plan's allowance.
func (r *Repository) GetEntitlement(ctx context.Context, ownerID int64) (domain.Entitlement, error) {
	var ent domain.Entitlement
	// cross-schema read from subscriptions + plans (owned by the subscriptions module)
	err := r.pool.QueryRowContext(ctx,
		`SELECT p.slug, p.ai_credits_per_month, s.seat_count, s.status
		 FROM subscriptions s JOIN plans p ON p.id = s.plan_id
		 WHERE s.owner_id = ?`, ownerID).
		Scan(&ent.PlanSlug, &ent.AICreditsPerMonth, &ent.SeatCount, &ent.Status)
	if err == nil {
		if ent.SeatCount <= 0 {
			ent.SeatCount = 1
		}
		return ent, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("GetEntitlement failed", logger.Error(err))
		return domain.Entitlement{}, err
	}

	// No subscription — fall back to the Free plan allowance.
	// cross-schema read from plans
	var credits int
	ferr := r.pool.QueryRowContext(ctx,
		`SELECT ai_credits_per_month FROM plans WHERE slug = 'free' AND is_active = 1 LIMIT 1`).
		Scan(&credits)
	if ferr != nil && !errors.Is(ferr, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("GetEntitlement free fallback failed", logger.Error(ferr))
		return domain.Entitlement{}, ferr
	}
	return domain.Entitlement{PlanSlug: "free", AICreditsPerMonth: credits, SeatCount: 1, Status: "none"}, nil
}

func (r *Repository) GetMode(ctx context.Context, ownerID int64) (domain.Mode, error) {
	var mode string
	err := r.pool.QueryRowContext(ctx,
		"SELECT mode FROM ai_quota_settings WHERE owner_id = ?", ownerID).Scan(&mode)
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
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO ai_quota_settings (owner_id, mode) VALUES (?, ?)
		 ON DUPLICATE KEY UPDATE mode = VALUES(mode)`, ownerID, string(mode.Normalize()))
	if err != nil {
		logger.Ctx(ctx).Error("SetMode failed", logger.Error(err))
	}
	return err
}

func (r *Repository) GetUsedTokens(ctx context.Context, ownerID, bucketUserID int64, month time.Time) (int64, error) {
	var used int64
	err := r.pool.QueryRowContext(ctx,
		`SELECT tokens_used FROM ai_usage_monthly
		 WHERE owner_id = ? AND user_id = ? AND billing_month = ?`,
		ownerID, bucketUserID, month).Scan(&used)
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
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO ai_usage_monthly (owner_id, user_id, billing_month, tokens_used, requests)
		 VALUES (?, ?, ?, ?, 1)
		 ON DUPLICATE KEY UPDATE tokens_used = tokens_used + VALUES(tokens_used), requests = requests + 1`,
		ownerID, bucketUserID, month, tokens)
	if err != nil {
		logger.Ctx(ctx).Error("AddUsage failed", logger.Error(err))
	}
	return err
}
