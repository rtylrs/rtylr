package application

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/voxire/onion/internal/platforms/zontriq/series/internal/domain"
	"github.com/voxire/onion/internal/platforms/zontriq/series/internal/infrastructure"
	"github.com/voxire/onion/internal/platforms/zontriq/series/internal/infrastructure/persistence/sqlc"
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

func (s *Service) ListSeries(ctx context.Context, userID int32, state string, page, pageSize int32) ([]domain.Series, int64, error) {
	offset := (page - 1) * pageSize

	seriesList, err := s.repo.ListSeriesByUserID(ctx, userID, state, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	totalCount, err := s.repo.CountSeriesByUserID(ctx, userID, state)
	if err != nil {
		return nil, 0, err
	}

	return seriesList, totalCount, nil
}

func (s *Service) GetSeries(ctx context.Context, id int32) (domain.SeriesDetail, domain.SeriesContext, error) {
	series, err := s.repo.GetSeriesByID(ctx, id)
	if err != nil {
		return domain.SeriesDetail{}, domain.SeriesContext{}, err
	}

	seriesCtx, err := s.repo.GetSeriesContext(ctx, id)
	if err != nil {
		// Context may not exist yet, return empty context
		return series, domain.SeriesContext{}, nil //nolint:nilerr // context is optional, absence is not an error
	}

	return series, seriesCtx, nil
}

func (s *Service) CreateSeries(ctx context.Context, userID int32, name, description, niche, voice, style, frequency, config, publishTime string) (domain.SeriesCreated, error) {
	configBytes := []byte("{}")
	if config != "" {
		configBytes = []byte(config)
	}

	series, err := s.repo.CreateSeries(ctx, sqlc.CreateSeriesParams{
		ZontriqUserID: userID,
		Name:          name,
		Description:   pgtype.Text{String: description, Valid: true},
		Niche:         niche,
		Voice:         voice,
		Style:         pgtype.Text{String: style, Valid: true},
		Frequency:     frequency,
		Config:        configBytes,
		PublishTime:   pgtype.Text{String: publishTime, Valid: publishTime != ""},
	})
	if err != nil {
		return domain.SeriesCreated{}, err
	}

	_, contextErr := s.repo.CreateSeriesContext(ctx, sqlc.CreateSeriesContextParams{
		SeriesID: int32(series.Id),
		Context:  "",
	})
	if contextErr != nil {
		return domain.SeriesCreated{}, contextErr
	}

	return series, nil
}

func (s *Service) UpdateSeries(ctx context.Context, id int32, name, description, niche, voice, style, frequency, config, publishTime, state string) (domain.SeriesUpdated, error) {
	configBytes := []byte("{}")
	if config != "" {
		configBytes = []byte(config)
	}

	return s.repo.UpdateSeries(ctx, sqlc.UpdateSeriesParams{
		ID:          id,
		Name:        name,
		Description: pgtype.Text{String: description, Valid: true},
		Niche:       niche,
		Voice:       voice,
		Style:       pgtype.Text{String: style, Valid: true},
		Frequency:   frequency,
		Config:      configBytes,
		PublishTime: pgtype.Text{String: publishTime, Valid: publishTime != ""},
		State:       state,
	})
}

func (s *Service) DeleteSeries(ctx context.Context, id int32) error {
	return s.repo.SoftDeleteSeries(ctx, id)
}

func (s *Service) UpdateSeriesContext(ctx context.Context, seriesID int32, contextText string) (domain.SeriesContextUpdated, error) {
	return s.repo.UpdateSeriesContext(ctx, sqlc.UpdateSeriesContextParams{
		SeriesID: seriesID,
		Context:  contextText,
	})
}

func (s *Service) CountSeriesForQuota(ctx context.Context, userID int32) (int64, error) {
	return s.repo.CountSeriesByUserIDForQuota(ctx, userID)
}
