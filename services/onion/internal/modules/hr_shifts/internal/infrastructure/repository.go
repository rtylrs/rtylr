package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

// mysqlDuplicateEntry is MySQL's ER_DUP_ENTRY. uk_shift_assignment turns a
// repeat assignment for the same employee and date into this error.
const mysqlDuplicateEntry = 1062

// The three safety limits below bound queries that have no caller-supplied
// pagination today. Each cap is set well above any legitimate result size for
// its scope, so it never truncates a real response; it exists only to stop a
// pathologically large table from being read unbounded into memory.
const (
	// listShiftsByLocationSafetyLimit bounds ListShiftsByLocation (all shift
	// rows for one location).
	listShiftsByLocationSafetyLimit int32 = 1000

	// listRosterAssignmentsSafetyLimit bounds ListRosterAssignments, which
	// backs the one-week roster view for a business.
	listRosterAssignmentsSafetyLimit int32 = 5000

	// listEmployeeAssignmentsForDateSafetyLimit bounds
	// ListEmployeeAssignmentsForDate (one employee, one date); legitimate
	// results are at most a handful of rows.
	listEmployeeAssignmentsForDateSafetyLimit int32 = 100
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
	total, err := r.queries.CountShifts(ctx, sqlc.CountShiftsParams{
		BusinessID: businessID,
		StaffID:    nullInt64(params.StaffID),
		Date:       nullDate(params.Date),
		LocationID: nullInt64(params.LocationID),
		Kind:       nullKind(params.Kind),
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to count shifts", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListShifts(ctx, sqlc.ListShiftsParams{
		BusinessID: businessID,
		StaffID:    nullInt64(params.StaffID),
		Date:       nullDate(params.Date),
		LocationID: nullInt64(params.LocationID),
		Kind:       nullKind(params.Kind),
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
		OwnerID:              businessID,
		BusinessID:           businessID,
		StaffID:              nullInt64(shift.StaffID),
		DepartmentID:         nullInt64(shift.DepartmentID),
		Date:                 nullDate(shift.Date),
		StartTime:            shift.StartTime,
		EndTime:              shift.EndTime,
		BreakDurationMinutes: nullInt32(shift.BreakDurationMinutes),
		Status:               nullStatus(shift.Status),
		Notes:                nullString(shift.Notes),
		LocationID:           nullInt64(shift.LocationID),
		Name:                 nullString(shift.Name),
		Color:                nullString(shift.Color),
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create shift", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("failed to read created shift id", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}

	// Read the row back so the caller sees the stored values, including the
	// column defaults for status and color rather than the empty input.
	return r.GetShift(ctx, businessID, id)
}

// UpdateShift row-locks the target shift and then writes it, so two concurrent
// updates to the same shift serialise instead of interleaving.
func (r *Repository) UpdateShift(ctx context.Context, businessID int64, shift domain.ShiftRecord) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("failed to begin shift update transaction", logger.Any("business_id", businessID), logger.Error(err))
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logger.Ctx(ctx).Error("failed to roll back after panic", logger.Error(rbErr))
			}
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	if _, err := qtx.GetShiftForUpdate(ctx, sqlc.GetShiftForUpdateParams{BusinessID: businessID, ID: shift.ID}); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Ctx(ctx).Error("failed to roll back shift update", logger.Error(rbErr))
		}
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrShiftNotFound
		}
		logger.Ctx(ctx).Error("failed to lock shift for update", logger.Any("business_id", businessID), logger.Any("shift_id", shift.ID), logger.Error(err))
		return err
	}

	if err := qtx.UpdateShift(ctx, sqlc.UpdateShiftParams{
		StaffID:              nullInt64(shift.StaffID),
		DepartmentID:         nullInt64(shift.DepartmentID),
		Date:                 nullDate(shift.Date),
		StartTime:            shift.StartTime,
		EndTime:              shift.EndTime,
		BreakDurationMinutes: nullInt32(shift.BreakDurationMinutes),
		Status:               nullStatus(shift.Status),
		Notes:                nullString(shift.Notes),
		LocationID:           nullInt64(shift.LocationID),
		Name:                 nullString(shift.Name),
		Color:                nullString(shift.Color),
		ID:                   shift.ID,
		BusinessID:           businessID,
	}); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Ctx(ctx).Error("failed to roll back shift update", logger.Error(rbErr))
		}
		logger.Ctx(ctx).Error("failed to update shift", logger.Any("business_id", businessID), logger.Any("shift_id", shift.ID), logger.Error(err))
		return err
	}

	return tx.Commit()
}

// DeleteShift row-locks the target shift before removing it, so a concurrent
// update cannot land between the existence check and the delete.
func (r *Repository) DeleteShift(ctx context.Context, businessID, shiftID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("failed to begin shift delete transaction", logger.Any("business_id", businessID), logger.Error(err))
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logger.Ctx(ctx).Error("failed to roll back after panic", logger.Error(rbErr))
			}
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	if _, err := qtx.GetShiftForUpdate(ctx, sqlc.GetShiftForUpdateParams{BusinessID: businessID, ID: shiftID}); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Ctx(ctx).Error("failed to roll back shift delete", logger.Error(rbErr))
		}
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrShiftNotFound
		}
		logger.Ctx(ctx).Error("failed to lock shift for delete", logger.Any("business_id", businessID), logger.Any("shift_id", shiftID), logger.Error(err))
		return err
	}

	if err := qtx.DeleteShift(ctx, sqlc.DeleteShiftParams{ID: shiftID, BusinessID: businessID}); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logger.Ctx(ctx).Error("failed to roll back shift delete", logger.Error(rbErr))
		}
		logger.Ctx(ctx).Error("failed to delete shift", logger.Any("business_id", businessID), logger.Any("shift_id", shiftID), logger.Error(err))
		return err
	}

	return tx.Commit()
}

func (r *Repository) ListShiftsByLocation(ctx context.Context, businessID, locationID int64) ([]domain.ShiftRecord, error) {
	rows, err := r.queries.ListShiftsByLocation(ctx, sqlc.ListShiftsByLocationParams{
		BusinessID: businessID,
		LocationID: nullInt64(locationID),
		Limit:      listShiftsByLocationSafetyLimit,
		Offset:     0,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list shifts by location", logger.Any("business_id", businessID), logger.Any("location_id", locationID), logger.Error(err))
		return nil, err
	}
	shifts := make([]domain.ShiftRecord, 0, len(rows))
	for _, row := range rows {
		shifts = append(shifts, shiftFromRow(row))
	}
	return shifts, nil
}

func (r *Repository) ListRosterAssignments(ctx context.Context, businessID int64, weekStart, weekEnd string) ([]domain.ShiftAssignmentRecord, error) {
	ws, err := time.Parse(time.DateOnly, weekStart)
	if err != nil {
		return nil, err
	}
	we, err := time.Parse(time.DateOnly, weekEnd)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListRosterAssignments(ctx, sqlc.ListRosterAssignmentsParams{
		BusinessID: businessID,
		Date:       ws,
		Date_2:     we,
		Limit:      listRosterAssignmentsSafetyLimit,
		Offset:     0,
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
	assignDate, err := time.Parse(time.DateOnly, assignment.Date)
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
		// uk_shift_assignment allows one assignment per (owner, employee, date).
		// The service checks first, but a concurrent request can still reach the
		// insert; report that as the roster conflict the endpoint documents
		// rather than an opaque 500.
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntry {
			return nil, domain.ErrRosterConflict
		}
		logger.Ctx(ctx).Error("failed to create shift assignment", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("failed to read created assignment id", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}

	// Read the row back rather than reconstructing it, so the caller gets the
	// stored status and the real created_at and updated_at instead of a zero time.
	row, err := r.queries.GetShiftAssignment(ctx, sqlc.GetShiftAssignmentParams{BusinessID: businessID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("failed to read back created assignment",
			logger.Any("business_id", businessID), logger.Any("assignment_id", id), logger.Error(err))
		return nil, err
	}

	created := assignmentFromRow(row)
	return &created, nil
}

func (r *Repository) PublishRosterAssignments(ctx context.Context, businessID int64, weekStart, weekEnd string) error {
	ws, err := time.Parse(time.DateOnly, weekStart)
	if err != nil {
		return err
	}
	we, err := time.Parse(time.DateOnly, weekEnd)
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
	d, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListEmployeeAssignmentsForDate(ctx, sqlc.ListEmployeeAssignmentsForDateParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		Limit:      listEmployeeAssignmentsForDateSafetyLimit,
		Offset:     0,
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
	s := domain.ShiftRecord{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		StartTime:  row.StartTime,
		EndTime:    row.EndTime,
	}
	if row.LocationID.Valid {
		s.LocationID = row.LocationID.Int64
	}
	if row.Name.Valid {
		s.Name = row.Name.String
	}
	if row.Color.Valid {
		s.Color = row.Color.String
	}
	if row.StaffID.Valid {
		s.StaffID = row.StaffID.Int64
	}
	if row.DepartmentID.Valid {
		s.DepartmentID = row.DepartmentID.Int64
	}
	if row.Date.Valid {
		s.Date = row.Date.Time.Format(time.DateOnly)
	}
	if row.BreakDurationMinutes.Valid {
		s.BreakDurationMinutes = row.BreakDurationMinutes.Int32
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
		s.HoursWorked = sqlutil.ParseDecimal(row.HoursWorked.String)
	}
	if row.OvertimeHours.Valid {
		s.OvertimeHours = sqlutil.ParseDecimal(row.OvertimeHours.String)
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

func assignmentFromRow(row sqlc.HrShiftAssignment) domain.ShiftAssignmentRecord {
	return domain.ShiftAssignmentRecord{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		EmployeeID: row.EmployeeID,
		ShiftID:    row.ShiftID,
		Date:       row.Date.Format(time.DateOnly),
		Status:     row.Status,
		Published:  row.Published,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}

var nullInt64 = sqlutil.NullInt64

func nullInt32(v int32) sql.NullInt32 {
	return sql.NullInt32{Int32: v, Valid: v != 0}
}

var nullString = sqlutil.NullString

func nullStatus(v string) sqlc.NullHrShiftsStatus {
	return sqlc.NullHrShiftsStatus{
		HrShiftsStatus: sqlc.HrShiftsStatus(v),
		Valid:          v != "",
	}
}

// nullKind renders the kind filter for the generated query, whose parameter is
// untyped because it is compared against string literals in SQL. A nil value
// leaves the filter off, so both kinds are returned.
func nullKind(kind domain.ShiftKind) interface{} {
	if kind == domain.ShiftKindAny {
		return nil
	}
	return string(kind)
}

// nullDate converts an ISO-8601 date to a nullable DATE. An unparseable or
// empty value leaves the filter off rather than silently matching nothing.
func nullDate(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		logger.Global().Warn("failed to parse shift date", logger.String("value", s), logger.Error(err))
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

// parseDecimal reads a DECIMAL column, which the MySQL driver returns as a
// string. A corrupt value logs and reads as zero rather than failing the row.
