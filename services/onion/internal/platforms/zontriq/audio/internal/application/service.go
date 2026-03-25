package application

import (
	"context"
	"fmt"

	"github.com/voxire/onion/internal/platforms/zontriq/audio/internal/domain"
	"github.com/voxire/onion/internal/platforms/zontriq/audio/internal/infrastructure"
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

type ListAudioParams struct {
	UserID   int32
	Page     int32
	PageSize int32
}

type ListAudioResult struct {
	Clips      []domain.AudioClip
	TotalCount int64
}

func (s *Service) ListAudio(ctx context.Context, params ListAudioParams) (ListAudioResult, error) {
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	if params.Page <= 0 {
		params.Page = 1
	}

	offset := (params.Page - 1) * params.PageSize

	clips, err := s.repo.ListAudioByUserID(ctx, params.UserID, params.PageSize, offset)
	if err != nil {
		return ListAudioResult{}, fmt.Errorf("failed to list audio: %w", err)
	}

	totalCount, err := s.repo.CountAudioByUserID(ctx, params.UserID)
	if err != nil {
		return ListAudioResult{}, fmt.Errorf("failed to count audio: %w", err)
	}

	return ListAudioResult{
		Clips:      clips,
		TotalCount: totalCount,
	}, nil
}

func (s *Service) GetAudioByID(ctx context.Context, audioID int32) (domain.AudioClip, error) {
	clip, err := s.repo.GetAudioByID(ctx, audioID)
	if err != nil {
		return domain.AudioClip{}, fmt.Errorf("failed to get audio: %w", err)
	}

	return clip, nil
}

func (s *Service) CreateAudioClip(ctx context.Context, userID int32, title, script, voiceID, voiceName, provider string) (domain.AudioClip, error) {
	clip, err := s.repo.CreateAudioClip(ctx, userID, title, script, voiceID, voiceName, provider, "queued")
	if err != nil {
		return domain.AudioClip{}, fmt.Errorf("failed to create audio clip: %w", err)
	}

	return clip, nil
}

func (s *Service) GetQueuedAudio(ctx context.Context) ([]domain.AudioClip, error) {
	return s.repo.GetQueuedAudio(ctx)
}

func (s *Service) UpdateAudioState(ctx context.Context, audioID int32, state, audioURL string, durationSeconds int32, errorMessage string) error {
	return s.repo.UpdateAudioState(ctx, audioID, state, audioURL, durationSeconds, errorMessage)
}

func (s *Service) IncrementRetryCount(ctx context.Context, audioID int32, errorMessage string) error {
	return s.repo.IncrementRetryCount(ctx, audioID, errorMessage)
}

func (s *Service) SetExternalJobID(ctx context.Context, audioID int32, jobID string) error {
	return s.repo.SetExternalJobID(ctx, audioID, jobID)
}

func (s *Service) GetGeneratingAudio(ctx context.Context) ([]domain.AudioClip, error) {
	return s.repo.GetGeneratingAudio(ctx)
}

func (s *Service) DeleteAudio(ctx context.Context, audioID int32) error {
	_, err := s.repo.GetAudioByID(ctx, audioID)
	if err != nil {
		return fmt.Errorf("failed to get audio for deletion: %w", err)
	}

	if err := s.repo.SoftDeleteAudio(ctx, audioID); err != nil {
		return fmt.Errorf("failed to delete audio: %w", err)
	}

	return nil
}
