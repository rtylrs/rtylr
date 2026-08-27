package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/benefits/internal/domain"
	"github.com/voxire/onion/internal/modules/benefits/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// mysqlDuplicateEntry is MySQL's ER_DUP_ENTRY. The uniq_hr_benefit_enrollment
// key turns a repeat enrollment into this error rather than a second row.
const mysqlDuplicateEntry = 1062

// CreateEnrollment enrolls one employee in one plan and returns the stored row.
// The plan is verified against the caller's business first, so a plan belonging
// to another tenant is indistinguishable from one that does not exist.
func (r *Repository) CreateEnrollment(ctx context.Context, businessID, employeeID, planID int64) (domain.BenefitEnrollment, error) {
	if _, err := r.queries.GetBenefitPlanIDForBusiness(ctx, sqlc.GetBenefitPlanIDForBusinessParams{
		BusinessID: businessID,
		ID:         planID,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.BenefitEnrollment{}, domain.ErrBenefitPlanNotFound
		}
		logger.Ctx(ctx).Error("failed to verify benefit plan for enrollment",
			logger.Any("business_id", businessID),
			logger.Any("plan_id", planID),
			logger.Error(err))
		return domain.BenefitEnrollment{}, err
	}

	res, err := r.queries.CreateBenefitEnrollment(ctx, sqlc.CreateBenefitEnrollmentParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		EmployeeID: employeeID,
		PlanID:     planID,
	})
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntry {
			return domain.BenefitEnrollment{}, domain.ErrAlreadyEnrolled
		}
		logger.Ctx(ctx).Error("failed to create benefit enrollment",
			logger.Any("business_id", businessID),
			logger.Any("employee_id", employeeID),
			logger.Any("plan_id", planID),
			logger.Error(err))
		return domain.BenefitEnrollment{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("failed to read new benefit enrollment id",
			logger.Any("business_id", businessID),
			logger.Error(err))
		return domain.BenefitEnrollment{}, err
	}

	row, err := r.queries.GetBenefitEnrollmentByID(ctx, sqlc.GetBenefitEnrollmentByIDParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to reload created benefit enrollment",
			logger.Any("business_id", businessID),
			logger.Any("enrollment_id", id),
			logger.Error(err))
		return domain.BenefitEnrollment{}, err
	}

	return enrollmentFromRow(row), nil
}

// ListEnrollmentsByEmployee returns one employee's own enrollments, newest
// first, along with the unpaginated total.
func (r *Repository) ListEnrollmentsByEmployee(ctx context.Context, businessID, employeeID int64, limit, offset int32) ([]domain.BenefitEnrollment, int32, error) {
	rows, err := r.queries.ListBenefitEnrollmentsByEmployee(ctx, sqlc.ListBenefitEnrollmentsByEmployeeParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list benefit enrollments",
			logger.Any("business_id", businessID),
			logger.Any("employee_id", employeeID),
			logger.Error(err))
		return nil, 0, err
	}

	total, err := r.queries.CountBenefitEnrollmentsByEmployee(ctx, sqlc.CountBenefitEnrollmentsByEmployeeParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to count benefit enrollments",
			logger.Any("business_id", businessID),
			logger.Any("employee_id", employeeID),
			logger.Error(err))
		return nil, 0, err
	}

	items := make([]domain.BenefitEnrollment, 0, len(rows))
	for _, row := range rows {
		items = append(items, enrollmentFromRow(row))
	}
	return items, int32(total), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func enrollmentFromRow(row sqlc.HrBenefitEnrollment) domain.BenefitEnrollment {
	item := domain.BenefitEnrollment{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		EmployeeID: row.EmployeeID,
		PlanID:     row.PlanID,
		Status:     string(row.Status),
		EnrolledAt: row.EnrolledAt,
	}
	if row.CancelledAt.Valid {
		cancelledAt := row.CancelledAt.Time
		item.CancelledAt = &cancelledAt
	}
	return item
}
