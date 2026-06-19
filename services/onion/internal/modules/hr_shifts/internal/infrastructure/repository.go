package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
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

func (r *Repository) ListShifts(ctx context.Context, businessID int64, params domain.ListShiftsParams) ([]domain.ShiftRecord, int64, error) {
	total, err := r.queries.CountShifts(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to count shifts", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListShifts(ctx, sqlc.ListShiftsParams{
		BusinessID: businessID,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list shifts", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	shifts := make([]domain.ShiftRecord, 0, len(rows))
	for _, row := range rows {
		shifts = append(shifts, shiftFromRow(row))
	}
	return shifts, total, nil
}

func (r *Repository) GetShift(ctx context.Context, businessID, shiftID int64) (*domain.ShiftRecord, error) {
	row, err := r.queries.GetShift(ctx, sqlc.GetShiftParams{BusinessID: businessID, ID: shiftID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrShiftNotFound
		}
		logger.Ctx(ctx).Error("failed to get shift", logger.Any("business_id", businessID), logger.Any("shift_id", shiftID), logger.Error(err))
		return nil, err
	}
	s := shiftFromRow(row)
	return &s, nil
}

func (r *Repository) CreateShift(ctx context.Context, businessID int64, shift domain.ShiftRecord) (*domain.ShiftRecord, error) {
	result, err := r.queries.CreateShift(ctx, sqlc.CreateShiftParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		LocationID: shift.LocationID,
		Name:       shift.Name,
		StartTime:  shift.StartTime,
		EndTime:    shift.EndTime,
		Color:      shift.Color,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create shift", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	shift.ID = id
	shift.OwnerID = businessID
	shift.BusinessID = businessID
	return &shift, nil
}

func (r *Repository) UpdateShift(ctx context.Context, businessID int64, shift domain.ShiftRecord) error {
	return r.queries.UpdateShift(ctx, sqlc.UpdateShiftParams{
		LocationID: shift.LocationID,
		Name:       shift.Name,
		StartTime:  shift.StartTime,
		EndTime:    shift.EndTime,
		Color:      shift.Color,
		ID:         shift.ID,
		BusinessID: businessID,
	})
}

func (r *Repository) ListShiftsByLocation(ctx context.Context, businessID, locationID int64) ([]domain.ShiftRecord, error) {
	rows, err := r.queries.ListShiftsByLocation(ctx, sqlc.ListShiftsByLocationParams{
		BusinessID: businessID,
		LocationID: locationID,
	})
	if err != nil {
		return nil, err
	}
	shifts := make([]domain.ShiftRecord, 0, len(rows))
	for _, row := range rows {
		shifts = append(shifts, shiftFromRow(row))
	}
	return shifts, nil
}

func (r *Repository) ListRosterAssignments(ctx context.Context, businessID int64, weekStart, weekEnd string) ([]domain.ShiftAssignmentRecord, error) {
	ws, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return nil, err
	}
	we, err := time.Parse("2006-01-02", weekEnd)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListRosterAssignments(ctx, sqlc.ListRosterAssignmentsParams{
		BusinessID: businessID,
		Date:       ws,
		Date_2:     we,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list roster assignments", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}

	assignments := make([]domain.ShiftAssignmentRecord, 0, len(rows))
	for _, row := range rows {
		assignments = append(assignments, assignmentFromRow(row))
	}
	return assignments, nil
}

func (r *Repository) CreateShiftAssignment(ctx context.Context, businessID int64, assignment domain.ShiftAssignmentRecord) (*domain.ShiftAssignmentRecord, error) {
	assignDate, err := time.Parse("2006-01-02", assignment.Date)
	if err != nil {
		return nil, err
	}

	result, err := r.queries.CreateShiftAssignment(ctx, sqlc.CreateShiftAssignmentParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		EmployeeID: assignment.EmployeeID,
		ShiftID:    assignment.ShiftID,
		Date:       assignDate,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create shift assignment", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	assignment.ID = id
	assignment.OwnerID = businessID
	assignment.BusinessID = businessID
	assignment.Status = "scheduled"
	return &assignment, nil
}

func (r *Repository) PublishRosterAssignments(ctx context.Context, businessID int64, weekStart, weekEnd string) error {
	ws, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return err
	}
	we, err := time.Parse("2006-01-02", weekEnd)
	if err != nil {
		return err
	}
	return r.queries.PublishRosterAssignments(ctx, sqlc.PublishRosterAssignmentsParams{
		BusinessID: businessID,
		Date:       ws,
		Date_2:     we,
	})
}

func (r *Repository) ListEmployeeAssignmentsForDate(ctx context.Context, businessID, employeeID int64, date string) ([]domain.ShiftAssignmentRecord, error) {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListEmployeeAssignmentsForDate(ctx, sqlc.ListEmployeeAssignmentsForDateParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		Date:       d,
	})
	if err != nil {
		return nil, err
	}
	assignments := make([]domain.ShiftAssignmentRecord, 0, len(rows))
	for _, row := range rows {
		assignments = append(assignments, assignmentFromRow(row))
	}
	return assignments, nil
}

func shiftFromRow(row sqlc.HrShift) domain.ShiftRecord {
	return domain.ShiftRecord{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		LocationID: row.LocationID,
		Name:       row.Name,
		StartTime:  row.StartTime,
		EndTime:    row.EndTime,
		Color:      row.Color,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}

func assignmentFromRow(row sqlc.HrShiftAssignment) domain.ShiftAssignmentRecord {
	return domain.ShiftAssignmentRecord{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		EmployeeID: row.EmployeeID,
		ShiftID:    row.ShiftID,
		Date:       row.Date.Format("2006-01-02"),
		Status:     row.Status,
		Published:  row.Published,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
