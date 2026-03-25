package infrastructure

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/voxire/onion/internal/platform/database"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/convert"
	"github.com/voxire/onion/internal/platforms/zontriq/videos/internal/domain"
	sqlc "github.com/voxire/onion/internal/platforms/zontriq/videos/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListVideosByUserID(ctx context.Context, userID int32, seriesID *int32, state *string, limit, offset int32) ([]domain.Video, error) {
	params := sqlc.ListVideosByUserIDParams{
		ZontriqUserID: userID,
		Limit:         limit,
		Offset:        offset,
	}

	if seriesID != nil {
		params.SeriesID = pgtype.Int4{Int32: *seriesID, Valid: true}
	}

	if state != nil {
		params.State = pgtype.Text{String: *state, Valid: true}
	}

	results, err := db.NewReadExecutor[domain.Video, sqlc.ZontriqVideo](db.MainDatabase).FetchOrdered(ctx, r.queries.ListVideosByUserID, params)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list videos", logger.Error(err))

		return nil, fmt.Errorf("failed to list videos: %w", err)
	}

	return results, nil
}

func (r *Repository) CountVideosByUserID(ctx context.Context, userID int32, seriesID *int32, state *string) (int64, error) {
	params := sqlc.CountVideosByUserIDParams{
		ZontriqUserID: userID,
	}

	if seriesID != nil {
		params.SeriesID = pgtype.Int4{Int32: *seriesID, Valid: true}
	}

	if state != nil {
		params.State = pgtype.Text{String: *state, Valid: true}
	}

	count, err := r.queries.CountVideosByUserID(ctx, params)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count videos", logger.Error(err))

		return 0, fmt.Errorf("failed to count videos: %w", err)
	}

	return count, nil
}

func (r *Repository) GetVideoByID(ctx context.Context, videoID int32) (domain.VideoDetail, error) {
	row, err := r.queries.GetVideoByID(ctx, videoID)
	if err != nil {
		return domain.VideoDetail{}, fmt.Errorf("failed to get video: %w", err)
	}

	var video domain.VideoDetail
	if fromErr := video.FromRows(row); fromErr != nil {
		return domain.VideoDetail{}, fmt.Errorf("failed to map video: %w", fromErr)
	}

	return video, nil
}

func (r *Repository) CreateVideo(ctx context.Context, userID int32, seriesID *int32, title, script, state string, config []byte) (domain.VideoDetail, error) {
	params := sqlc.CreateVideoParams{
		ZontriqUserID: userID,
		Title:         title,
		Script:        pgtype.Text{String: script, Valid: script != ""},
		State:         state,
		Config:        config,
	}

	if seriesID != nil {
		params.SeriesID = pgtype.Int4{Int32: *seriesID, Valid: true}
	}

	row, err := r.queries.CreateVideo(ctx, params)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create video", logger.Error(err))
		return domain.VideoDetail{}, fmt.Errorf("failed to create video: %w", err)
	}

	var video domain.VideoDetail
	video.ID = row.ID
	video.ZontriqUserID = row.ZontriqUserID
	video.SeriesID = convert.ToInt32(row.SeriesID)
	video.Title = row.Title
	video.Script = convert.ToString(row.Script)
	video.State = row.State
	video.Status = row.Status
	video.CreatedAt = row.CreatedAt.Time
	video.UpdatedAt = row.UpdatedAt.Time

	return video, nil
}

func (r *Repository) SoftDeleteVideo(ctx context.Context, videoID int32) error {
	err := r.queries.SoftDeleteVideo(ctx, videoID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to soft delete video", logger.Error(err))

		return fmt.Errorf("failed to soft delete video: %w", err)
	}

	return nil
}

func (r *Repository) CreateVideoTarget(ctx context.Context, videoID, channelID int32) error {
	return db.NewWriteExecutor[any, sqlc.ZontriqVideoTarget](db.MainDatabase).Exec(ctx, r.queries.CreateVideoTarget, sqlc.CreateVideoTargetParams{
		VideoID:   videoID,
		ChannelID: channelID,
	})
}

func (r *Repository) ListVideoTargetsByVideoID(ctx context.Context, videoID int32) ([]domain.VideoTarget, error) {
	rows, err := r.queries.ListVideoTargetsByVideoID(ctx, videoID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list video targets", logger.Error(err))
		return nil, fmt.Errorf("failed to list video targets: %w", err)
	}

	targets := make([]domain.VideoTarget, 0, len(rows))
	for i := range rows {
		var t domain.VideoTarget
		if fromErr := t.FromRows(rows[i]); fromErr != nil {
			return nil, fmt.Errorf("failed to map video target: %w", fromErr)
		}
		targets = append(targets, t)
	}

	return targets, nil
}

func (r *Repository) GetActiveJobByVideoID(ctx context.Context, videoID int32) (domain.Job, error) {
	row, err := r.queries.GetActiveJobByVideoID(ctx, videoID)
	if err != nil {
		return domain.Job{}, domain.ErrJobNotFound
	}

	var job domain.Job
	if fromErr := job.FromRows(row); fromErr != nil {
		return domain.Job{}, fmt.Errorf("failed to map job: %w", fromErr)
	}

	return job, nil
}

func (r *Repository) ListPublishLogByVideoID(ctx context.Context, videoID int32) ([]domain.PublishLogEntry, error) {
	rows, err := r.queries.ListPublishLogByVideoID(ctx, videoID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list publish log", logger.Error(err))
		return nil, fmt.Errorf("failed to list publish log: %w", err)
	}

	entries := make([]domain.PublishLogEntry, 0, len(rows))
	for i := range rows {
		var e domain.PublishLogEntry
		if fromErr := e.FromRows(rows[i]); fromErr != nil {
			return nil, fmt.Errorf("failed to map publish log: %w", fromErr)
		}
		entries = append(entries, e)
	}

	return entries, nil
}

func (r *Repository) GetZontriqUserID(ctx context.Context, userID int64) (int32, error) {
	zontriqUserID, err := r.queries.GetZontriqUserIDByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get zontriq user ID: %w", err)
	}

	return zontriqUserID, nil
}
