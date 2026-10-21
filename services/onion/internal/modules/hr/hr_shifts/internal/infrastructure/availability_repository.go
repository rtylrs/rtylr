package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// listAvailabilityForStaffDateSafetyLimit bounds ListAvailabilityForStaffDate
// (one staff member, one date/weekday pair). No caller paginates this query;
// legitimate results are at most a couple of rows (a dated exception plus the
// matching weekly pattern), so the cap only guards against pathological data.
const listAvailabilityForStaffDateSafetyLimit int32 = 100

// availabilityFromRow maps the generated model onto the domain record. A dated
// exception carries no weekday, which the domain represents as NoDayOfWeek
// rather than a nullable field.
func availabilityFromRow(row sqlc.HrStaffAvailability) domain.AvailabilityRecord {
	rec := domain.AvailabilityRecord{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		StaffID:     row.StaffID,
		DayOfWeek:   domain.NoDayOfWeek,
		StartTime:   trimSeconds(row.StartTime),
		EndTime:     trimSeconds(row.EndTime),
		IsAvailable: row.IsAvailable,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.DayOfWeek.Valid {
		rec.DayOfWeek = int32(row.DayOfWeek.Int16)
	}
	if row.SpecificDate.Valid {
		rec.SpecificDate = row.SpecificDate.Time.Format(time.DateOnly)
	}
	return rec
}

// trimSeconds reduces the "HH:MM:SS" MySQL hands back for a TIME column to the
// "HH:MM" the contract specifies. A value that is not in that shape is returned
// unchanged rather than silently truncated.
func trimSeconds(s string) string {
	if len(s) == len("15:04:05") && s[2] == ':' && s[5] == ':' {
		return s[:5]
	}
	return s
}

func nullDayOfWeek(day int32) sql.NullInt16 {
	if day == domain.NoDayOfWeek {
		return sql.NullInt16{}
	}
	return sql.NullInt16{Int16: int16(day), Valid: true} // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func (r *Repository) ListAvailability(ctx context.Context, businessID int64, params domain.ListAvailabilityParams) ([]domain.AvailabilityRecord, int64, error) {
	staffFilter := sql.NullInt64{}
	if params.StaffID > 0 {
		staffFilter = sql.NullInt64{Int64: params.StaffID, Valid: true}
	}
	dayFilter := sql.NullInt16{}
	if params.DayOfWeek != nil {
		dayFilter = sql.NullInt16{Int16: int16(*params.DayOfWeek), Valid: true} // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}

	total, err := r.queries.CountStaffAvailability(ctx, sqlc.CountStaffAvailabilityParams{
		BusinessID: businessID,
		StaffID:    staffFilter,
		DayOfWeek:  dayFilter,
		DateFrom:   nullDate(params.DateFrom),
		DateTo:     nullDate(params.DateTo),
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to count staff availability",
			logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListStaffAvailability(ctx, sqlc.ListStaffAvailabilityParams{
		BusinessID: businessID,
		StaffID:    staffFilter,
		DayOfWeek:  dayFilter,
		DateFrom:   nullDate(params.DateFrom),
		DateTo:     nullDate(params.DateTo),
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list staff availability",
			logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	entries := make([]domain.AvailabilityRecord, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, availabilityFromRow(row))
	}
	return entries, total, nil
}

func (r *Repository) GetAvailability(ctx context.Context, businessID, id int64) (*domain.AvailabilityRecord, error) {
	row, err := r.queries.GetStaffAvailability(ctx, sqlc.GetStaffAvailabilityParams{BusinessID: businessID, ID: id})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAvailabilityNotFound
		}
		logger.Ctx(ctx).Error("failed to get staff availability",
			logger.Any("business_id", businessID), logger.Any("availability_id", id), logger.Error(err))
		return nil, err
	}
	rec := availabilityFromRow(row)
	return &rec, nil
}

// UpsertAvailability creates the entry, or updates it in place when one already
// exists. An explicit ID targets that row; without one the contract's natural
// key (business, staff, weekday-or-date, start time) is used, so re-submitting
// an unchanged window is idempotent rather than producing a duplicate.
func (r *Repository) UpsertAvailability(ctx context.Context, businessID int64, entry domain.AvailabilityRecord) (*domain.AvailabilityRecord, error) {
	targetID := entry.ID
	if targetID == 0 {
		existing, err := r.queries.GetStaffAvailabilityByNaturalKey(ctx, sqlc.GetStaffAvailabilityByNaturalKeyParams{
			BusinessID:   businessID,
			StaffID:      entry.StaffID,
			DayOfWeek:    nullDayOfWeek(entry.DayOfWeek),
			SpecificDate: nullDate(entry.SpecificDate),
			StartTime:    entry.StartTime,
		})
		switch {
		case err == nil:
			targetID = existing.ID
		case errors.Is(err, sql.ErrNoRows):
			targetID = 0
		default:
			logger.Ctx(ctx).Error("failed to look up staff availability by natural key",
				logger.Any("business_id", businessID), logger.Any("staff_id", entry.StaffID), logger.Error(err))
			return nil, err
		}
	}

	if targetID == 0 {
		res, err := r.queries.CreateStaffAvailability(ctx, sqlc.CreateStaffAvailabilityParams{
			OwnerID:      entry.OwnerID,
			BusinessID:   businessID,
			StaffID:      entry.StaffID,
			DayOfWeek:    nullDayOfWeek(entry.DayOfWeek),
			SpecificDate: nullDate(entry.SpecificDate),
			StartTime:    entry.StartTime,
			EndTime:      entry.EndTime,
			IsAvailable:  entry.IsAvailable,
		})
		if err != nil {
			logger.Ctx(ctx).Error("failed to create staff availability",
				logger.Any("business_id", businessID), logger.Any("staff_id", entry.StaffID), logger.Error(err))
			return nil, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			logger.Ctx(ctx).Error("failed to read created staff availability id",
				logger.Any("business_id", businessID), logger.Error(err))
			return nil, err
		}
		targetID = id
	} else {
		if err := r.queries.UpdateStaffAvailability(ctx, sqlc.UpdateStaffAvailabilityParams{
			DayOfWeek:    nullDayOfWeek(entry.DayOfWeek),
			SpecificDate: nullDate(entry.SpecificDate),
			StartTime:    entry.StartTime,
			EndTime:      entry.EndTime,
			IsAvailable:  entry.IsAvailable,
			BusinessID:   businessID,
			ID:           targetID,
		}); err != nil {
			logger.Ctx(ctx).Error("failed to update staff availability",
				logger.Any("business_id", businessID), logger.Any("availability_id", targetID), logger.Error(err))
			return nil, err
		}
	}

	return r.GetAvailability(ctx, businessID, targetID)
}

func (r *Repository) DeleteAvailability(ctx context.Context, businessID, id int64) error {
	res, err := r.queries.DeleteStaffAvailability(ctx, sqlc.DeleteStaffAvailabilityParams{BusinessID: businessID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("failed to delete staff availability",
			logger.Any("business_id", businessID), logger.Any("availability_id", id), logger.Error(err))
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		logger.Ctx(ctx).Error("failed to read delete result for staff availability",
			logger.Any("business_id", businessID), logger.Any("availability_id", id), logger.Error(err))
		return err
	}
	if affected == 0 {
		return domain.ErrAvailabilityNotFound
	}
	return nil
}

func (r *Repository) ListAvailabilityForStaffDate(ctx context.Context, businessID, staffID int64, date string, dayOfWeek int32) ([]domain.AvailabilityRecord, error) {
	rows, err := r.queries.ListStaffAvailabilityForDate(ctx, sqlc.ListStaffAvailabilityForDateParams{
		BusinessID: businessID,
		StaffID:    staffID,
		OnDate:     nullDate(date),
		DayOfWeek:  nullDayOfWeek(dayOfWeek),
		Limit:      listAvailabilityForStaffDateSafetyLimit,
		Offset:     0,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list staff availability for date",
			logger.Any("business_id", businessID), logger.Any("staff_id", staffID),
			logger.Any("date", date), logger.Error(err))
		return nil, err
	}

	entries := make([]domain.AvailabilityRecord, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, availabilityFromRow(row))
	}
	return entries, nil
}

// StaffIDByUserID resolves the signed-in user to their hr_staff row.
//
// cross-schema read from hr — hr_staff is owned by the staff module; availability
// keys on hr_staff.id and the caller is identified by user_id.
func (r *Repository) StaffIDByUserID(ctx context.Context, businessID, userID int64) (int64, error) {
	const q = `SELECT id FROM hr_staff WHERE business_id = ? AND user_id = ? ORDER BY id ASC LIMIT 1`
	var staffID int64
	if err := r.pool.QueryRowContext(ctx, q, businessID, userID).Scan(&staffID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrStaffNotLinked
		}
		logger.Ctx(ctx).Error("failed to resolve staff id from user id",
			logger.Any("business_id", businessID), logger.Any("user_id", userID), logger.Error(err))
		return 0, err
	}
	return staffID, nil
}

// StaffIDByEmployeeID crosses the hr_staff.employee_id bridge, which is nullable.
// An employee with no linked staff row has no availability to check, reported as
// ErrStaffNotLinked so the caller can decide what that means.
//
// cross-schema read from hr — hr_staff is owned by the staff module; shift
// assignments key on hr_employees.id and availability keys on hr_staff.id.
func (r *Repository) StaffIDByEmployeeID(ctx context.Context, businessID, employeeID int64) (int64, error) {
	const q = `SELECT id FROM hr_staff WHERE business_id = ? AND employee_id = ? ORDER BY id ASC LIMIT 1`
	var staffID int64
	if err := r.pool.QueryRowContext(ctx, q, businessID, employeeID).Scan(&staffID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrStaffNotLinked
		}
		logger.Ctx(ctx).Error("failed to resolve staff id from employee id",
			logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return 0, err
	}
	return staffID, nil
}
