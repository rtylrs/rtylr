package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/voxire/onion/internal/platform/database"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/domain"
	sqlc "github.com/voxire/onion/internal/platforms/zontriq/users/internal/infrastructure/persistence/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()
	if pool == nil {
		return nil
	}

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) GetZontriqUserByUserID(ctx context.Context, userID int) (*domain.ZontriqUser, error) {
	result, err := db.NewReadExecutor[domain.ZontriqUser, sqlc.GetZontriqUserByUserIDRow](db.MainDatabase).Fetch(ctx, r.queries.GetZontriqUserByUserID, int64(userID))
	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			return nil, domain.ErrZontriqUserNotFound
		}

		return nil, fmt.Errorf("failed to get zontriq user: %w", err)
	}

	return &result, nil
}

func (r *Repository) CreateZontriqUser(ctx context.Context, userID, planID int) (*domain.ZontriqUser, error) {
	arg := sqlc.CreateZontriqUserParams{
		UserID: int64(userID),
		PlanID: int32(planID),
	}

	result, err := db.NewWriteExecutor[domain.ZontriqUser, sqlc.ZontriqUser](db.MainDatabase).Insert(ctx, r.queries.CreateZontriqUser, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create zontriq user", logger.Error(err))

		return nil, fmt.Errorf("failed to create zontriq user: %w", err)
	}

	return &result, nil
}

func (r *Repository) UpdateZontriqUserPlan(ctx context.Context, id, planID int) error {
	arg := sqlc.UpdateZontriqUserPlanParams{
		ID:     int32(id),
		PlanID: int32(planID),
	}

	return r.queries.UpdateZontriqUserPlan(ctx, arg)
}

func (r *Repository) CompleteOnboarding(ctx context.Context, userID int) error {
	return r.queries.CompleteOnboarding(ctx, int64(userID))
}

func (r *Repository) SoftDeleteZontriqUser(ctx context.Context, id int) error {
	return r.queries.SoftDeleteZontriqUser(ctx, int32(id))
}

func (r *Repository) GetUserPlanByUserID(ctx context.Context, userID int) (*domain.Plan, error) {
	result, err := db.NewReadExecutor[domain.Plan, sqlc.ZontriqPlan](db.MainDatabase).Fetch(ctx, r.queries.GetUserPlanByUserID, int64(userID))
	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			return nil, domain.ErrPlanNotFound
		}

		return nil, fmt.Errorf("failed to get user plan: %w", err)
	}

	return &result, nil
}

func (r *Repository) ListPlans(ctx context.Context) ([]domain.Plan, error) {
	results, err := db.NewReadExecutor[domain.Plan, sqlc.ZontriqPlan](db.MainDatabase).FetchOrdered(ctx, r.queries.ListPlans)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list plans", logger.Error(err))

		return nil, fmt.Errorf("failed to list plans: %w", err)
	}

	return results, nil
}

func (r *Repository) GetPlanByID(ctx context.Context, planID int) (*domain.Plan, error) {
	result, err := db.NewReadExecutor[domain.Plan, sqlc.ZontriqPlan](db.MainDatabase).Fetch(ctx, r.queries.GetPlanByID, int32(planID))
	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			return nil, domain.ErrPlanNotFound
		}

		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	return &result, nil
}

func (r *Repository) GetCurrentUsage(ctx context.Context, zontriqUserID int) (*domain.Usage, error) {
	result, err := db.NewReadExecutor[domain.Usage, sqlc.ZontriqUsage](db.MainDatabase).Fetch(ctx, r.queries.GetCurrentUsage, int32(zontriqUserID))
	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			return nil, domain.ErrUsageNotFound
		}

		return nil, fmt.Errorf("failed to get current usage: %w", err)
	}

	return &result, nil
}

func (r *Repository) IncrementVideosGenerated(ctx context.Context, usageID int) error {
	return r.queries.IncrementVideosGenerated(ctx, int32(usageID))
}

func (r *Repository) IncrementVideosPublished(ctx context.Context, usageID int) error {
	return r.queries.IncrementVideosPublished(ctx, int32(usageID))
}

func (r *Repository) CreateUsagePeriod(ctx context.Context, zontriqUserID int, periodStart, periodEnd time.Time) (*domain.Usage, error) {
	arg := sqlc.CreateUsagePeriodParams{
		ZontriqUserID: int32(zontriqUserID),
		PeriodStart:   pgtype.Timestamptz{Time: periodStart, Valid: true},
		PeriodEnd:     pgtype.Timestamptz{Time: periodEnd, Valid: true},
	}

	result, err := db.NewWriteExecutor[domain.Usage, sqlc.ZontriqUsage](db.MainDatabase).Insert(ctx, r.queries.CreateUsagePeriod, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create usage period", logger.Error(err))

		return nil, fmt.Errorf("failed to create usage period: %w", err)
	}

	return &result, nil
}

func (r *Repository) GetPreferences(ctx context.Context, zontriqUserID int) (*domain.Preferences, error) {
	result, err := db.NewReadExecutor[domain.Preferences, sqlc.ZontriqPreference](db.MainDatabase).Fetch(ctx, r.queries.GetPreferences, int32(zontriqUserID))
	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			return nil, domain.ErrPreferencesNotFound
		}

		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}

	return &result, nil
}

func (r *Repository) CreatePreferences(ctx context.Context, prefs domain.Preferences) (*domain.Preferences, error) {
	arg := sqlc.CreatePreferencesParams{
		ZontriqUserID:        int32(prefs.ZontriqUserID),
		DefaultVoice:         pgtype.Text{String: prefs.DefaultVoice, Valid: prefs.DefaultVoice != ""},
		DefaultNiche:         pgtype.Text{String: prefs.DefaultNiche, Valid: prefs.DefaultNiche != ""},
		NotifyVideoReady:     prefs.NotifyVideoReady,
		NotifyPublishSuccess: prefs.NotifyPublishSuccess,
		NotifyPublishFailure: prefs.NotifyPublishFailure,
		NotifyQuotaWarning:   prefs.NotifyQuotaWarning,
		NotifyWeeklyReport:   prefs.NotifyWeeklyReport,
	}

	result, err := db.NewWriteExecutor[domain.Preferences, sqlc.ZontriqPreference](db.MainDatabase).Insert(ctx, r.queries.CreatePreferences, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create preferences", logger.Error(err))

		return nil, fmt.Errorf("failed to create preferences: %w", err)
	}

	return &result, nil
}

func (r *Repository) CreateWaitlistEntry(ctx context.Context, email, source string) error {
	arg := sqlc.CreateWaitlistEntryParams{
		Email:  email,
		Source: pgtype.Text{String: source, Valid: source != ""},
	}

	return r.queries.CreateWaitlistEntry(ctx, arg)
}

func (r *Repository) UpdatePreferences(ctx context.Context, zontriqUserID int, prefs domain.Preferences) (*domain.Preferences, error) {
	arg := sqlc.UpdatePreferencesParams{
		ZontriqUserID:        int32(zontriqUserID),
		DefaultVoice:         pgtype.Text{String: prefs.DefaultVoice, Valid: prefs.DefaultVoice != ""},
		DefaultNiche:         pgtype.Text{String: prefs.DefaultNiche, Valid: prefs.DefaultNiche != ""},
		NotifyVideoReady:     prefs.NotifyVideoReady,
		NotifyPublishSuccess: prefs.NotifyPublishSuccess,
		NotifyPublishFailure: prefs.NotifyPublishFailure,
		NotifyQuotaWarning:   prefs.NotifyQuotaWarning,
		NotifyWeeklyReport:   prefs.NotifyWeeklyReport,
	}

	result, err := db.NewWriteExecutor[domain.Preferences, sqlc.ZontriqPreference](db.MainDatabase).Update(ctx, r.queries.UpdatePreferences, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update preferences", logger.Error(err))

		return nil, fmt.Errorf("failed to update preferences: %w", err)
	}

	return &result, nil
}
