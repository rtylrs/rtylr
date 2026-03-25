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
	"github.com/voxire/onion/internal/platforms/zontriq/channels/internal/domain"
	sqlc "github.com/voxire/onion/internal/platforms/zontriq/channels/internal/infrastructure/persistence/sqlc"
)

type CreateChannelInput struct {
	ZontriqUserID     int32
	Platform          string
	PlatformChannelID string
	ChannelName       string
	ChannelHandle     string
	ChannelAvatarUrl  string
	AccessToken       string
	RefreshToken      string
	TokenExpiresAt    time.Time
}

type UpdateTokensInput struct {
	ChannelID       int32
	AccessToken     string
	RefreshToken    string
	TokenExpiresAt  time.Time
	ConnectionState string
}

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

func (r *Repository) ListChannelsByUserID(ctx context.Context, zontriqUserID int32, platform string) ([]domain.Channel, error) {
	var platformArg pgtype.Text
	if platform != "" {
		platformArg = pgtype.Text{String: platform, Valid: true}
	}

	results, err := db.NewReadExecutor[domain.Channel, sqlc.ZontriqChannel](db.MainDatabase).FetchOrdered(ctx, r.queries.ListChannelsByUserID, sqlc.ListChannelsByUserIDParams{
		ZontriqUserID: zontriqUserID,
		Platform:      platformArg,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list channels", logger.Error(err))

		return nil, fmt.Errorf("failed to list channels: %w", err)
	}

	return results, nil
}

func (r *Repository) GetChannelByID(ctx context.Context, channelID int32) (domain.Channel, error) {
	result, err := db.NewReadExecutor[domain.Channel, sqlc.ZontriqChannel](db.MainDatabase).Fetch(ctx, r.queries.GetChannelByID, channelID)
	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			return domain.Channel{}, domain.ErrChannelNotFound
		}

		return domain.Channel{}, fmt.Errorf("failed to get channel: %w", err)
	}

	return result, nil
}

func (r *Repository) CreateChannel(ctx context.Context, input CreateChannelInput) (domain.Channel, error) {
	result, err := db.NewWriteExecutor[domain.Channel, sqlc.ZontriqChannel](db.MainDatabase).Insert(ctx, r.queries.CreateChannel, sqlc.CreateChannelParams{
		ZontriqUserID:     input.ZontriqUserID,
		Platform:          input.Platform,
		PlatformChannelID: input.PlatformChannelID,
		ChannelName:       input.ChannelName,
		ChannelHandle:     pgtype.Text{String: input.ChannelHandle, Valid: input.ChannelHandle != ""},
		ChannelAvatarUrl:  pgtype.Text{String: input.ChannelAvatarUrl, Valid: input.ChannelAvatarUrl != ""},
		AccessToken:       input.AccessToken,
		RefreshToken:      pgtype.Text{String: input.RefreshToken, Valid: input.RefreshToken != ""},
		TokenExpiresAt:    pgtype.Timestamptz{Time: input.TokenExpiresAt, Valid: !input.TokenExpiresAt.IsZero()},
		ConnectionState:   "connected",
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create channel", logger.Error(err))

		return domain.Channel{}, fmt.Errorf("failed to create channel: %w", err)
	}

	return result, nil
}

func (r *Repository) UpdateChannelTokens(ctx context.Context, input UpdateTokensInput) (domain.Channel, error) {
	result, err := db.NewWriteExecutor[domain.Channel, sqlc.ZontriqChannel](db.MainDatabase).Update(ctx, r.queries.UpdateChannelTokens, sqlc.UpdateChannelTokensParams{
		ID:              input.ChannelID,
		AccessToken:     input.AccessToken,
		RefreshToken:    pgtype.Text{String: input.RefreshToken, Valid: input.RefreshToken != ""},
		TokenExpiresAt:  pgtype.Timestamptz{Time: input.TokenExpiresAt, Valid: !input.TokenExpiresAt.IsZero()},
		ConnectionState: input.ConnectionState,
	})
	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			return domain.Channel{}, domain.ErrChannelNotFound
		}

		logger.Ctx(ctx).Error("Failed to update channel tokens", logger.Error(err))

		return domain.Channel{}, fmt.Errorf("failed to update channel tokens: %w", err)
	}

	return result, nil
}

func (r *Repository) SoftDeleteChannel(ctx context.Context, channelID int32) error {
	err := r.queries.SoftDeleteChannel(ctx, channelID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to soft delete channel", logger.Error(err))

		return fmt.Errorf("failed to soft delete channel: %w", err)
	}

	return nil
}

func (r *Repository) CountChannelsByUserID(ctx context.Context, zontriqUserID int32) (int64, error) {
	count, err := r.queries.CountChannelsByUserID(ctx, zontriqUserID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count channels", logger.Error(err))

		return 0, fmt.Errorf("failed to count channels: %w", err)
	}

	return count, nil
}

func (r *Repository) GetZontriqUserID(ctx context.Context, userID int64) (int32, error) {
	zontriqUserID, err := r.queries.GetZontriqUserIDByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get zontriq user ID: %w", err)
	}

	return zontriqUserID, nil
}
