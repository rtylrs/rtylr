package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

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
