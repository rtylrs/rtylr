package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/staff/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) ListStaff(ctx context.Context, ownerID int64, search, departmentID, status string) ([]domain.Staff, error) {
	query := `
		SELECT id, owner_id, user_id, department_id, first_name, last_name, email, phone,
		       role, employment_type, status, salary, hourly_rate, overtime_rate,
		       start_date, end_date, address, emergency_contact_name, emergency_contact_phone,
		       apps, notes, created_at, updated_at
		FROM hr_staff
		WHERE owner_id = ?`

	args := []interface{}{ownerID}

	if search != "" {
		query += ` AND (first_name LIKE ? OR last_name LIKE ? OR email LIKE ? OR role LIKE ?)`
		like := "%" + search + "%"
		args = append(args, like, like, like, like)
	}
	if departmentID != "" {
		query += ` AND department_id = ?`
		args = append(args, departmentID)
	}
	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}

	query += ` ORDER BY last_name ASC, first_name ASC`

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list staff", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var staff []domain.Staff
	for rows.Next() {
		var s domain.Staff
		if err := s.FromRow(rows); err != nil {
			return nil, err
		}
		staff = append(staff, s)
	}

	return staff, rows.Err()
}

func (r *Repository) GetStaffByID(ctx context.Context, ownerID, staffID int64) (*domain.Staff, error) {
	query := `
		SELECT id, owner_id, user_id, department_id, first_name, last_name, email, phone,
		       role, employment_type, status, salary, hourly_rate, overtime_rate,
		       start_date, end_date, address, emergency_contact_name, emergency_contact_phone,
		       apps, notes, created_at, updated_at
		FROM hr_staff
		WHERE id = ? AND owner_id = ?`

	rows, err := r.pool.QueryContext(ctx, query, staffID, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrStaffNotFound
	}

	var s domain.Staff
	if err := s.FromRow(rows); err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *Repository) CreateStaff(ctx context.Context, input domain.StaffInput) (int64, error) {
	query := `
		INSERT INTO hr_staff
		(owner_id, user_id, department_id, first_name, last_name, email, phone, role,
		 employment_type, status, salary, hourly_rate, overtime_rate, start_date, end_date,
		 address, emergency_contact_name, emergency_contact_phone, apps, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var (
		userID       sql.NullInt64
		departmentID sql.NullInt64
		salary       sql.NullFloat64
		hourlyRate   sql.NullFloat64
		overtimeRate sql.NullFloat64
		endDate      sql.NullString
	)

	if input.UserID != 0 {
		userID = sql.NullInt64{Int64: input.UserID, Valid: true}
	}
	if input.DepartmentID != 0 {
		departmentID = sql.NullInt64{Int64: input.DepartmentID, Valid: true}
	}
	if input.Salary != 0 {
		salary = sql.NullFloat64{Float64: input.Salary, Valid: true}
	}
	if input.HourlyRate != 0 {
		hourlyRate = sql.NullFloat64{Float64: input.HourlyRate, Valid: true}
	}
	if input.OvertimeRate != 0 {
		overtimeRate = sql.NullFloat64{Float64: input.OvertimeRate, Valid: true}
	}
	if input.EndDate != "" {
		endDate = sql.NullString{String: input.EndDate, Valid: true}
	}

	employmentType := input.EmploymentType
	if employmentType == "" {
		employmentType = "full_time"
	}

	staffStatus := input.Status
	if staffStatus == "" {
		staffStatus = "active"
	}

	result, err := r.pool.ExecContext(ctx, query,
		input.OwnerID, userID, departmentID, input.FirstName, input.LastName, input.Email,
		sql.NullString{String: input.Phone, Valid: input.Phone != ""},
		input.Role, employmentType, staffStatus, salary, hourlyRate, overtimeRate,
		input.StartDate, endDate,
		sql.NullString{String: input.Address, Valid: input.Address != ""},
		sql.NullString{String: input.EmergencyContactName, Valid: input.EmergencyContactName != ""},
		sql.NullString{String: input.EmergencyContactPhone, Valid: input.EmergencyContactPhone != ""},
		sql.NullString{String: input.Apps, Valid: input.Apps != ""},
		sql.NullString{String: input.Notes, Valid: input.Notes != ""},
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create staff", logger.Error(err))
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) UpdateStaff(ctx context.Context, ownerID int64, input domain.StaffUpdateInput) error {
	query := `
		UPDATE hr_staff
		SET department_id = ?, first_name = ?, last_name = ?, email = ?, phone = ?,
		    role = ?, employment_type = ?, status = ?, salary = ?, hourly_rate = ?,
		    overtime_rate = ?, end_date = ?, address = ?, emergency_contact_name = ?,
		    emergency_contact_phone = ?, apps = ?, notes = ?, updated_at = NOW()
		WHERE id = ? AND owner_id = ?`

	var (
		departmentID sql.NullInt64
		salary       sql.NullFloat64
		hourlyRate   sql.NullFloat64
		overtimeRate sql.NullFloat64
		endDate      sql.NullString
	)

	if input.DepartmentID != 0 {
		departmentID = sql.NullInt64{Int64: input.DepartmentID, Valid: true}
	}
	if input.Salary != 0 {
		salary = sql.NullFloat64{Float64: input.Salary, Valid: true}
	}
	if input.HourlyRate != 0 {
		hourlyRate = sql.NullFloat64{Float64: input.HourlyRate, Valid: true}
	}
	if input.OvertimeRate != 0 {
		overtimeRate = sql.NullFloat64{Float64: input.OvertimeRate, Valid: true}
	}
	if input.EndDate != "" {
		endDate = sql.NullString{String: input.EndDate, Valid: true}
	}

	_, err := r.pool.ExecContext(ctx, query,
		departmentID, input.FirstName, input.LastName, input.Email,
		sql.NullString{String: input.Phone, Valid: input.Phone != ""},
		input.Role, input.EmploymentType, input.Status, salary, hourlyRate, overtimeRate, endDate,
		sql.NullString{String: input.Address, Valid: input.Address != ""},
		sql.NullString{String: input.EmergencyContactName, Valid: input.EmergencyContactName != ""},
		sql.NullString{String: input.EmergencyContactPhone, Valid: input.EmergencyContactPhone != ""},
		sql.NullString{String: input.Apps, Valid: input.Apps != ""},
		sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		input.ID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update staff", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) DeleteStaff(ctx context.Context, ownerID, staffID int64) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE hr_staff SET status = 'terminated', updated_at = NOW() WHERE id = ? AND owner_id = ?`,
		staffID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to terminate staff", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) CountActiveStaff(ctx context.Context, ownerID int64) (int, error) {
	var count int
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hr_staff WHERE owner_id = ? AND status = 'active'`, ownerID,
	).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) CountOnShiftToday(ctx context.Context, ownerID int64) (int, error) {
	var count int
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hr_shifts WHERE owner_id = ? AND date = CURDATE() AND status IN ('scheduled','in_progress')`, ownerID,
	).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) CountPendingTimeOff(ctx context.Context, ownerID int64) (int, error) {
	var count int
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hr_time_off_requests WHERE owner_id = ? AND status = 'pending'`, ownerID,
	).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) CountDraftPayroll(ctx context.Context, ownerID int64) (int, error) {
	var count int
	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hr_payroll_records WHERE owner_id = ? AND status = 'draft'`, ownerID,
	).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

