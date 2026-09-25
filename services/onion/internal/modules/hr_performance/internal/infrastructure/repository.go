package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/infrastructure/persistence/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

// ===== Config =====

func (r *Repository) GetPerformanceConfig(ctx context.Context, businessID int64) (*domain.PerformanceConfigRecord, error) {
	row, err := r.queries.GetPerformanceConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPerformanceConfigNotFound
		}
		return nil, err
	}
	record := configFromRow(row)
	return &record, nil
}

func (r *Repository) UpsertPerformanceConfig(ctx context.Context, businessID int64, params domain.UpsertConfigParams) (*domain.PerformanceConfigRecord, error) {
	var labels json.RawMessage
	if params.RatingLabels != "" {
		labels = json.RawMessage(params.RatingLabels)
	}

	_, err := r.queries.UpsertPerformanceConfig(ctx, sqlc.UpsertPerformanceConfigParams{
		OwnerID:             businessID,
		BusinessID:          businessID,
		RatingScaleType:     params.RatingScaleType,
		RatingMin:           params.RatingMin,
		RatingMax:           params.RatingMax,
		RatingLabels:        labels,
		PeerFeedbackEnabled: params.PeerFeedbackEnabled,
	})
	if err != nil {
		return nil, err
	}

	return r.GetPerformanceConfig(ctx, businessID)
}

// ===== Review Cycles =====

func (r *Repository) CreateReviewCycle(ctx context.Context, businessID int64, params domain.CreateReviewCycleParams) (*domain.ReviewCycleRecord, error) {
	periodStart, err := time.Parse("2006-01-02", params.PeriodStart)
	if err != nil {
		return nil, fmt.Errorf("invalid period_start: %w", err)
	}
	periodEnd, err := time.Parse("2006-01-02", params.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid period_end: %w", err)
	}

	result, err := r.queries.CreateReviewCycle(ctx, sqlc.CreateReviewCycleParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		Name:        params.Name,
		CycleType:   params.CycleType,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetReviewCycle(ctx, businessID, id)
}

func (r *Repository) GetReviewCycle(ctx context.Context, businessID, id int64) (*domain.ReviewCycleRecord, error) {
	row, err := r.queries.GetReviewCycle(ctx, sqlc.GetReviewCycleParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReviewCycleNotFound
		}
		return nil, err
	}
	record := cycleFromRow(row)
	return &record, nil
}

func (r *Repository) ListReviewCycles(ctx context.Context, businessID int64, limit, offset int32) ([]domain.ReviewCycleRecord, int64, error) {
	limit, offset = normalizeListPage(limit, offset)
	rows, err := r.queries.ListReviewCycles(ctx, sqlc.ListReviewCyclesParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := r.queries.CountReviewCycles(ctx, businessID)
	if err != nil {
		return nil, 0, err
	}

	records := make([]domain.ReviewCycleRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, cycleFromRow(row))
	}

	return records, total, nil
}

func (r *Repository) UpdateReviewCycleStatus(ctx context.Context, businessID, id int64, status string) error {
	return r.queries.UpdateReviewCycleStatus(ctx, sqlc.UpdateReviewCycleStatusParams{
		Status:     status,
		BusinessID: businessID,
		ID:         id,
	})
}

// dynamic-query: squirrel
func (r *Repository) OpenReviewCycle(ctx context.Context, businessID, id int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	err = qtx.UpdateReviewCycleStatus(ctx, sqlc.UpdateReviewCycleStatusParams{
		Status:     "open",
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// cross-schema read from hr_core — active employee list for review cycle initialisation
	employeeIDs, err := qtx.ListActiveEmployeeIDs(ctx, businessID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Insert all self-reviews in one multi-row INSERT instead of a statement
	// per employee, removing N-1 round-trips inside the transaction.
	if len(employeeIDs) > 0 {
		ib := sq.Insert("hr_performance_reviews").
			Columns("owner_id", "business_id", "employee_id", "review_cycle_id", "reviewer_employee_id", "review_type", "status")
		for _, eid := range employeeIDs {
			ib = ib.Values(businessID, businessID, eid, id, eid, "self", "draft")
		}

		insertStr, insertArgs, buildErr := ib.PlaceholderFormat(sq.Question).ToSql()
		if buildErr != nil {
			_ = tx.Rollback()
			return buildErr
		}

		if _, err = tx.ExecContext(ctx, insertStr, insertArgs...); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// ===== Reviews =====

func (r *Repository) CreatePerformanceReview(ctx context.Context, businessID, employeeID, cycleID, reviewerID int64, reviewType string) error {
	return r.queries.CreatePerformanceReview(ctx, sqlc.CreatePerformanceReviewParams{
		OwnerID:            businessID,
		BusinessID:         businessID,
		EmployeeID:         employeeID,
		ReviewCycleID:      cycleID,
		ReviewerEmployeeID: reviewerID,
		ReviewType:         reviewType,
	})
}
