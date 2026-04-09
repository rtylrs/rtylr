package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/shifts/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) ListShifts(ctx context.Context, ownerID int64, staffID int64, date string) ([]domain.Shift, error) {
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
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO hr_shifts (owner_id, staff_id, department_id, date, start_time, end_time, break_duration_minutes, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		input.OwnerID, input.StaffID,
		sql.NullInt64{Int64: input.DepartmentID, Valid: input.DepartmentID != 0},
		input.Date, input.StartTime, input.EndTime,
		sql.NullInt64{Int64: int64(input.BreakDurationMinutes), Valid: input.BreakDurationMinutes != 0},
		sql.NullString{String: input.Notes, Valid: input.Notes != ""},
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create shift", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateShift(ctx context.Context, ownerID int64, input domain.ShiftUpdateInput) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE hr_shifts
		 SET department_id = ?, date = ?, start_time = ?, end_time = ?, break_duration_minutes = ?,
		     status = ?, notes = ?, updated_at = NOW()
		 WHERE id = ? AND owner_id = ?`,
		sql.NullInt64{Int64: input.DepartmentID, Valid: input.DepartmentID != 0},
		input.Date, input.StartTime, input.EndTime,
		sql.NullInt64{Int64: int64(input.BreakDurationMinutes), Valid: input.BreakDurationMinutes != 0},
		input.Status,
		sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		input.ID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update shift", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteShift(ctx context.Context, ownerID, shiftID int64) error {
	_, err := r.pool.ExecContext(ctx,
		`DELETE FROM hr_shifts WHERE id = ? AND owner_id = ?`,
		shiftID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete shift", logger.Error(err))
		return err
	}
	return nil
}
