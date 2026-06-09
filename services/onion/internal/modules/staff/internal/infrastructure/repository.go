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
	// cross-schema read from hr_departments — to include department name in staff list response
	// Dynamic WHERE clause - built with squirrel
	builder := sq.
		Select("s.id", "s.owner_id", "s.user_id", "s.department_id",
			"COALESCE(d.name, '') AS department_name",
			"s.first_name", "s.last_name", "s.email", "s.phone",
			"s.role", "s.employment_type", "s.status", "s.salary", "s.hourly_rate", "s.overtime_rate",
			"s.start_date", "s.end_date", "s.address", "s.emergency_contact_name", "s.emergency_contact_phone",
			"s.apps", "s.notes", "s.created_at", "s.updated_at").
		From("hr_staff s").
		LeftJoin("hr_departments d ON d.id = s.department_id AND d.owner_id = s.owner_id").
		Where(sq.Eq{"s.owner_id": ownerID})

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
			&s.ID, &s.OwnerID, &userID, &departmentID_, &departmentName,
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
		StaffID: sql.NullInt64{Int64: staffID, Valid: true},
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
		StaffID: sql.NullInt64{Int64: staffID, Valid: true},
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
