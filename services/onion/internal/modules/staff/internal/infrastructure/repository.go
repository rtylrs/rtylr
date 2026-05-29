package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

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
		s.Salary, _ = strconv.ParseFloat(row.Salary.String, 64)
	}
	if row.HourlyRate.Valid {
		s.HourlyRate, _ = strconv.ParseFloat(row.HourlyRate.String, 64)
	}
	if row.OvertimeRate.Valid {
		s.OvertimeRate, _ = strconv.ParseFloat(row.OvertimeRate.String, 64)
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
	// Dynamic WHERE clauses - keep on pool
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

	staff := make([]domain.Staff, 0)
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
	row, err := r.queries.GetStaffByID(ctx, sqlc.GetStaffByIDParams{
		ID:      staffID,
		OwnerID: ownerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrStaffNotFound
		}
		return nil, err
	}

	s := staffFromRow(row)
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

	result, err := r.queries.CreateStaff(ctx, sqlc.CreateStaffParams{
		OwnerID:               input.OwnerID,
		UserID:                sql.NullInt64{Int64: input.UserID, Valid: input.UserID != 0},
		DepartmentID:          sql.NullInt64{Int64: input.DepartmentID, Valid: input.DepartmentID != 0},
		FirstName:             input.FirstName,
		LastName:              input.LastName,
		Email:                 input.Email,
		Phone:                 sql.NullString{String: input.Phone, Valid: input.Phone != ""},
		Role:                  input.Role,
		EmploymentType:        sqlc.NullHrStaffEmploymentType{HrStaffEmploymentType: sqlc.HrStaffEmploymentType(employmentType), Valid: true},
		Status:                sqlc.NullHrStaffStatus{HrStaffStatus: sqlc.HrStaffStatus(staffStatus), Valid: true},
		Salary:                nullDecimal(input.Salary),
		HourlyRate:            nullDecimal(input.HourlyRate),
		OvertimeRate:          nullDecimal(input.OvertimeRate),
		StartDate:             parseDate(input.StartDate),
		EndDate:               parseNullDate(input.EndDate),
		Address:               sql.NullString{String: input.Address, Valid: input.Address != ""},
		EmergencyContactName:  sql.NullString{String: input.EmergencyContactName, Valid: input.EmergencyContactName != ""},
		EmergencyContactPhone: sql.NullString{String: input.EmergencyContactPhone, Valid: input.EmergencyContactPhone != ""},
		Apps:                  appsJSON,
		Notes:                 sql.NullString{String: input.Notes, Valid: input.Notes != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create staff", logger.Error(err))
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) UpdateStaff(ctx context.Context, ownerID int64, input domain.StaffUpdateInput) error {
	appsJSON := json.RawMessage("[]")
	if input.Apps != "" {
		appsJSON = json.RawMessage(input.Apps)
	}

	err := r.queries.UpdateStaff(ctx, sqlc.UpdateStaffParams{
		DepartmentID:          sql.NullInt64{Int64: input.DepartmentID, Valid: input.DepartmentID != 0},
		FirstName:             input.FirstName,
		LastName:              input.LastName,
		Email:                 input.Email,
		Phone:                 sql.NullString{String: input.Phone, Valid: input.Phone != ""},
		Role:                  input.Role,
		EmploymentType:        sqlc.NullHrStaffEmploymentType{HrStaffEmploymentType: sqlc.HrStaffEmploymentType(input.EmploymentType), Valid: input.EmploymentType != ""},
		Status:                sqlc.NullHrStaffStatus{HrStaffStatus: sqlc.HrStaffStatus(input.Status), Valid: input.Status != ""},
		Salary:                nullDecimal(input.Salary),
		HourlyRate:            nullDecimal(input.HourlyRate),
		OvertimeRate:          nullDecimal(input.OvertimeRate),
		EndDate:               parseNullDate(input.EndDate),
		Address:               sql.NullString{String: input.Address, Valid: input.Address != ""},
		EmergencyContactName:  sql.NullString{String: input.EmergencyContactName, Valid: input.EmergencyContactName != ""},
		EmergencyContactPhone: sql.NullString{String: input.EmergencyContactPhone, Valid: input.EmergencyContactPhone != ""},
		Apps:                  appsJSON,
		Notes:                 sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		ID:                    input.ID,
		OwnerID:               ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update staff", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) DeleteStaff(ctx context.Context, ownerID, staffID int64) error {
	err := r.queries.TerminateStaff(ctx, sqlc.TerminateStaffParams{
		ID:      staffID,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to terminate staff", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) ClockIn(ctx context.Context, ownerID, staffID int64) error {
	err := r.queries.ClockIn(ctx, sqlc.ClockInParams{
		OwnerID: ownerID,
		StaffID: staffID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to clock in", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) ClockOut(ctx context.Context, ownerID, staffID int64) error {
	err := r.queries.ClockOut(ctx, sqlc.ClockOutParams{
		OwnerID: ownerID,
		StaffID: staffID,
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
