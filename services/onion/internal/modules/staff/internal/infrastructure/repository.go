package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/staff/internal/domain"
	"github.com/voxire/onion/internal/modules/staff/internal/infrastructure/persistence/sqlc"
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

// staffFromRow maps an sqlc HrStaff model to the domain Staff type.
func staffFromRow(row sqlc.HrStaff) domain.Staff {
	var s domain.Staff
	s.ID = row.ID
	s.OwnerID = row.OwnerID
	s.BusinessID = row.BusinessID
	if row.UserID.Valid {
		s.UserID = row.UserID.Int64
	}
	if row.DepartmentID.Valid {
		s.DepartmentID = row.DepartmentID.Int64
	}
	s.FirstName = row.FirstName
	s.LastName = row.LastName
	s.Email = row.Email
	if row.Phone.Valid {
		s.Phone = row.Phone.String
	}
	s.Role = row.Role
	if row.EmploymentType.Valid {
		s.EmploymentType = string(row.EmploymentType.HrStaffEmploymentType)
	}
	if row.Status.Valid {
		s.Status = string(row.Status.HrStaffStatus)
	}
	if row.Salary.Valid {
		s.Salary = parseDecimal(row.Salary.String)
	}
	if row.HourlyRate.Valid {
		s.HourlyRate = parseDecimal(row.HourlyRate.String)
	}
	if row.OvertimeRate.Valid {
		s.OvertimeRate = parseDecimal(row.OvertimeRate.String)
	}
	s.StartDate = row.StartDate
	if row.EndDate.Valid {
		s.EndDate = row.EndDate.Time
	}
	if row.Address.Valid {
		s.Address = row.Address.String
	}
	if row.EmergencyContactName.Valid {
		s.EmergencyContactName = row.EmergencyContactName.String
	}
	if row.EmergencyContactPhone.Valid {
		s.EmergencyContactPhone = row.EmergencyContactPhone.String
	}
	if row.Apps != nil {
		s.Apps = string(row.Apps)
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

func (r *Repository) ListStaff(ctx context.Context, ownerID int64, search, departmentID, status string) ([]domain.Staff, error) {
	businessID := ownerID
	// cross-schema read from hr_departments — to include department name in staff list response
	// Dynamic WHERE clause - built with squirrel
	builder := sq.
		Select("s.id", "s.owner_id", "s.business_id", "s.user_id", "s.department_id",
			"COALESCE(d.name, '') AS department_name",
			"s.first_name", "s.last_name", "s.email", "s.phone",
			"s.role", "s.employment_type", "s.status", "s.salary", "s.hourly_rate", "s.overtime_rate",
			"s.start_date", "s.end_date", "s.address", "s.emergency_contact_name", "s.emergency_contact_phone",
			"s.apps", "s.notes", "s.created_at", "s.updated_at").
		From("hr_staff s").
		LeftJoin("hr_departments d ON d.id = s.department_id AND d.business_id = s.business_id").
		Where(sq.Eq{"s.business_id": businessID})

	if search != "" {
		like := "%" + search + "%"
		builder = builder.Where(sq.Or{
			sq.Like{"s.first_name": like},
			sq.Like{"s.last_name": like},
			sq.Like{"s.email": like},
			sq.Like{"s.role": like},
		})
	}
	if departmentID != "" {
		builder = builder.Where(sq.Eq{"s.department_id": departmentID})
	}
	if status != "" {
		builder = builder.Where(sq.Eq{"s.status": status})
	}

	builder = builder.OrderBy("s.last_name ASC", "s.first_name ASC")

	query, args, err := builder.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build list staff query", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list staff", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	staff := make([]domain.Staff, 0)
	for rows.Next() {
		var s domain.Staff
		var (
			userID                sql.NullInt64
			departmentID_         sql.NullInt64
			departmentName        sql.NullString
			phone                 sql.NullString
			employmentType        sql.NullString
			statusVal             sql.NullString
			salary                sql.NullString
			hourlyRate            sql.NullString
			overtimeRate          sql.NullString
			endDate               sql.NullTime
			address               sql.NullString
			emergencyContactName  sql.NullString
			emergencyContactPhone sql.NullString
			apps                  []byte
			notes                 sql.NullString
			createdAt             sql.NullTime
			updatedAt             sql.NullTime
		)
		if err := rows.Scan(
			&s.ID, &s.OwnerID, &s.BusinessID, &userID, &departmentID_, &departmentName,
			&s.FirstName, &s.LastName, &s.Email, &phone,
			&s.Role, &employmentType, &statusVal, &salary, &hourlyRate, &overtimeRate,
			&s.StartDate, &endDate, &address, &emergencyContactName, &emergencyContactPhone,
			&apps, &notes, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		if userID.Valid {
			s.UserID = userID.Int64
		}
		if departmentID_.Valid {
			s.DepartmentID = departmentID_.Int64
		}
		if departmentName.Valid {
			s.DepartmentName = departmentName.String
		}
		if phone.Valid {
			s.Phone = phone.String
		}
		if employmentType.Valid {
			s.EmploymentType = employmentType.String
		}
		if statusVal.Valid {
			s.Status = statusVal.String
		}
		if salary.Valid {
			if v, err2 := strconv.ParseFloat(salary.String, 64); err2 == nil {
				s.Salary = v
			}
		}
		if hourlyRate.Valid {
			if v, err2 := strconv.ParseFloat(hourlyRate.String, 64); err2 == nil {
				s.HourlyRate = v
			}
		}
		if overtimeRate.Valid {
			if v, err2 := strconv.ParseFloat(overtimeRate.String, 64); err2 == nil {
				s.OvertimeRate = v
			}
		}
		if endDate.Valid {
			s.EndDate = endDate.Time
		}
		if address.Valid {
			s.Address = address.String
		}
		if emergencyContactName.Valid {
			s.EmergencyContactName = emergencyContactName.String
		}
		if emergencyContactPhone.Valid {
			s.EmergencyContactPhone = emergencyContactPhone.String
		}
		if apps != nil {
			s.Apps = string(apps)
		}
		if notes.Valid {
			s.Notes = notes.String
		}
		if createdAt.Valid {
			s.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			s.UpdatedAt = updatedAt.Time
		}
		staff = append(staff, s)
	}

	return staff, rows.Err()
}

func (r *Repository) GetStaffByID(ctx context.Context, ownerID, staffID int64) (*domain.Staff, error) {
	businessID := ownerID
	var s domain.Staff
	var (
		userID                sql.NullInt64
		departmentID          sql.NullInt64
		phone                 sql.NullString
		employmentType        sql.NullString
		statusVal             sql.NullString
		salary                sql.NullString
		hourlyRate            sql.NullString
		overtimeRate          sql.NullString
		endDate               sql.NullTime
		address               sql.NullString
		emergencyContactName  sql.NullString
		emergencyContactPhone sql.NullString
		apps                  []byte
		notes                 sql.NullString
		createdAt             sql.NullTime
		updatedAt             sql.NullTime
	)

	err := r.pool.QueryRowContext(ctx, `
		SELECT id, owner_id, business_id, user_id, department_id, first_name, last_name, email, phone,
		       role, employment_type, status, salary, hourly_rate, overtime_rate,
		       start_date, end_date, address, emergency_contact_name, emergency_contact_phone,
		       apps, notes, created_at, updated_at
		FROM hr_staff
		WHERE id = ? AND business_id = ?
	`, staffID, businessID).Scan(
		&s.ID, &s.OwnerID, &s.BusinessID, &userID, &departmentID,
		&s.FirstName, &s.LastName, &s.Email, &phone,
		&s.Role, &employmentType, &statusVal, &salary, &hourlyRate, &overtimeRate,
		&s.StartDate, &endDate, &address, &emergencyContactName, &emergencyContactPhone,
		&apps, &notes, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrStaffNotFound
		}
		return nil, err
	}

	if userID.Valid {
		s.UserID = userID.Int64
	}
	if departmentID.Valid {
		s.DepartmentID = departmentID.Int64
	}
	if phone.Valid {
		s.Phone = phone.String
	}
	if employmentType.Valid {
		s.EmploymentType = employmentType.String
	}
	if statusVal.Valid {
		s.Status = statusVal.String
	}
	if salary.Valid {
		s.Salary = parseDecimal(salary.String)
	}
	if hourlyRate.Valid {
		s.HourlyRate = parseDecimal(hourlyRate.String)
	}
	if overtimeRate.Valid {
		s.OvertimeRate = parseDecimal(overtimeRate.String)
	}
	if endDate.Valid {
		s.EndDate = endDate.Time
	}
	if address.Valid {
		s.Address = address.String
	}
	if emergencyContactName.Valid {
		s.EmergencyContactName = emergencyContactName.String
	}
	if emergencyContactPhone.Valid {
		s.EmergencyContactPhone = emergencyContactPhone.String
	}
	if apps != nil {
		s.Apps = string(apps)
	}
	if notes.Valid {
		s.Notes = notes.String
	}
	if createdAt.Valid {
		s.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		s.UpdatedAt = updatedAt.Time
	}

	return &s, nil
}

func (r *Repository) CreateStaff(ctx context.Context, input domain.StaffInput) (int64, error) {
	employmentType := input.EmploymentType
	if employmentType == "" {
		employmentType = "full_time"
	}

	staffStatus := input.Status
	if staffStatus == "" {
		staffStatus = "active"
	}

	appsJSON := json.RawMessage("[]")
	if input.Apps != "" {
		appsJSON = json.RawMessage(input.Apps)
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO hr_staff
			(owner_id, business_id, user_id, department_id, first_name, last_name, email, phone, role,
			 employment_type, status, salary, hourly_rate, overtime_rate, start_date, end_date,
			 address, emergency_contact_name, emergency_contact_phone, apps, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		input.OwnerID,
		input.BusinessID,
		sql.NullInt64{Int64: input.UserID, Valid: input.UserID != 0},
		sql.NullInt64{Int64: input.DepartmentID, Valid: input.DepartmentID != 0},
		input.FirstName,
		input.LastName,
		input.Email,
		sql.NullString{String: input.Phone, Valid: input.Phone != ""},
		input.Role,
		employmentType,
		staffStatus,
		nullDecimal(input.Salary),
		nullDecimal(input.HourlyRate),
		nullDecimal(input.OvertimeRate),
		parseDate(input.StartDate),
		parseNullDate(input.EndDate),
		sql.NullString{String: input.Address, Valid: input.Address != ""},
		sql.NullString{String: input.EmergencyContactName, Valid: input.EmergencyContactName != ""},
		sql.NullString{String: input.EmergencyContactPhone, Valid: input.EmergencyContactPhone != ""},
		appsJSON,
		sql.NullString{String: input.Notes, Valid: input.Notes != ""},
	)
	if err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to create staff", logger.Error(err))
		return 0, err
	}

	staffID, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	if err := r.syncEmployeeFromStaff(ctx, tx, input.BusinessID, staffID); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to sync created staff to HR employee", logger.Error(err))
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return staffID, nil
}

func (r *Repository) UpdateStaff(ctx context.Context, ownerID int64, input domain.StaffUpdateInput) error {
	businessID := ownerID
	appsJSON := json.RawMessage("[]")
	if input.Apps != "" {
		appsJSON = json.RawMessage(input.Apps)
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	_, err = tx.ExecContext(ctx, `
		UPDATE hr_staff
		SET department_id = ?, first_name = ?, last_name = ?, email = ?, phone = ?,
		    role = ?, employment_type = ?, status = ?, salary = ?, hourly_rate = ?,
		    overtime_rate = ?, end_date = ?, address = ?, emergency_contact_name = ?,
		    emergency_contact_phone = ?, apps = ?, notes = ?, updated_at = NOW()
		WHERE id = ? AND business_id = ?
	`,
		sql.NullInt64{Int64: input.DepartmentID, Valid: input.DepartmentID != 0},
		input.FirstName,
		input.LastName,
		input.Email,
		sql.NullString{String: input.Phone, Valid: input.Phone != ""},
		input.Role,
		sql.NullString{String: input.EmploymentType, Valid: input.EmploymentType != ""},
		sql.NullString{String: input.Status, Valid: input.Status != ""},
		nullDecimal(input.Salary),
		nullDecimal(input.HourlyRate),
		nullDecimal(input.OvertimeRate),
		parseNullDate(input.EndDate),
		sql.NullString{String: input.Address, Valid: input.Address != ""},
		sql.NullString{String: input.EmergencyContactName, Valid: input.EmergencyContactName != ""},
		sql.NullString{String: input.EmergencyContactPhone, Valid: input.EmergencyContactPhone != ""},
		appsJSON,
		sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		input.ID,
		businessID,
	)
	if err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to update staff", logger.Error(err))
		return err
	}

	if err := r.syncEmployeeFromStaff(ctx, tx, businessID, input.ID); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to sync updated staff to HR employee", logger.Error(err))
		return err
	}

	return tx.Commit()
}

func (r *Repository) DeleteStaff(ctx context.Context, ownerID, staffID int64) error {
	businessID := ownerID
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	_, err = tx.ExecContext(ctx, `
		UPDATE hr_staff SET status = 'terminated', updated_at = NOW()
		WHERE id = ? AND business_id = ?
	`, staffID, businessID)
	if err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to terminate staff", logger.Error(err))
		return err
	}

	if err := r.syncEmployeeFromStaff(ctx, tx, businessID, staffID); err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("Failed to sync terminated staff to HR employee", logger.Error(err))
		return err
	}

	return tx.Commit()
}

func (r *Repository) ClockIn(ctx context.Context, ownerID, staffID int64) error {
	err := r.queries.ClockIn(ctx, sqlc.ClockInParams{
		BusinessID: ownerID,
		StaffID:    sql.NullInt64{Int64: staffID, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to clock in", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) ClockOut(ctx context.Context, ownerID, staffID int64) error {
	err := r.queries.ClockOut(ctx, sqlc.ClockOutParams{
		BusinessID: ownerID,
		StaffID:    sql.NullInt64{Int64: staffID, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to clock out", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) CountActiveStaff(ctx context.Context, ownerID int64) (int, error) {
	count, err := r.queries.CountActiveStaff(ctx, ownerID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *Repository) CountOnShiftToday(ctx context.Context, ownerID int64) (int, error) {
	count, err := r.queries.CountOnShiftToday(ctx, ownerID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *Repository) CountPendingTimeOff(ctx context.Context, ownerID int64) (int, error) {
	count, err := r.queries.CountPendingTimeOff(ctx, ownerID)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *Repository) CountDraftPayroll(ctx context.Context, ownerID int64) (int, error) {
	count, err := r.queries.CountDraftPayroll(ctx, ownerID)
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
	EmploymentType        sql.NullString
	Status                sql.NullString
	Salary                sql.NullString
	HourlyRate            sql.NullString
	OvertimeRate          sql.NullString
	StartDate             sql.NullTime
	EndDate               sql.NullTime
	Address               sql.NullString
	EmergencyContactName  sql.NullString
	EmergencyContactPhone sql.NullString
	Notes                 sql.NullString
}

func (r *Repository) syncEmployeeFromStaff(ctx context.Context, tx *sql.Tx, ownerID, staffID int64) error {
	businessID := ownerID
	staff, err := loadStaffEmployeeSyncRow(ctx, tx, businessID, staffID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	employeeID := staff.EmployeeID
	if !employeeID.Valid && staff.UserID.Valid {
		employeeID, err = findEmployeeIDByUser(ctx, tx, staff.BusinessID, staff.UserID.Int64)
		if err != nil {
			return err
		}
	}
	if !employeeID.Valid && staff.Email != "" {
		employeeID, err = findEmployeeIDByEmail(ctx, tx, staff.BusinessID, staff.Email)
		if err != nil {
			return err
		}
	}

	if employeeID.Valid {
		if err := updateEmployeeFromStaff(ctx, tx, employeeID.Int64, staff); err != nil {
			return err
		}
		return linkStaffEmployee(ctx, tx, staff.BusinessID, staff.StaffID, employeeID.Int64)
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO hr_employees
			(owner_id, business_id, user_id, department_id, first_name, last_name, email, phone,
			 employment_type, status, salary, hourly_rate, overtime_rate, start_date,
			 end_date, address, emergency_contact_name, emergency_contact_phone, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		staff.OwnerID,
		staff.BusinessID,
		staff.UserID,
		staff.DepartmentID,
		staff.FirstName,
		staff.LastName,
		staff.Email,
		staff.Phone,
		normalizedStaffEmploymentType(staff.EmploymentType),
		normalizedStaffStatus(staff.Status),
		staff.Salary,
		staff.HourlyRate,
		staff.OvertimeRate,
		staff.StartDate,
		staff.EndDate,
		staff.Address,
		staff.EmergencyContactName,
		staff.EmergencyContactPhone,
		staff.Notes,
	)
	if err != nil {
		return err
	}

	newEmployeeID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	return linkStaffEmployee(ctx, tx, staff.BusinessID, staff.StaffID, newEmployeeID)
}

func loadStaffEmployeeSyncRow(ctx context.Context, tx *sql.Tx, businessID, staffID int64) (staffEmployeeSyncRow, error) {
	var staff staffEmployeeSyncRow
	err := tx.QueryRowContext(ctx, `
		SELECT owner_id, business_id, id, employee_id, user_id, department_id, first_name, last_name,
		       email, phone, employment_type, status, salary, hourly_rate, overtime_rate,
		       start_date, end_date, address, emergency_contact_name, emergency_contact_phone, notes
		FROM hr_staff
		WHERE business_id = ? AND id = ?
	`, businessID, staffID).Scan(
		&staff.OwnerID,
		&staff.BusinessID,
		&staff.StaffID,
		&staff.EmployeeID,
		&staff.UserID,
		&staff.DepartmentID,
		&staff.FirstName,
		&staff.LastName,
		&staff.Email,
		&staff.Phone,
		&staff.EmploymentType,
		&staff.Status,
		&staff.Salary,
		&staff.HourlyRate,
		&staff.OvertimeRate,
		&staff.StartDate,
		&staff.EndDate,
		&staff.Address,
		&staff.EmergencyContactName,
		&staff.EmergencyContactPhone,
		&staff.Notes,
	)
	return staff, err
}

func findEmployeeIDByUser(ctx context.Context, tx *sql.Tx, businessID, userID int64) (sql.NullInt64, error) {
	var employeeID sql.NullInt64
	err := tx.QueryRowContext(ctx, `
		SELECT id FROM hr_employees
		WHERE business_id = ? AND user_id = ?
		ORDER BY id ASC
		LIMIT 1
	`, businessID, userID).Scan(&employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return sql.NullInt64{}, nil
	}
	return employeeID, err
}

func findEmployeeIDByEmail(ctx context.Context, tx *sql.Tx, businessID int64, email string) (sql.NullInt64, error) {
	var employeeID sql.NullInt64
	err := tx.QueryRowContext(ctx, `
		SELECT id FROM hr_employees
		WHERE business_id = ? AND LOWER(email) = LOWER(?)
		ORDER BY id ASC
		LIMIT 1
	`, businessID, email).Scan(&employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return sql.NullInt64{}, nil
	}
	return employeeID, err
}

func updateEmployeeFromStaff(ctx context.Context, tx *sql.Tx, employeeID int64, staff staffEmployeeSyncRow) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE hr_employees
		SET user_id = COALESCE(?, user_id),
		    department_id = ?, first_name = ?, last_name = ?, email = ?, phone = ?,
		    employment_type = ?, status = ?, salary = ?, hourly_rate = ?,
		    overtime_rate = ?, start_date = ?, end_date = ?, address = ?,
		    emergency_contact_name = ?, emergency_contact_phone = ?, notes = ?,
		    updated_at = NOW()
		WHERE id = ? AND business_id = ?
	`,
		staff.UserID,
		staff.DepartmentID,
		staff.FirstName,
		staff.LastName,
		staff.Email,
		staff.Phone,
		normalizedStaffEmploymentType(staff.EmploymentType),
		normalizedStaffStatus(staff.Status),
		staff.Salary,
		staff.HourlyRate,
		staff.OvertimeRate,
		staff.StartDate,
		staff.EndDate,
		staff.Address,
		staff.EmergencyContactName,
		staff.EmergencyContactPhone,
		staff.Notes,
		employeeID,
		staff.BusinessID,
	)
	return err
}

func linkStaffEmployee(ctx context.Context, tx *sql.Tx, businessID, staffID, employeeID int64) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE hr_staff
		SET employee_id = ?
		WHERE id = ? AND business_id = ?
	`, employeeID, staffID, businessID)
	return err
}

func normalizedStaffEmploymentType(value sql.NullString) string {
	if value.Valid && value.String != "" {
		return value.String
	}
	return "full_time"
}

func normalizedStaffStatus(value sql.NullString) string {
	if value.Valid && value.String != "" {
		return value.String
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
func parseDecimal(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse staff decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}
