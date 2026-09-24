package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

func (r *Repository) EmployeeBelongsToBusiness(ctx context.Context, businessID, employeeID int64) (bool, error) {
	if r.pool == nil || businessID <= 0 || employeeID <= 0 {
		return false, nil
	}

	var exists int
	err := r.pool.QueryRowContext(ctx, `
		SELECT 1
		FROM hr_employees
		WHERE id = ? AND business_id = ? AND status <> 'terminated'
		LIMIT 1`, employeeID, businessID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *Repository) GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error) {
	if r.pool == nil || businessID <= 0 || userID <= 0 {
		return 0, domain.ErrEmployeeNotFound
	}

	var employeeID int64
	err := r.pool.QueryRowContext(ctx, `
		SELECT id
		FROM hr_employees
		WHERE business_id = ? AND user_id = ? AND status <> 'terminated'
		ORDER BY id ASC
		LIMIT 1`, businessID, userID).Scan(&employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrEmployeeNotFound
	}
	if err != nil {
		return 0, err
	}
	return employeeID, nil
}

func (r *Repository) ensureGoalReferences(ctx context.Context, businessID, employeeID, cycleID int64) error {
	belongs, err := r.EmployeeBelongsToBusiness(ctx, businessID, employeeID)
	if err != nil {
		return err
	}
	if !belongs {
		return domain.ErrEmployeeNotFound
	}

	var cycleExists int
	err = r.pool.QueryRowContext(ctx, `
		SELECT 1
		FROM hr_performance_review_cycles
		WHERE id = ? AND business_id = ?
		LIMIT 1`, cycleID, businessID).Scan(&cycleExists)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrReviewCycleNotFound
	}
	if err != nil {
		return err
	}
	return nil
}

// ===== Row conversion helpers =====

func normalizeListPage(limit, offset int32) (int32, int32) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func configFromRow(row sqlc.HrPerformanceConfig) domain.PerformanceConfigRecord {
	labels := ""
	if row.RatingLabels != nil {
		labels = string(row.RatingLabels)
	}
	return domain.PerformanceConfigRecord{
		ID:                  row.ID,
		OwnerID:             row.OwnerID,
		BusinessID:          row.BusinessID,
		RatingScaleType:     row.RatingScaleType,
		RatingMin:           row.RatingMin,
		RatingMax:           row.RatingMax,
		RatingLabels:        labels,
		PeerFeedbackEnabled: row.PeerFeedbackEnabled,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}

func cycleFromRow(row sqlc.HrPerformanceReviewCycle) domain.ReviewCycleRecord {
	return domain.ReviewCycleRecord{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		Name:        row.Name,
		CycleType:   row.CycleType,
		PeriodStart: row.PeriodStart.Format("2006-01-02"),
		PeriodEnd:   row.PeriodEnd.Format("2006-01-02"),
		Status:      row.Status,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func reviewFromRow(row sqlc.HrPerformanceReview) domain.PerformanceReviewRecord {
	return domain.PerformanceReviewRecord{
		ID:                   row.ID,
		OwnerID:              row.OwnerID,
		BusinessID:           row.BusinessID,
		EmployeeID:           row.EmployeeID,
		ReviewCycleID:        row.ReviewCycleID,
		ReviewerEmployeeID:   row.ReviewerEmployeeID,
		ReviewType:           row.ReviewType,
		OverallRating:        row.OverallRating.String,
		Narrative:            row.Narrative.String,
		Status:               row.Status,
		SharedWithEmployeeAt: row.SharedWithEmployeeAt.Time,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}

func goalFromRow(row sqlc.HrPerformanceReviewGoal) domain.ReviewGoalRecord {
	weight := 0.0
	if row.Weight.Valid {
		parsed, err := strconv.ParseFloat(row.Weight.String, 64)
		if err != nil {
			logger.Global().Warn("failed to parse weight decimal from database",
				logger.String("value", row.Weight.String),
				logger.Error(err),
			)
		} else {
			weight = parsed
		}
	}
	return domain.ReviewGoalRecord{
		ID:              row.ID,
		OwnerID:         row.OwnerID,
		BusinessID:      row.BusinessID,
		EmployeeID:      row.EmployeeID,
		ReviewCycleID:   row.ReviewCycleID,
		Title:           row.Title,
		Description:     row.Description.String,
		Weight:          weight,
		SuccessCriteria: row.SuccessCriteria.String,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func nullFloat64String(v float64) sql.NullString {
	if v == 0 {
		return sql.NullString{}
	}
	return sql.NullString{String: fmt.Sprintf("%.2f", v), Valid: true}
}
