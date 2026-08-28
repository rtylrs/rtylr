package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/staff/internal/domain"
	"github.com/voxire/onion/internal/modules/staff/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

// dynamic-query: squirrel
func (r *Repository) ListStaff(ctx context.Context, businessID int64, search, departmentID, status string) ([]domain.Staff, error) {
	// cross-schema read from hr_departments — to include department name in staff list response
	builder := sq.
		Select("s.id", "s.owner_id", "s.business_id", "s.user_id", "s.department_id",
			"COALESCE(d.name, '') AS department_name",
			"s.first_name", "s.last_name", "s.email", "s.phone",
			"s.role", "s.employment_type", "s.status", "s.salary", "s.hourly_rate", "s.overtime_rate",
			"s.start_date", "s.end_date", "s.address", "s.emergency_contact_name", "s.emergency_contact_phone",
			"s.apps", "s.notes", "s.photo_url", "s.bio", "s.created_at", "s.updated_at").
		From("hr_staff s").
		LeftJoin("hr_departments d ON d.id = s.department_id AND d.business_id = s.business_id").
		Where(sq.Eq{"s.business_id": businessID})

	search = strings.TrimSpace(search)
	if search != "" {
		searchTerm := searchutil.BooleanPrefixQuery(search)
		builder = builder.Where(sq.Or{
			sq.Eq{"s.email": search},
			sq.Eq{"s.phone": search},
			sq.Expr("(? <> '' AND MATCH(s.first_name, s.last_name, s.email, s.role) AGAINST (? IN BOOLEAN MODE))", searchTerm, searchTerm),
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

	return scanStaffRows(rows)
}

// scanStaffRows reads the shared staff-list projection (s.id through
// s.updated_at, plus the joined department name) used by both ListStaff and
// ListStaffDirectory. Keeping the scan in one place means a column added to
// one caller's SELECT is scanned the same way for the other.
func scanStaffRows(rows *sql.Rows) ([]domain.Staff, error) {
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
			photoURL              sql.NullString
			bio                   sql.NullString
			createdAt             sql.NullTime
			updatedAt             sql.NullTime
		)
		if err := rows.Scan(
			&s.ID, &s.OwnerID, &s.BusinessID, &userID, &departmentID_, &departmentName,
			&s.FirstName, &s.LastName, &s.Email, &phone,
			&s.Role, &employmentType, &statusVal, &salary, &hourlyRate, &overtimeRate,
			&s.StartDate, &endDate, &address, &emergencyContactName, &emergencyContactPhone,
			&apps, &notes, &photoURL, &bio, &createdAt, &updatedAt,
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
		if photoURL.Valid {
			s.PhotoURL = photoURL.String
		}
		if bio.Valid {
			s.Bio = bio.String
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

// ListStaffDirectory returns one page of the business-wide staff list,
// ordered the same way as ListStaff. Unlike ListStaff it takes no search/
// department/status filters - the directory is deliberately the plain
// roster, not the admin screen's filtered view.
func (r *Repository) ListStaffDirectory(ctx context.Context, businessID, limit, offset int64) ([]domain.Staff, error) {
	query, args, err := sq.
		Select("s.id", "s.owner_id", "s.business_id", "s.user_id", "s.department_id",
			"COALESCE(d.name, '') AS department_name",
			"s.first_name", "s.last_name", "s.email", "s.phone",
			"s.role", "s.employment_type", "s.status", "s.salary", "s.hourly_rate", "s.overtime_rate",
			"s.start_date", "s.end_date", "s.address", "s.emergency_contact_name", "s.emergency_contact_phone",
			"s.apps", "s.notes", "s.photo_url", "s.bio", "s.created_at", "s.updated_at").
		From("hr_staff s").
		LeftJoin("hr_departments d ON d.id = s.department_id AND d.business_id = s.business_id").
		Where(sq.Eq{"s.business_id": businessID}).
		OrderBy("s.last_name ASC", "s.first_name ASC").
		Limit(uint64(limit)).   // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset(uint64(offset)). // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build staff directory query", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list staff directory", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	return scanStaffRows(rows)
}

// CountStaffDirectory returns the total staff count for a business, used
// alongside ListStaffDirectory to build the paginated response's total.
func (r *Repository) CountStaffDirectory(ctx context.Context, businessID int64) (int64, error) {
	var total int64
	err := r.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM hr_staff WHERE business_id = ?", businessID,
	).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count staff directory", logger.Error(err))
		return 0, err
	}
	return total, nil
}

func (r *Repository) GetStaffByID(ctx context.Context, businessID, staffID int64) (*domain.Staff, error) {
	row, err := r.queries.GetStaffByID(ctx, sqlc.GetStaffByIDParams{
		ID:         staffID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrStaffNotFound
		}
		return nil, err
	}
	s := staffFromGetRow(row)
	return &s, nil
}

func staffFromGetRow(row sqlc.GetStaffByIDRow) domain.Staff {
	var s domain.Staff
	s.ID = row.ID
	s.OwnerID = row.OwnerID
	s.BusinessID = row.BusinessID
	if row.UserID.Valid {
		s.UserID = row.UserID.Int64
	}
	if row.ManagerStaffID.Valid {
		s.ManagerStaffID = row.ManagerStaffID.Int64
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
		s.Salary = sqlutil.ParseDecimal(row.Salary.String)
	}
	if row.HourlyRate.Valid {
		s.HourlyRate = sqlutil.ParseDecimal(row.HourlyRate.String)
	}
	if row.OvertimeRate.Valid {
		s.OvertimeRate = sqlutil.ParseDecimal(row.OvertimeRate.String)
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
	if row.PhotoUrl.Valid {
		s.PhotoURL = row.PhotoUrl.String
	}
	if row.Bio.Valid {
		s.Bio = row.Bio.String
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
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Ctx(ctx).Error("failed to rollback after create staff panic", logger.Error(rollbackErr))
			}
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	result, err := qtx.CreateStaff(ctx, sqlc.CreateStaffParams{
		OwnerID:      input.OwnerID,
		BusinessID:   input.BusinessID,
		UserID:       sql.NullInt64{Int64: input.UserID, Valid: input.UserID != 0},
		DepartmentID: sql.NullInt64{Int64: input.DepartmentID, Valid: input.DepartmentID != 0},
		// Zero means no direct reporting line, which is the common case, so it
		// stores NULL rather than a manager with staff id 0.
		ManagerStaffID:        sql.NullInt64{Int64: input.ManagerStaffID, Valid: input.ManagerStaffID != 0},
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
		PhotoUrl:              sql.NullString{String: input.PhotoURL, Valid: input.PhotoURL != ""},
		Bio:                   sql.NullString{String: input.Bio, Valid: input.Bio != ""},
	})
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback after create staff", logger.Error(rollbackErr))
		}
		logger.Ctx(ctx).Error("Failed to create staff", logger.Error(err))
		return 0, err
	}

	staffID, err := result.LastInsertId()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback after create staff id", logger.Error(rollbackErr))
		}
		return 0, err
	}

	if err := r.syncEmployeeFromStaff(ctx, qtx, input.BusinessID, staffID); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback after staff sync", logger.Error(rollbackErr))
		}
		logger.Ctx(ctx).Error("Failed to sync created staff to HR employee", logger.Error(err))
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return staffID, nil
}

func (r *Repository) UpdateStaff(ctx context.Context, businessID int64, input domain.StaffUpdateInput) error {
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
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Ctx(ctx).Error("failed to rollback after update staff panic", logger.Error(rollbackErr))
			}
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	err = qtx.UpdateStaff(ctx, sqlc.UpdateStaffParams{
		DepartmentID: sql.NullInt64{Int64: input.DepartmentID, Valid: input.DepartmentID != 0},
		// Zero means no direct reporting line, which is the common case, so it
		// stores NULL rather than a manager with staff id 0.
		ManagerStaffID:        sql.NullInt64{Int64: input.ManagerStaffID, Valid: input.ManagerStaffID != 0},
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
		PhotoUrl:              sql.NullString{String: input.PhotoURL, Valid: input.PhotoURL != ""},
		Bio:                   sql.NullString{String: input.Bio, Valid: input.Bio != ""},
		ID:                    input.ID,
		BusinessID:            businessID,
	})
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback after update staff", logger.Error(rollbackErr))
		}
		logger.Ctx(ctx).Error("Failed to update staff", logger.Error(err))
		return err
	}

	if err := r.syncEmployeeFromStaff(ctx, qtx, businessID, input.ID); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback after staff sync", logger.Error(rollbackErr))
		}
		logger.Ctx(ctx).Error("Failed to sync updated staff to HR employee", logger.Error(err))
		return err
	}

	return tx.Commit()
}

func (r *Repository) DeleteStaff(ctx context.Context, businessID, staffID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Ctx(ctx).Error("failed to rollback after delete staff panic", logger.Error(rollbackErr))
			}
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	err = qtx.TerminateStaff(ctx, sqlc.TerminateStaffParams{
		ID:         staffID,
		BusinessID: businessID,
	})
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback after terminate staff", logger.Error(rollbackErr))
		}
		logger.Ctx(ctx).Error("Failed to terminate staff", logger.Error(err))
		return err
	}

	if err := r.syncEmployeeFromStaff(ctx, qtx, businessID, staffID); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback after staff sync", logger.Error(rollbackErr))
		}
		logger.Ctx(ctx).Error("Failed to sync terminated staff to HR employee", logger.Error(err))
		return err
	}

	return tx.Commit()
}

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
