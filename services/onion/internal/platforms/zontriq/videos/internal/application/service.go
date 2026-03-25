package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/platforms/zontriq/videos/internal/domain"
	"github.com/voxire/onion/internal/platforms/zontriq/videos/internal/infrastructure"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo: repo,
	}
}

func (s *Service) GetZontriqUserID(ctx context.Context, userID int64) (int32, error) {
	zontriqUserID, err := s.repo.GetZontriqUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get zontriq user ID: %w", err)
	}

	return zontriqUserID, nil
}

type ListVideosParams struct {
	UserID   int32
	SeriesID *int32
	State    *string
	Page     int32
	PageSize int32
}

type ListVideosResult struct {
	Videos     []domain.Video
	TotalCount int64
}

func (s *Service) ListVideos(ctx context.Context, params ListVideosParams) (ListVideosResult, error) {
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	if params.Page <= 0 {
		params.Page = 1
	}

	offset := (params.Page - 1) * params.PageSize

	videos, err := s.repo.ListVideosByUserID(ctx, params.UserID, params.SeriesID, params.State, params.PageSize, offset)
	if err != nil {
		return ListVideosResult{}, fmt.Errorf("failed to list videos: %w", err)
	}

	totalCount, err := s.repo.CountVideosByUserID(ctx, params.UserID, params.SeriesID, params.State)
	if err != nil {
		return ListVideosResult{}, fmt.Errorf("failed to count videos: %w", err)
	}

	return ListVideosResult{
		Videos:     videos,
		TotalCount: totalCount,
	}, nil
}

type GetVideoResult struct {
	Video      domain.VideoDetail
	Targets    []domain.VideoTarget
	ActiveJob  *domain.Job
	PublishLog []domain.PublishLogEntry
}

func (s *Service) GetVideoByID(ctx context.Context, videoID int32) (GetVideoResult, error) {
	video, err := s.repo.GetVideoByID(ctx, videoID)
	if err != nil {
		return GetVideoResult{}, fmt.Errorf("failed to get video: %w", err)
	}

	targets, err := s.repo.ListVideoTargetsByVideoID(ctx, videoID)
	if err != nil {
		return GetVideoResult{}, fmt.Errorf("failed to get video targets: %w", err)
	}

	var activeJob *domain.Job

	job, err := s.repo.GetActiveJobByVideoID(ctx, videoID)
	if err != nil {
		if !errors.Is(err, domain.ErrJobNotFound) {
			return GetVideoResult{}, fmt.Errorf("failed to get active job: %w", err)
		}
	} else {
		activeJob = &job
	}

	publishLog, err := s.repo.ListPublishLogByVideoID(ctx, videoID)
	if err != nil {
		return GetVideoResult{}, fmt.Errorf("failed to get publish log: %w", err)
	}

	return GetVideoResult{
		Video:      video,
		Targets:    targets,
		ActiveJob:  activeJob,
		PublishLog: publishLog,
	}, nil
}

func (s *Service) CreateVideo(ctx context.Context, userID int32, seriesID *int32, title, script string, config []byte, channelIDs []int64) (domain.VideoDetail, error) {
	video, err := s.repo.CreateVideo(ctx, userID, seriesID, title, script, "draft", config)
	if err != nil {
		return domain.VideoDetail{}, fmt.Errorf("failed to create video: %w", err)
	}

	for _, channelID := range channelIDs {
		if err := s.repo.CreateVideoTarget(ctx, video.ID, int32(channelID)); err != nil {
			return domain.VideoDetail{}, fmt.Errorf("failed to create video target for channel %d: %w", channelID, err)
		}
	}

	return video, nil
}

func (s *Service) DeleteVideo(ctx context.Context, videoID int32) error {
	_, err := s.repo.GetVideoByID(ctx, videoID)
	if err != nil {
		return fmt.Errorf("failed to get video for deletion: %w", err)
	}

	if err := s.repo.SoftDeleteVideo(ctx, videoID); err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}

	return nil
}
