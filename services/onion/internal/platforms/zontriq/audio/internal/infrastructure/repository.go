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
	"github.com/voxire/onion/internal/platforms/zontriq/audio/internal/domain"
	sqlc "github.com/voxire/onion/internal/platforms/zontriq/audio/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListAudioByUserID(ctx context.Context, userID, limit, offset int32) ([]domain.AudioClip, error) {
	params := sqlc.ListAudioByUserIDParams{
		ZontriqUserID: userID,
		Limit:         limit,
		Offset:        offset,
	}

	results, err := db.NewReadExecutor[domain.AudioClip, sqlc.ZontriqAudio](db.MainDatabase).FetchOrdered(ctx, r.queries.ListAudioByUserID, params)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list audio", logger.Error(err))

		return nil, fmt.Errorf("failed to list audio: %w", err)
	}

	return results, nil
}

func (r *Repository) CountAudioByUserID(ctx context.Context, userID int32) (int64, error) {
	count, err := r.queries.CountAudioByUserID(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count audio", logger.Error(err))

		return 0, fmt.Errorf("failed to count audio: %w", err)
	}

	return count, nil
}

func (r *Repository) GetAudioByID(ctx context.Context, audioID int32) (domain.AudioClip, error) {
	result, err := db.NewReadExecutor[domain.AudioClip, sqlc.ZontriqAudio](db.MainDatabase).Fetch(ctx, r.queries.GetAudioByID, audioID)
	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			return domain.AudioClip{}, domain.ErrAudioNotFound
		}

		return domain.AudioClip{}, fmt.Errorf("failed to get audio: %w", err)
	}

	return result, nil
}

func (r *Repository) CreateAudioClip(ctx context.Context, userID int32, title, script, voiceID, voiceName, provider, state string) (domain.AudioClip, error) {
	params := sqlc.CreateAudioClipParams{
		ZontriqUserID: userID,
		Title:         title,
		Script:        script,
		VoiceID:       voiceID,
		VoiceName:     pgtype.Text{String: voiceName, Valid: voiceName != ""},
		Provider:      provider,
		State:         state,
	}

	result, err := db.NewWriteExecutor[domain.AudioClip, sqlc.ZontriqAudio](db.MainDatabase).Insert(ctx, r.queries.CreateAudioClip, params)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create audio clip", logger.Error(err))

		return domain.AudioClip{}, fmt.Errorf("failed to create audio clip: %w", err)
	}

	return result, nil
}

func (r *Repository) GetQueuedAudio(ctx context.Context) ([]domain.AudioClip, error) {
	results, err := db.NewReadExecutor[domain.AudioClip, sqlc.ZontriqAudio](db.MainDatabase).FetchOrdered(ctx, r.queries.GetQueuedAudio)
	if err != nil {
		return nil, fmt.Errorf("failed to get queued audio: %w", err)
	}

	return results, nil
}

func (r *Repository) UpdateAudioState(ctx context.Context, audioID int32, state, audioURL string, durationSeconds int32, errorMessage string) error {
	params := sqlc.UpdateAudioStateParams{
		ID:              audioID,
		State:           state,
		AudioUrl:        pgtype.Text{String: audioURL, Valid: audioURL != ""},
		DurationSeconds: pgtype.Int4{Int32: durationSeconds, Valid: durationSeconds > 0},
		ErrorMessage:    pgtype.Text{String: errorMessage, Valid: errorMessage != ""},
	}

	return r.queries.UpdateAudioState(ctx, params)
}

func (r *Repository) SetExternalJobID(ctx context.Context, audioID int32, jobID string) error {
	return r.queries.SetExternalJobID(ctx, sqlc.SetExternalJobIDParams{
		ID:            audioID,
		ExternalJobID: pgtype.Text{String: jobID, Valid: jobID != ""},
	})
}

func (r *Repository) GetGeneratingAudio(ctx context.Context) ([]domain.AudioClip, error) {
	results, err := db.NewReadExecutor[domain.AudioClip, sqlc.ZontriqAudio](db.MainDatabase).FetchOrdered(ctx, r.queries.GetGeneratingAudio)
	if err != nil {
		return nil, fmt.Errorf("failed to get generating audio: %w", err)
	}

	return results, nil
}

func (r *Repository) IncrementRetryCount(ctx context.Context, audioID int32, errorMessage string) error {
	return r.queries.IncrementRetryCount(ctx, sqlc.IncrementRetryCountParams{
		ID:           audioID,
		ErrorMessage: pgtype.Text{String: errorMessage, Valid: errorMessage != ""},
	})
}

func (r *Repository) SoftDeleteAudio(ctx context.Context, audioID int32) error {
	err := r.queries.SoftDeleteAudio(ctx, audioID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to soft delete audio", logger.Error(err))

		return fmt.Errorf("failed to soft delete audio: %w", err)
	}

	return nil
}

func (r *Repository) GetZontriqUserID(ctx context.Context, userID int64) (int32, error) {
	zontriqUserID, err := r.queries.GetZontriqUserIDByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get zontriq user ID: %w", err)
	}

	return zontriqUserID, nil
}
