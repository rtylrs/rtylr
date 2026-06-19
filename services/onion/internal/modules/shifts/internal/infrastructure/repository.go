package infrastructure

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/shifts/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListShifts(ctx context.Context, businessID, staffID int64, date string) ([]domain.Shift, error) {
	// Dynamic WHERE clauses - built with squirrel
	builder := sq.Select(
		"id", "owner_id", "business_id", "staff_id", "department_id", "date", "start_time", "end_time",
		"break_duration_minutes", "status", "clock_in_at", "clock_out_at",
		"hours_worked", "overtime_hours", "notes", "created_at", "updated_at",
	).
		From("hr_shifts").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.Expr("staff_id IS NOT NULL"))

	if staffID != 0 {
		builder = builder.Where(sq.Eq{"staff_id": staffID})
	}
	if date != "" {
		builder = builder.Where(sq.Eq{"date": date})
	}

	builder = builder.OrderBy("date DESC", "start_time ASC")

	query, args, err := builder.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build list shifts query", logger.Error(err))
		return nil, err
	}

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
		BusinessID:           input.BusinessID,
		StaffID:              sql.NullInt64{Int64: input.StaffID, Valid: true},
		DepartmentID:         sql.NullInt64{Int64: input.DepartmentID, Valid: input.DepartmentID != 0},
		Date:                 sql.NullTime{Time: parseDate(input.Date), Valid: true},
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

func (r *Repository) UpdateShift(ctx context.Context, businessID int64, input domain.ShiftUpdateInput) error {
	err := r.queries.UpdateShift(ctx, sqlc.UpdateShiftParams{
		DepartmentID:         sql.NullInt64{Int64: input.DepartmentID, Valid: input.DepartmentID != 0},
		Date:                 sql.NullTime{Time: parseDate(input.Date), Valid: true},
		StartTime:            parseTime(input.StartTime),
		EndTime:              parseTime(input.EndTime),
		BreakDurationMinutes: sql.NullInt32{Int32: int32(input.BreakDurationMinutes), Valid: input.BreakDurationMinutes != 0},
		Status:               sqlc.NullHrShiftsStatus{HrShiftsStatus: sqlc.HrShiftsStatus(input.Status), Valid: input.Status != ""},
		Notes:                sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		ID:                   input.ID,
		BusinessID:           businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update shift", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteShift(ctx context.Context, businessID, shiftID int64) error {
	err := r.queries.DeleteShift(ctx, sqlc.DeleteShiftParams{
		ID:         shiftID,
		BusinessID: businessID,
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
	ref := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if t, err := time.Parse("15:04:05", s); err == nil {
		return ref.Add(time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute + time.Duration(t.Second())*time.Second)
	}
	if t, err := time.Parse("15:04", s); err == nil {
		return ref.Add(time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute)
	}
	return ref
}
