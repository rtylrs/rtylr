package infrastructure

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/voxire/onion/internal/platform/database"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platforms/zontriq/series/internal/domain"
	"github.com/voxire/onion/internal/platforms/zontriq/series/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListSeriesByUserID(ctx context.Context, userID int32, state string, pageSize, pageOffset int32) ([]domain.Series, error) {
	stateParam := pgtype.Text{}
	if state != "" {
		stateParam = pgtype.Text{String: state, Valid: true}
	}

	return db.NewReadExecutor[domain.Series, sqlc.ZontriqSeries](db.MainDatabase).FetchOrdered(ctx, r.queries.ListSeriesByUserID, sqlc.ListSeriesByUserIDParams{
		ZontriqUserID: userID,
		State:         stateParam,
		PageSize:      pageSize,
		PageOffset:    pageOffset,
	})
}

func (r *Repository) CountSeriesByUserID(ctx context.Context, userID int32, state string) (int64, error) {
	stateParam := pgtype.Text{}
	if state != "" {
		stateParam = pgtype.Text{String: state, Valid: true}
	}

	count, err := r.queries.CountSeriesByUserID(ctx, sqlc.CountSeriesByUserIDParams{
		ZontriqUserID: userID,
		State:         stateParam,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count series", logger.Error(err))

		return 0, err
	}

	return count, nil
}

func (r *Repository) GetSeriesByID(ctx context.Context, id int32) (domain.SeriesDetail, error) {
	series, err := db.NewReadExecutor[domain.SeriesDetail, sqlc.ZontriqSeries](db.MainDatabase).Fetch(ctx, r.queries.GetSeriesByID, id)
	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			return domain.SeriesDetail{}, domain.ErrSeriesNotFound
		}

		return domain.SeriesDetail{}, err
	}

	return series, nil
}

func (r *Repository) CreateSeries(ctx context.Context, params sqlc.CreateSeriesParams) (domain.SeriesCreated, error) {
	return db.NewWriteExecutor[domain.SeriesCreated, sqlc.ZontriqSeries](db.MainDatabase).Insert(ctx, r.queries.CreateSeries, params)
}

func (r *Repository) UpdateSeries(ctx context.Context, params sqlc.UpdateSeriesParams) (domain.SeriesUpdated, error) {
	return db.NewWriteExecutor[domain.SeriesUpdated, sqlc.ZontriqSeries](db.MainDatabase).Update(ctx, r.queries.UpdateSeries, params)
}

func (r *Repository) SoftDeleteSeries(ctx context.Context, id int32) error {
	return db.NewWriteExecutor[any, any](db.MainDatabase).Exec(ctx, r.queries.SoftDeleteSeries, id)
}

func (r *Repository) GetSeriesContext(ctx context.Context, seriesID int32) (domain.SeriesContext, error) {
	seriesCtx, err := db.NewReadExecutor[domain.SeriesContext, sqlc.ZontriqSeriesContext](db.MainDatabase).Fetch(ctx, r.queries.GetSeriesContext, seriesID)
	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			return domain.SeriesContext{}, domain.ErrContextNotFound
		}

		return domain.SeriesContext{}, err
	}

	return seriesCtx, nil
}

func (r *Repository) CreateSeriesContext(ctx context.Context, params sqlc.CreateSeriesContextParams) (domain.SeriesContextCreated, error) {
	return db.NewWriteExecutor[domain.SeriesContextCreated, sqlc.ZontriqSeriesContext](db.MainDatabase).Insert(ctx, r.queries.CreateSeriesContext, params)
}

func (r *Repository) UpdateSeriesContext(ctx context.Context, params sqlc.UpdateSeriesContextParams) (domain.SeriesContextUpdated, error) {
	return db.NewWriteExecutor[domain.SeriesContextUpdated, sqlc.ZontriqSeriesContext](db.MainDatabase).Update(ctx, r.queries.UpdateSeriesContext, params)
}

func (r *Repository) CountSeriesByUserIDForQuota(ctx context.Context, userID int32) (int64, error) {
	count, err := r.queries.CountSeriesByUserIDForQuota(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count series for quota", logger.Error(err))

		return 0, err
	}

	return count, nil
}
