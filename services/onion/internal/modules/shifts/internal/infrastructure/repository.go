package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/shifts/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) ListShifts(ctx context.Context, ownerID, staffID int64, date string) ([]domain.Shift, error) {
	// Dynamic WHERE clauses - keep on pool
	query := `
		SELECT id, owner_id, staff_id, department_id, date, start_time, end_time,
		       break_duration_minutes, status, clock_in_at, clock_out_at,
		       hours_worked, overtime_hours, notes, created_at, updated_at
		FROM hr_shifts WHERE owner_id = ?`

	args := []interface{}{ownerID}

	if staffID != 0 {
		query += ` AND staff_id = ?`
		args = append(args, staffID)
	}
	if date != "" {
		query += ` AND date = ?`
		args = append(args, date)
	}

	query += ` ORDER BY date DESC, start_time ASC`

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list shifts", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var shifts []domain.Shift
	for rows.Next() {
		var s domain.Shift
		if err := s.FromRow(rows); err != nil {
			return nil, err
		}
		shifts = append(shifts, s)
	}

	return shifts, rows.Err()
}

func (r *Repository) CreateShift(ctx context.Context, input domain.ShiftInput) (int64, error) {
	result, err := r.queries.CreateShift(ctx, sqlc.CreateShiftParams{
		OwnerID:              input.OwnerID,
		StaffID:              input.StaffID,
		DepartmentID:         sql.NullInt64{Int64: input.DepartmentID, Valid: input.DepartmentID != 0},
		Date:                 parseDate(input.Date),
		StartTime:            parseTime(input.StartTime),
		EndTime:              parseTime(input.EndTime),
		BreakDurationMinutes: sql.NullInt32{Int32: int32(input.BreakDurationMinutes), Valid: input.BreakDurationMinutes != 0},
		Notes:                sql.NullString{String: input.Notes, Valid: input.Notes != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create shift", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateShift(ctx context.Context, ownerID int64, input domain.ShiftUpdateInput) error {
	err := r.queries.UpdateShift(ctx, sqlc.UpdateShiftParams{
		DepartmentID:         sql.NullInt64{Int64: input.DepartmentID, Valid: input.DepartmentID != 0},
		Date:                 parseDate(input.Date),
		StartTime:            parseTime(input.StartTime),
		EndTime:              parseTime(input.EndTime),
		BreakDurationMinutes: sql.NullInt32{Int32: int32(input.BreakDurationMinutes), Valid: input.BreakDurationMinutes != 0},
		Status:               sqlc.NullHrShiftsStatus{HrShiftsStatus: sqlc.HrShiftsStatus(input.Status), Valid: input.Status != ""},
		Notes:                sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		ID:                   input.ID,
		OwnerID:              ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update shift", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteShift(ctx context.Context, ownerID, shiftID int64) error {
	err := r.queries.DeleteShift(ctx, sqlc.DeleteShiftParams{
		ID:      shiftID,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete shift", logger.Error(err))
		return err
	}
	return nil
}

func parseDate(s string) time.Time {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	return time.Time{}
}

func parseTime(s string) time.Time {
	if t, err := time.Parse("15:04:05", s); err == nil {
		return t
	}
	if t, err := time.Parse("15:04", s); err == nil {
		return t
	}
	return time.Time{}
}
