package infrastructure

import (
	"context"
	"database/sql"
	"strconv"
	"time"

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
	params := sqlc.ListShiftsParams{BusinessID: businessID}
	if staffID != 0 {
		params.StaffID = sql.NullInt64{Int64: staffID, Valid: true}
	}
	if date != "" {
		if t, err := time.Parse("2006-01-02", date); err == nil {
			params.Date = sql.NullTime{Time: t, Valid: true}
		}
	}

	rows, err := r.queries.ListShifts(ctx, params)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list shifts", logger.Error(err))
		return nil, err
	}

	shifts := make([]domain.Shift, 0, len(rows))
	for _, row := range rows {
		shifts = append(shifts, shiftFromSqlcRow(row))
	}

	return shifts, nil
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

func parseDecimal(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse shift decimal value from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}

func shiftFromSqlcRow(row sqlc.ListShiftsRow) domain.Shift {
	s := domain.Shift{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		StartTime:  row.StartTime.Format("15:04:05"),
		EndTime:    row.EndTime.Format("15:04:05"),
	}
	if row.StaffID.Valid {
		s.StaffID = row.StaffID.Int64
	}
	if row.DepartmentID.Valid {
		s.DepartmentID = row.DepartmentID.Int64
	}
	if row.Date.Valid {
		s.Date = row.Date.Time
	}
	if row.BreakDurationMinutes.Valid {
		s.BreakDurationMinutes = int(row.BreakDurationMinutes.Int32)
	}
	if row.Status.Valid {
		s.Status = string(row.Status.HrShiftsStatus)
	}
	if row.ClockInAt.Valid {
		s.ClockInAt = row.ClockInAt.Time
	}
	if row.ClockOutAt.Valid {
		s.ClockOutAt = row.ClockOutAt.Time
	}
	if row.HoursWorked.Valid {
		s.HoursWorked = parseDecimal(row.HoursWorked.String)
	}
	if row.OvertimeHours.Valid {
		s.OvertimeHours = parseDecimal(row.OvertimeHours.String)
	}
	if row.Notes.Valid {
		s.Notes = row.Notes.String
	}
	if row.CreatedAt.Valid {
		s.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		s.UpdatedAt = row.UpdatedAt.Time
	}
	return s
}
