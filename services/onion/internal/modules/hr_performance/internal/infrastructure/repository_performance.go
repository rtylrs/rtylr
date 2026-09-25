package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func (r *Repository) GetPerformanceReview(ctx context.Context, businessID, id int64) (*domain.PerformanceReviewRecord, error) {
	row, err := r.queries.GetPerformanceReview(ctx, sqlc.GetPerformanceReviewParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReviewNotFound
		}
		return nil, err
	}
	record := reviewFromRow(row)
	return &record, nil
}

func (r *Repository) ListPerformanceReviews(ctx context.Context, businessID, cycleID int64, limit, offset int32) ([]domain.PerformanceReviewRecord, int64, error) {
	rows, err := r.queries.ListPerformanceReviewsByCycle(ctx, sqlc.ListPerformanceReviewsByCycleParams{
		BusinessID:    businessID,
		ReviewCycleID: cycleID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := r.queries.CountPerformanceReviewsByCycle(ctx, sqlc.CountPerformanceReviewsByCycleParams{
		BusinessID:    businessID,
		ReviewCycleID: cycleID,
	})
	if err != nil {
		return nil, 0, err
	}

	records := make([]domain.PerformanceReviewRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, reviewFromRow(row))
	}

	return records, total, nil
}

func (r *Repository) ListPerformanceReviewsByEmployee(ctx context.Context, businessID, cycleID, employeeID int64, limit, offset int32) ([]domain.PerformanceReviewRecord, int64, error) {
	rows, err := r.queries.ListPerformanceReviewsByCycleAndEmployee(ctx, sqlc.ListPerformanceReviewsByCycleAndEmployeeParams{
		BusinessID:    businessID,
		ReviewCycleID: cycleID,
		EmployeeID:    employeeID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := r.queries.CountPerformanceReviewsByCycleAndEmployee(ctx, sqlc.CountPerformanceReviewsByCycleAndEmployeeParams{
		BusinessID:    businessID,
		ReviewCycleID: cycleID,
		EmployeeID:    employeeID,
	})
	if err != nil {
		return nil, 0, err
	}

	records := make([]domain.PerformanceReviewRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, reviewFromRow(row))
	}

	return records, total, nil
}

func (r *Repository) UpdateReviewSubmission(ctx context.Context, businessID, id int64, rating, narrative string) error {
	return r.queries.UpdateReviewSubmission(ctx, sqlc.UpdateReviewSubmissionParams{
		OverallRating: sqlutil.NullString(rating),
		Narrative:     sqlutil.NullString(narrative),
		BusinessID:    businessID,
		ID:            id,
	})
}

func (r *Repository) UpdateReviewStatus(ctx context.Context, businessID, id int64, status string) error {
	return r.queries.UpdateReviewStatus(ctx, sqlc.UpdateReviewStatusParams{
		Status:     status,
		BusinessID: businessID,
		ID:         id,
	})
}

func (r *Repository) ShareReview(ctx context.Context, businessID, id int64) error {
	return r.queries.ShareReview(ctx, sqlc.ShareReviewParams{
		BusinessID: businessID,
		ID:         id,
	})
}

// ===== Goals =====

func (r *Repository) CreateGoal(ctx context.Context, businessID int64, params domain.CreateGoalParams) (*domain.ReviewGoalRecord, error) {
	if err := r.ensureGoalReferences(ctx, businessID, params.EmployeeID, params.ReviewCycleID); err != nil {
		return nil, err
	}

	result, err := r.queries.CreateReviewGoal(ctx, sqlc.CreateReviewGoalParams{
		OwnerID:         businessID,
		BusinessID:      businessID,
		EmployeeID:      params.EmployeeID,
		ReviewCycleID:   params.ReviewCycleID,
		Title:           params.Title,
		Description:     sqlutil.NullString(params.Description),
		Weight:          nullFloat64String(params.Weight),
		SuccessCriteria: sqlutil.NullString(params.SuccessCriteria),
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetGoal(ctx, businessID, id)
}

func (r *Repository) GetGoal(ctx context.Context, businessID, id int64) (*domain.ReviewGoalRecord, error) {
	row, err := r.queries.GetReviewGoal(ctx, sqlc.GetReviewGoalParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrGoalNotFound
		}
		return nil, err
	}
	record := goalFromRow(row)
	return &record, nil
}

func (r *Repository) UpdateGoal(ctx context.Context, businessID, id int64, params domain.UpdateGoalParams) error {
	return r.queries.UpdateReviewGoal(ctx, sqlc.UpdateReviewGoalParams{
		Title:           params.Title,
		Description:     sqlutil.NullString(params.Description),
		Weight:          nullFloat64String(params.Weight),
		SuccessCriteria: sqlutil.NullString(params.SuccessCriteria),
		Status:          params.Status,
		BusinessID:      businessID,
		ID:              id,
	})
}

func (r *Repository) ListGoals(ctx context.Context, businessID, employeeID, cycleID int64, limit, offset int32) ([]domain.ReviewGoalRecord, int64, error) {
	limit, offset = normalizeListPage(limit, offset)
	if employeeID < 0 {
		employeeID = 0
	}
	if cycleID < 0 {
		cycleID = 0
	}
	rows, err := r.queries.ListReviewGoals(ctx, sqlc.ListReviewGoalsParams{
		BusinessID:    businessID,
		EmployeeID:    employeeID,
		ReviewCycleID: cycleID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := r.queries.CountReviewGoals(ctx, sqlc.CountReviewGoalsParams{
		BusinessID:    businessID,
		EmployeeID:    employeeID,
		ReviewCycleID: cycleID,
	})
	if err != nil {
		return nil, 0, err
	}

	records := make([]domain.ReviewGoalRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, goalFromRow(row))
	}

	return records, total, nil
}

// ===== Employees (cross-schema read) =====

func (r *Repository) ListActiveEmployeeIDs(ctx context.Context, businessID int64) ([]int64, error) {
	// cross-schema read from hr_core — active employee IDs for performance review initialisation
	return r.queries.ListActiveEmployeeIDs(ctx, businessID)
}
