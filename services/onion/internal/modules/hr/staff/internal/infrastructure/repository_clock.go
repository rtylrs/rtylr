package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr/staff/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) ClockIn(ctx context.Context, businessID, staffID int64) error {
	err := r.queries.ClockIn(ctx, sqlc.ClockInParams{
		BusinessID: businessID,
		StaffID:    sql.NullInt64{Int64: staffID, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to clock in", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) ClockOut(ctx context.Context, businessID, staffID int64) error {
	err := r.queries.ClockOut(ctx, sqlc.ClockOutParams{
		BusinessID: businessID,
		StaffID:    sql.NullInt64{Int64: staffID, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to clock out", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) CountActiveStaff(ctx context.Context, businessID int64) (int, error) {
	count, err := r.queries.CountActiveStaff(ctx, businessID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *Repository) CountOnShiftToday(ctx context.Context, businessID int64) (int, error) {
	count, err := r.queries.CountOnShiftToday(ctx, businessID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *Repository) CountPendingTimeOff(ctx context.Context, businessID int64) (int, error) {
	count, err := r.queries.CountPendingTimeOff(ctx, businessID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *Repository) CountDraftPayroll(ctx context.Context, businessID int64) (int, error) {
	count, err := r.queries.CountDraftPayroll(ctx, businessID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

type staffEmployeeSyncRow struct {
	OwnerID               int64
	BusinessID            int64
	StaffID               int64
	EmployeeID            sql.NullInt64
	UserID                sql.NullInt64
	DepartmentID          sql.NullInt64
	FirstName             string
	LastName              string
	Email                 string
	Phone                 sql.NullString
	EmploymentType        sqlc.NullHrStaffEmploymentType
	Status                sqlc.NullHrStaffStatus
	Salary                sql.NullString
	HourlyRate            sql.NullString
	OvertimeRate          sql.NullString
	StartDate             time.Time
	EndDate               sql.NullTime
	Address               sql.NullString
	EmergencyContactName  sql.NullString
	EmergencyContactPhone sql.NullString
	Notes                 sql.NullString
}

func (r *Repository) syncEmployeeFromStaff(ctx context.Context, qtx *sqlc.Queries, businessID, staffID int64) error {
	staff, err := loadStaffEmployeeSyncRow(ctx, qtx, businessID, staffID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	employeeID := staff.EmployeeID
	if !employeeID.Valid && staff.UserID.Valid {
		employeeID, err = findEmployeeIDByUser(ctx, qtx, staff.BusinessID, staff.UserID.Int64)
		if err != nil {
			return err
		}
	}
	if !employeeID.Valid && staff.Email != "" {
		employeeID, err = findEmployeeIDByEmail(ctx, qtx, staff.BusinessID, staff.Email)
		if err != nil {
			return err
		}
	}

	if employeeID.Valid {
		if err := updateEmployeeFromStaff(ctx, qtx, employeeID.Int64, staff); err != nil {
			return err
		}
		return linkStaffEmployee(ctx, qtx, staff.BusinessID, staff.StaffID, employeeID.Int64)
	}

	result, err := qtx.InsertEmployeeFromStaff(ctx, sqlc.InsertEmployeeFromStaffParams{
		OwnerID:               staff.OwnerID,
		BusinessID:            staff.BusinessID,
		UserID:                staff.UserID,
		DepartmentID:          staff.DepartmentID,
		FirstName:             staff.FirstName,
		LastName:              staff.LastName,
		Email:                 staff.Email,
		Phone:                 staff.Phone,
		EmploymentType:        normalizedStaffEmploymentType(staff.EmploymentType),
		Status:                normalizedStaffStatus(staff.Status),
		Salary:                staff.Salary,
		HourlyRate:            staff.HourlyRate,
		OvertimeRate:          staff.OvertimeRate,
		StartDate:             sql.NullTime{Time: staff.StartDate, Valid: !staff.StartDate.IsZero()},
		EndDate:               staff.EndDate,
		Address:               staff.Address,
		EmergencyContactName:  staff.EmergencyContactName,
		EmergencyContactPhone: staff.EmergencyContactPhone,
		Notes:                 staff.Notes,
	})
	if err != nil {
		return err
	}

	newEmployeeID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	return linkStaffEmployee(ctx, qtx, staff.BusinessID, staff.StaffID, newEmployeeID)
}

func loadStaffEmployeeSyncRow(ctx context.Context, qtx *sqlc.Queries, businessID, staffID int64) (staffEmployeeSyncRow, error) {
	row, err := qtx.GetStaffEmployeeSyncRow(ctx, sqlc.GetStaffEmployeeSyncRowParams{
		BusinessID: businessID,
		ID:         staffID,
	})
	if err != nil {
		return staffEmployeeSyncRow{}, err
	}
	return staffEmployeeSyncRow{
		OwnerID:               row.OwnerID,
		BusinessID:            row.BusinessID,
		StaffID:               row.ID,
		EmployeeID:            row.EmployeeID,
		UserID:                row.UserID,
		DepartmentID:          row.DepartmentID,
		FirstName:             row.FirstName,
		LastName:              row.LastName,
		Email:                 row.Email,
		Phone:                 row.Phone,
		EmploymentType:        row.EmploymentType,
		Status:                row.Status,
		Salary:                row.Salary,
		HourlyRate:            row.HourlyRate,
		OvertimeRate:          row.OvertimeRate,
		StartDate:             row.StartDate,
		EndDate:               row.EndDate,
		Address:               row.Address,
		EmergencyContactName:  row.EmergencyContactName,
		EmergencyContactPhone: row.EmergencyContactPhone,
		Notes:                 row.Notes,
	}, nil
}

func findEmployeeIDByUser(ctx context.Context, qtx *sqlc.Queries, businessID, userID int64) (sql.NullInt64, error) {
	id, err := qtx.FindEmployeeIDByUser(ctx, sqlc.FindEmployeeIDByUserParams{
		BusinessID: businessID,
		UserID:     sql.NullInt64{Int64: userID, Valid: true},
	})
	if errors.Is(err, sql.ErrNoRows) {
		return sql.NullInt64{}, nil
	}
	if err != nil {
		return sql.NullInt64{}, err
	}
	return sql.NullInt64{Int64: id, Valid: true}, nil
}

func findEmployeeIDByEmail(ctx context.Context, qtx *sqlc.Queries, businessID int64, email string) (sql.NullInt64, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	id, err := qtx.FindEmployeeIDByEmail(ctx, sqlc.FindEmployeeIDByEmailParams{
		BusinessID: businessID,
		Email:      email,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return sql.NullInt64{}, nil
	}
	if err != nil {
		return sql.NullInt64{}, err
	}
	return sql.NullInt64{Int64: id, Valid: true}, nil
}

func updateEmployeeFromStaff(ctx context.Context, qtx *sqlc.Queries, employeeID int64, staff staffEmployeeSyncRow) error {
	return qtx.UpdateEmployeeFromStaff(ctx, sqlc.UpdateEmployeeFromStaffParams{
		UserID:                staff.UserID,
		DepartmentID:          staff.DepartmentID,
		FirstName:             staff.FirstName,
		LastName:              staff.LastName,
		Email:                 staff.Email,
		Phone:                 staff.Phone,
		EmploymentType:        normalizedStaffEmploymentType(staff.EmploymentType),
		Status:                normalizedStaffStatus(staff.Status),
		Salary:                staff.Salary,
		HourlyRate:            staff.HourlyRate,
		OvertimeRate:          staff.OvertimeRate,
		StartDate:             sql.NullTime{Time: staff.StartDate, Valid: !staff.StartDate.IsZero()},
		EndDate:               staff.EndDate,
		Address:               staff.Address,
		EmergencyContactName:  staff.EmergencyContactName,
		EmergencyContactPhone: staff.EmergencyContactPhone,
		Notes:                 staff.Notes,
		ID:                    employeeID,
		BusinessID:            staff.BusinessID,
	})
}

func linkStaffEmployee(ctx context.Context, qtx *sqlc.Queries, businessID, staffID, employeeID int64) error {
	return qtx.LinkStaffEmployee(ctx, sqlc.LinkStaffEmployeeParams{
		EmployeeID: sql.NullInt64{Int64: employeeID, Valid: true},
		ID:         staffID,
		BusinessID: businessID,
	})
}

func normalizedStaffEmploymentType(value sqlc.NullHrStaffEmploymentType) string {
	if value.Valid && value.HrStaffEmploymentType != "" {
		return string(value.HrStaffEmploymentType)
	}
	return "full_time"
}

func normalizedStaffStatus(value sqlc.NullHrStaffStatus) string {
	if value.Valid && value.HrStaffStatus != "" {
		return string(value.HrStaffStatus)
	}
	return "active"
}

func nullDecimal(v float64) sql.NullString {
	if v == 0 {
		return sql.NullString{}
	}
	return sql.NullString{String: strconv.FormatFloat(v, 'g', -1, 64), Valid: true}
}

func parseDate(s string) time.Time {
	if p, err := time.Parse("2006-01-02", s); err == nil {
		return p
	}
	return time.Time{}
}

func parseNullDate(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return sql.NullTime{Time: t, Valid: true}
	}
	return sql.NullTime{}
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing them to 0.

// ListManagedUserIDs returns the user ids reporting to managerUserID inside the
// business. Managing nobody is the common case and returns an empty slice, not
// an error.
//
// user_id is nullable on hr_staff, so the query already excludes NULLs; the
// Valid check here is belt-and-braces against a driver returning a zero id as
// valid, which would grant visibility over user 0.
func (r *Repository) ListManagedUserIDs(ctx context.Context, businessID, managerUserID int64) ([]int64, error) {
	// The query unions two rule sources (department manager, direct reporting
	// line), so the same manager and business fill both halves.
	rows, err := r.queries.ListManagedUserIDs(ctx, sqlc.ListManagedUserIDsParams{
		UserID:       sql.NullInt64{Int64: managerUserID, Valid: true},
		BusinessID:   businessID,
		UserID_2:     sql.NullInt64{Int64: managerUserID, Valid: true},
		BusinessID_2: businessID,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.Valid && row.Int64 > 0 {
			ids = append(ids, row.Int64)
		}
	}
	return ids, nil
}
