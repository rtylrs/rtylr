package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/voxire/onion/internal/platform/database"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platforms/zontriq/automations/internal/domain"
	sqlc "github.com/voxire/onion/internal/platforms/zontriq/automations/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListAutomationsByUserID(ctx context.Context, zontriqUserID int, state *string, pageSize, pageOffset int) ([]domain.Automation, error) {
	var stateFilter pgtype.Text
	if state != nil {
		stateFilter = pgtype.Text{String: *state, Valid: true}
	}

	results, err := db.NewReadExecutor[domain.Automation, sqlc.ListAutomationsByUserIDRow](db.MainDatabase).FetchOrdered(ctx, r.queries.ListAutomationsByUserID, sqlc.ListAutomationsByUserIDParams{
		ZontriqUserID: int32(zontriqUserID),
		State:         stateFilter,
		PageSize:      int32(pageSize),
		PageOffset:    int32(pageOffset),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list automations", logger.Error(err))

		return nil, fmt.Errorf("failed to list automations: %w", err)
	}

	return results, nil
}

func (r *Repository) CountAutomationsByUserID(ctx context.Context, zontriqUserID int, state *string) (int64, error) {
	var stateFilter pgtype.Text
	if state != nil {
		stateFilter = pgtype.Text{String: *state, Valid: true}
	}

	count, err := r.queries.CountAutomationsByUserID(ctx, sqlc.CountAutomationsByUserIDParams{
		ZontriqUserID: int32(zontriqUserID),
		State:         stateFilter,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count automations", logger.Error(err))

		return 0, fmt.Errorf("failed to count automations: %w", err)
	}

	return count, nil
}

func (r *Repository) GetAutomationByID(ctx context.Context, id int) (*domain.AutomationDetail, error) {
	result, err := db.NewReadExecutor[domain.AutomationDetail, sqlc.GetAutomationByIDRow](db.MainDatabase).Fetch(ctx, r.queries.GetAutomationByID, int32(id))
	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			return nil, domain.ErrAutomationNotFound
		}

		return nil, fmt.Errorf("failed to get automation: %w", err)
	}

	return &result, nil
}

func (r *Repository) CreateAutomation(ctx context.Context, zontriqUserID, seriesID int, name, frequency string) (*domain.AutomationRow, error) {
	arg := sqlc.CreateAutomationParams{
		ZontriqUserID: int32(zontriqUserID),
		SeriesID:      int32(seriesID),
		Name:          name,
		Frequency:     frequency,
	}

	result, err := db.NewWriteExecutor[domain.AutomationRow, sqlc.ZontriqAutomation](db.MainDatabase).Insert(ctx, r.queries.CreateAutomation, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create automation", logger.Error(err))

		return nil, fmt.Errorf("failed to create automation: %w", err)
	}

	return &result, nil
}

func (r *Repository) UpdateAutomation(ctx context.Context, id int, name, frequency, state string) (*domain.AutomationUpdatedRow, error) {
	arg := sqlc.UpdateAutomationParams{
		ID:        int32(id),
		Name:      name,
		Frequency: frequency,
		State:     state,
	}

	result, err := db.NewWriteExecutor[domain.AutomationUpdatedRow, sqlc.ZontriqAutomation](db.MainDatabase).Update(ctx, r.queries.UpdateAutomation, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update automation", logger.Error(err))

		return nil, fmt.Errorf("failed to update automation: %w", err)
	}

	return &result, nil
}

func (r *Repository) SoftDeleteAutomation(ctx context.Context, id int) error {
	return r.queries.SoftDeleteAutomation(ctx, int32(id))
}

func (r *Repository) ListAutomationChannels(ctx context.Context, automationID int) ([]domain.Channel, error) {
	results, err := db.NewReadExecutor[domain.Channel, sqlc.ListAutomationChannelsRow](db.MainDatabase).FetchOrdered(ctx, r.queries.ListAutomationChannels, int32(automationID))
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list automation channels", logger.Error(err))

		return nil, fmt.Errorf("failed to list automation channels: %w", err)
	}

	return results, nil
}

func (r *Repository) CreateAutomationChannel(ctx context.Context, automationID, channelID int) error {
	_, err := r.queries.CreateAutomationChannel(ctx, sqlc.CreateAutomationChannelParams{
		AutomationID: int32(automationID),
		ChannelID:    int32(channelID),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create automation channel", logger.Error(err))

		return fmt.Errorf("failed to create automation channel: %w", err)
	}

	return nil
}

func (r *Repository) DeleteAutomationChannels(ctx context.Context, automationID int) error {
	return r.queries.DeleteAutomationChannels(ctx, int32(automationID))
}

func (r *Repository) GetDueAutomations(ctx context.Context) ([]domain.DueAutomation, error) {
	results, err := db.NewReadExecutor[domain.DueAutomation, sqlc.GetDueAutomationsRow](db.MainDatabase).FetchOrdered(ctx, r.queries.GetDueAutomations)
	if err != nil {
		return nil, fmt.Errorf("failed to get due automations: %w", err)
	}

	return results, nil
}

func (r *Repository) CountAutomationsByUserIDForQuota(ctx context.Context, zontriqUserID int) (int64, error) {
	count, err := r.queries.CountAutomationsByUserIDForQuota(ctx, int32(zontriqUserID))
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count automations for quota", logger.Error(err))

		return 0, fmt.Errorf("failed to count automations for quota: %w", err)
	}

	return count, nil
}

func (r *Repository) GetZontriqUserIDByUserID(ctx context.Context, userID int) (int, error) {
	id, err := r.queries.GetZontriqUserIDByUserID(ctx, int64(userID))
	if err != nil {
		return 0, fmt.Errorf("failed to get zontriq user ID: %w", err)
	}

	return int(id), nil
}
