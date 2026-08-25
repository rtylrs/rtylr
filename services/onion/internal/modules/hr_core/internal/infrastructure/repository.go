package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_core/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// isDuplicateUserIDError reports whether err is a MySQL duplicate-entry error
// on hr_employees' global uniq_employee_user_id constraint - the case where a
// caller tries to link a user_id already linked to another employee record
// (possibly in a different business, since the constraint isn't tenant-scoped).
func isDuplicateUserIDError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 &&
		strings.Contains(mysqlErr.Message, "uniq_employee_user_id")
}

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

// --- Employees ---

func (r *Repository) ListEmployees(ctx context.Context, businessID int64, params domain.ListEmployeesParams) ([]domain.EmployeeRecord, int64, error) {
	total, err := r.queries.CountEmployees(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to count employees", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListEmployees(ctx, sqlc.ListEmployeesParams{
		BusinessID: businessID,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list employees", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	employees := make([]domain.EmployeeRecord, 0, len(rows))
	for _, row := range rows {
		employees = append(employees, employeeFromRow(row))
	}
	return employees, total, nil
}

func (r *Repository) GetEmployee(ctx context.Context, businessID, employeeID int64) (*domain.EmployeeRecord, error) {
	row, err := r.queries.GetEmployee(ctx, sqlc.GetEmployeeParams{BusinessID: businessID, ID: employeeID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrEmployeeNotFound
		}
		logger.Ctx(ctx).Error("failed to get employee", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return nil, err
	}
	emp := employeeFromRow(row)
	return &emp, nil
}

func (r *Repository) GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error) {
	id, err := r.queries.GetEmployeeIDByUserID(ctx, sqlc.GetEmployeeIDByUserIDParams{BusinessID: businessID, UserID: sql.NullInt64{Int64: userID, Valid: true}})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrEmployeeNotFound
		}
		logger.Ctx(ctx).Error("failed to get employee by user id", logger.Any("business_id", businessID), logger.Any("user_id", userID), logger.Error(err))
		return 0, err
	}
	return id, nil
}

// ListActiveLeaveTypeIDs is intentionally kept in hr_core because employee
// creation owns the transaction that seeds the employee's balance rows. The
// leave module remains independent of hr_core; this read is the small shared
// boundary required to avoid creating employees with no balances.
func (r *Repository) ListActiveLeaveTypeIDs(ctx context.Context, businessID int64) ([]int64, error) {
	rows, err := r.pool.QueryContext(ctx,
		"SELECT id FROM hr_leave_types WHERE business_id = ? AND is_active = TRUE ORDER BY id",
		businessID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *Repository) CreateEmployee(ctx context.Context, businessID int64, emp domain.EmployeeRecord) (*domain.EmployeeRecord, error) {
	result, err := r.queries.CreateEmployee(ctx, employeeToParams(businessID, emp))
	if err != nil {
		if isDuplicateUserIDError(err) {
			return nil, domain.ErrUserAlreadyLinked
		}
		logger.Ctx(ctx).Error("failed to create employee", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	emp.ID = id
	emp.OwnerID = businessID
	emp.BusinessID = businessID
	return &emp, nil
}

func (r *Repository) UpdateEmployee(ctx context.Context, businessID int64, emp domain.EmployeeRecord) (*domain.EmployeeRecord, error) {
	err := r.queries.UpdateEmployee(ctx, sqlc.UpdateEmployeeParams{
		UserID:                nullInt64(emp.UserID),
		DepartmentID:          nullInt64(emp.DepartmentID),
		LocationID:            nullInt64(emp.LocationID),
		WorkScheduleID:        nullInt64(emp.WorkScheduleID),
		FirstName:             emp.FirstName,
		LastName:              emp.LastName,
		Email:                 emp.Email,
		Phone:                 nullString(emp.Phone),
		EmploymentType:        emp.EmploymentType,
		Status:                emp.Status,
		Salary:                centsToNullDecimalString(emp.SalaryCents),
		HourlyRate:            centsToNullDecimalString(emp.HourlyRateCents),
		OvertimeRate:          centsToNullDecimalString(emp.OvertimeRateCents),
		StartDate:             nullTime(emp.StartDate),
		EndDate:               nullTime(emp.EndDate),
		Address:               nullString(emp.Address),
		EmergencyContactName:  nullString(emp.EmergencyContactName),
		EmergencyContactPhone: nullString(emp.EmergencyContactPhone),
		Notes:                 nullString(emp.Notes),
		ID:                    emp.ID,
		BusinessID:            businessID,
	})
	if err != nil {
		if isDuplicateUserIDError(err) {
			return nil, domain.ErrUserAlreadyLinked
		}
		logger.Ctx(ctx).Error("failed to update employee", logger.Any("business_id", businessID), logger.Any("employee_id", emp.ID), logger.Error(err))
		return nil, err
	}
	emp.OwnerID = businessID
	emp.BusinessID = businessID
	return &emp, nil
}

func (r *Repository) CreateEmployeeWithBalances(ctx context.Context, businessID int64, emp domain.EmployeeRecord, leaveTypeIDs []int64) (*domain.EmployeeRecord, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	result, err := qtx.CreateEmployee(ctx, employeeToParams(businessID, emp))
	if err != nil {
		_ = tx.Rollback()
		if isDuplicateUserIDError(err) {
			return nil, domain.ErrUserAlreadyLinked
		}
		logger.Ctx(ctx).Error("failed to create employee in tx", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}

	employeeID, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	currentYear := int32(time.Now().Year())
	for _, leaveTypeID := range leaveTypeIDs {
		err := qtx.CreateLeaveBalance(ctx, sqlc.CreateLeaveBalanceParams{
			OwnerID:     businessID,
			BusinessID:  businessID,
			EmployeeID:  employeeID,
			LeaveTypeID: leaveTypeID,
			Year:        currentYear,
		})
		if err != nil {
			_ = tx.Rollback()
			logger.Ctx(ctx).Error("failed to create leave balance", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	emp.ID = employeeID
	emp.OwnerID = businessID
	emp.BusinessID = businessID
	return &emp, nil
}

// --- Invites ---

func (r *Repository) CreateInviteToken(ctx context.Context, businessID, employeeID int64, expiresAt time.Time) (string, error) {
	token := uuid.New().String()
	_, err := r.queries.CreateInviteToken(ctx, sqlc.CreateInviteTokenParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		EmployeeID: employeeID,
		Token:      token,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create invite token", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return "", err
	}
	return token, nil
}

func (r *Repository) GetInviteByToken(ctx context.Context, token string) (*domain.InviteRecord, error) {
	row, err := r.queries.GetInviteByToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInviteTokenInvalid
		}
		return nil, err
	}
	return &domain.InviteRecord{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		EmployeeID: row.EmployeeID,
		Token:      row.Token,
		ExpiresAt:  row.ExpiresAt,
		CreatedAt:  row.CreatedAt,
	}, nil
}

func (r *Repository) ConsumeInviteToken(ctx context.Context, token string) error {
	return r.queries.ConsumeInviteToken(ctx, token)
}

// --- Locations ---

func (r *Repository) ListLocations(ctx context.Context, businessID int64) ([]domain.LocationRecord, error) {
	rows, err := r.queries.ListLocations(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list locations", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	locations := make([]domain.LocationRecord, 0, len(rows))
	for _, row := range rows {
		locations = append(locations, locationFromRow(row))
	}
	return locations, nil
}

func (r *Repository) GetLocation(ctx context.Context, businessID, locationID int64) (*domain.LocationRecord, error) {
	row, err := r.queries.GetLocation(ctx, sqlc.GetLocationParams{BusinessID: businessID, ID: locationID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrLocationNotFound
		}
		return nil, err
	}
	loc := locationFromRow(row)
	return &loc, nil
}

func (r *Repository) CreateLocation(ctx context.Context, businessID int64, loc domain.LocationRecord) (*domain.LocationRecord, error) {
	result, err := r.queries.CreateLocation(ctx, sqlc.CreateLocationParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		Name:       loc.Name,
		Address:    nullString(loc.Address),
		City:       nullString(loc.City),
		Country:    nullString(loc.Country),
		Timezone:   nullString(loc.Timezone),
		Latitude:   nullFloat64(loc.Latitude),
		Longitude:  nullFloat64(loc.Longitude),
		IsDefault:  loc.IsDefault,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create location", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	loc.ID = id
	loc.OwnerID = businessID
	loc.BusinessID = businessID
	return &loc, nil
}

func (r *Repository) UpdateLocation(ctx context.Context, businessID int64, loc domain.LocationRecord) (*domain.LocationRecord, error) {
	err := r.queries.UpdateLocation(ctx, sqlc.UpdateLocationParams{
		Name:       loc.Name,
		Address:    nullString(loc.Address),
		City:       nullString(loc.City),
		Country:    nullString(loc.Country),
		Timezone:   nullString(loc.Timezone),
		Latitude:   nullFloat64(loc.Latitude),
		Longitude:  nullFloat64(loc.Longitude),
		IsDefault:  loc.IsDefault,
		ID:         loc.ID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to update location", logger.Any("business_id", businessID), logger.Any("location_id", loc.ID), logger.Error(err))
		return nil, err
	}
	loc.OwnerID = businessID
	loc.BusinessID = businessID
	return &loc, nil
}

// --- Departments ---

func (r *Repository) ListDepartments(ctx context.Context, businessID, _ int64) ([]domain.DepartmentRecord, error) {
	rows, err := r.queries.ListHRDepartments(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list departments", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	departments := make([]domain.DepartmentRecord, 0, len(rows))
	for _, row := range rows {
		departments = append(departments, departmentFromRow(row))
	}
	return departments, nil
}

func (r *Repository) CreateDepartment(ctx context.Context, businessID int64, dept domain.DepartmentRecord) (*domain.DepartmentRecord, error) {
	result, err := r.queries.CreateHRDepartment(ctx, sqlc.CreateHRDepartmentParams{
		OwnerID:               businessID,
		BusinessID:            businessID,
		LocationID:            dept.LocationID,
		ParentDepartmentID:    nullInt64(dept.ParentDepartmentID),
		Name:                  dept.Name,
		DefaultWorkScheduleID: nullInt64(dept.DefaultWorkScheduleID),
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create department", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	dept.ID = id
	dept.OwnerID = businessID
	dept.BusinessID = businessID
	return &dept, nil
}

func (r *Repository) UpdateDepartment(ctx context.Context, businessID int64, dept domain.DepartmentRecord) (*domain.DepartmentRecord, error) {
	err := r.queries.UpdateHRDepartment(ctx, sqlc.UpdateHRDepartmentParams{
		LocationID:            dept.LocationID,
		ParentDepartmentID:    nullInt64(dept.ParentDepartmentID),
		Name:                  dept.Name,
		DefaultWorkScheduleID: nullInt64(dept.DefaultWorkScheduleID),
		ID:                    dept.ID,
		BusinessID:            businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to update department", logger.Any("business_id", businessID), logger.Any("department_id", dept.ID), logger.Error(err))
		return nil, err
	}
	dept.OwnerID = businessID
	dept.BusinessID = businessID
	return &dept, nil
}

// --- Work Schedules ---

func (r *Repository) ListWorkSchedules(ctx context.Context, businessID int64, limit, offset int32) ([]domain.WorkScheduleRecord, error) {
	rows, err := r.queries.ListWorkSchedules(ctx, sqlc.ListWorkSchedulesParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list work schedules", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	schedules := make([]domain.WorkScheduleRecord, 0, len(rows))
	for _, row := range rows {
		schedules = append(schedules, workScheduleFromRow(row))
	}
	return schedules, nil
}

// CountWorkSchedules returns the tenant's total, for the list response's
// pagination maths.
func (r *Repository) CountWorkSchedules(ctx context.Context, businessID int64) (int64, error) {
	total, err := r.queries.CountWorkSchedules(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to count work schedules", logger.Any("business_id", businessID), logger.Error(err))
		return 0, err
	}

	return total, nil
}

func (r *Repository) GetWorkScheduleForEmployee(ctx context.Context, businessID, employeeID int64) (*domain.WorkScheduleRecord, error) {
	row, err := r.queries.GetWorkScheduleForEmployee(ctx, sqlc.GetWorkScheduleForEmployeeParams{BusinessID: businessID, ID: employeeID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	ws := workScheduleFromRow(row)
	return &ws, nil
}

func (r *Repository) CreateWorkSchedule(ctx context.Context, businessID int64, ws domain.WorkScheduleRecord) (*domain.WorkScheduleRecord, error) {
	// schedule_data is a JSON column, so an empty string is not valid input:
	// MySQL rejects it with "Invalid JSON text: The document is empty", which
	// reaches the caller as a 500 rather than a validation error. An absent
	// schedule is an empty object.
	ws.ScheduleData = defaultedScheduleData(ws.ScheduleData)

	result, err := r.queries.CreateWorkSchedule(ctx, sqlc.CreateWorkScheduleParams{
		OwnerID:      businessID,
		BusinessID:   businessID,
		Name:         ws.Name,
		ScheduleData: json.RawMessage(ws.ScheduleData),
		IsDefault:    ws.IsDefault,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create work schedule", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	ws.ID = id
	ws.OwnerID = businessID
	ws.BusinessID = businessID
	return &ws, nil
}

func (r *Repository) UpdateWorkSchedule(ctx context.Context, businessID int64, ws domain.WorkScheduleRecord) (*domain.WorkScheduleRecord, error) {
	ws.ScheduleData = defaultedScheduleData(ws.ScheduleData)

	err := r.queries.UpdateWorkSchedule(ctx, sqlc.UpdateWorkScheduleParams{
		Name:         ws.Name,
		ScheduleData: json.RawMessage(ws.ScheduleData),
		IsDefault:    ws.IsDefault,
		ID:           ws.ID,
		BusinessID:   businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to update work schedule", logger.Any("business_id", businessID), logger.Any("schedule_id", ws.ID), logger.Error(err))
		return nil, err
	}
	ws.OwnerID = businessID
	ws.BusinessID = businessID
	return &ws, nil
}

// --- Roles & Permissions ---

func (r *Repository) ListRoles(ctx context.Context, businessID int64) ([]domain.RoleRecord, error) {
	rows, err := r.queries.ListRoles(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list roles", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	roles := make([]domain.RoleRecord, 0, len(rows))
	for _, row := range rows {
		roles = append(roles, roleFromRow(row))
	}
	return roles, nil
}

func (r *Repository) CreateRole(ctx context.Context, businessID int64, role domain.RoleRecord) (*domain.RoleRecord, error) {
	result, err := r.queries.CreateRole(ctx, sqlc.CreateRoleParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		Name:        role.Name,
		Description: nullString(role.Description),
		Permissions: json.RawMessage(role.Permissions),
		IsSystem:    role.IsSystem,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create role", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	role.ID = id
	role.OwnerID = businessID
	role.BusinessID = businessID
	return &role, nil
}

func (r *Repository) UpdateRole(ctx context.Context, businessID int64, role domain.RoleRecord) (*domain.RoleRecord, error) {
	err := r.queries.UpdateRole(ctx, sqlc.UpdateRoleParams{
		Name:        role.Name,
		Description: nullString(role.Description),
		Permissions: json.RawMessage(role.Permissions),
		ID:          role.ID,
		BusinessID:  businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to update role", logger.Any("business_id", businessID), logger.Any("role_id", role.ID), logger.Error(err))
		return nil, err
	}
	role.OwnerID = businessID
	role.BusinessID = businessID
	return &role, nil
}

func (r *Repository) AssignRole(ctx context.Context, businessID, employeeID, roleID int64) error {
	return r.queries.AssignRole(ctx, sqlc.AssignRoleParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		EmployeeID: employeeID,
		RoleID:     roleID,
	})
}

func (r *Repository) HasPermission(ctx context.Context, businessID, employeeID int64, permissionKey string) (bool, error) {
	roles, err := r.queries.GetEmployeeRoles(ctx, sqlc.GetEmployeeRolesParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
	})
	if err != nil {
		return false, err
	}
	for _, role := range roles {
		var perms map[string]bool
		if err := json.Unmarshal(role.Permissions, &perms); err != nil {
			continue
		}
		if perms[permissionKey] {
			return true, nil
		}
	}
	return false, nil
}

func (r *Repository) HasRole(ctx context.Context, businessID, employeeID, roleID int64) (bool, error) {
	count, err := r.queries.HasEmployeeRole(ctx, sqlc.HasEmployeeRoleParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		RoleID:     roleID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// --- Setup Wizard ---

func (r *Repository) GetSetupSteps(ctx context.Context, businessID int64) ([]domain.SetupStepRecord, error) {
	rows, err := r.queries.GetSetupSteps(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to get setup steps", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	steps := make([]domain.SetupStepRecord, 0, len(rows))
	for _, row := range rows {
		steps = append(steps, setupStepFromRow(row))
	}
	return steps, nil
}

func (r *Repository) CompleteSetupStep(ctx context.Context, businessID int64, stepKey string) (*domain.SetupStepRecord, error) {
	err := r.queries.CompleteSetupStep(ctx, sqlc.CompleteSetupStepParams{
		BusinessID: businessID,
		StepKey:    stepKey,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to complete setup step", logger.Any("business_id", businessID), logger.String("step_key", stepKey), logger.Error(err))
		return nil, err
	}
	row, err := r.queries.GetSetupStepByKey(ctx, sqlc.GetSetupStepByKeyParams{
		BusinessID: businessID,
		StepKey:    stepKey,
	})
	if err != nil {
		return nil, err
	}
	step := setupStepFromRow(row)
	return &step, nil
}

func (r *Repository) GetHRConfig(ctx context.Context, businessID int64) (*domain.HRCompanyConfigRecord, error) {
	row, err := r.queries.GetHRConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSetupNotComplete
		}
		logger.Ctx(ctx).Error("failed to get hr config", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	return &domain.HRCompanyConfigRecord{
		ID:                    row.ID,
		OwnerID:               row.OwnerID,
		BusinessID:            row.BusinessID,
		CompanyName:           row.CompanyName,
		FiscalYearStartMonth:  row.FiscalYearStartMonth,
		DefaultCurrency:       row.DefaultCurrency,
		DefaultTimezone:       row.DefaultTimezone,
		DefaultWorkScheduleID: row.DefaultWorkScheduleID.Int64,
		OvertimeEnabled:       row.OvertimeEnabled,
		OvertimeMultiplier:    parseFloat(row.OvertimeMultiplier),
		DirectoryEnabled:      row.DirectoryEnabled,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}, nil
}

func (r *Repository) UpsertHRConfig(ctx context.Context, businessID int64, cfg domain.HRCompanyConfigRecord) (*domain.HRCompanyConfigRecord, error) {
	_, err := r.queries.UpsertHRConfig(ctx, sqlc.UpsertHRConfigParams{
		OwnerID:               businessID,
		BusinessID:            businessID,
		CompanyName:           cfg.CompanyName,
		FiscalYearStartMonth:  cfg.FiscalYearStartMonth,
		DefaultCurrency:       cfg.DefaultCurrency,
		DefaultTimezone:       cfg.DefaultTimezone,
		DefaultWorkScheduleID: nullInt64(cfg.DefaultWorkScheduleID),
		OvertimeEnabled:       cfg.OvertimeEnabled,
		OvertimeMultiplier:    strconv.FormatFloat(cfg.OvertimeMultiplier, 'f', 2, 64),
		DirectoryEnabled:      cfg.DirectoryEnabled,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to upsert hr config", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	return r.GetHRConfig(ctx, businessID)
}

// SeedDefaultRoles inserts any of the given roles that don't already exist for
// this business (matched by name). It never touches a role that already
// exists. A business's customization via UpdateRole must survive every
// subsequent hire, not just the first one.
func (r *Repository) SeedDefaultRoles(ctx context.Context, businessID int64, roles []domain.RoleRecord) error {
	for _, role := range roles {
		_, err := r.queries.InsertRoleIfMissing(ctx, sqlc.InsertRoleIfMissingParams{
			OwnerID:     businessID,
			BusinessID:  businessID,
			Name:        role.Name,
			Description: nullString(role.Description),
			Permissions: json.RawMessage(role.Permissions),
			IsSystem:    role.IsSystem,
		})
		if err != nil {
			logger.Ctx(ctx).Error("failed to seed role", logger.Any("business_id", businessID), logger.String("role", role.Name), logger.Error(err))
			return err
		}
	}
	return nil
}

// --- Approval Chains ---

func (r *Repository) GetApprovalChain(ctx context.Context, businessID int64, requestType string) (*domain.ApprovalChainRecord, error) {
	chain, err := r.queries.GetApprovalChain(ctx, sqlc.GetApprovalChainParams{
		BusinessID:  businessID,
		RequestType: requestType,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	levelRows, err := r.queries.GetApprovalChainLevels(ctx, sqlc.GetApprovalChainLevelsParams{
		BusinessID:      businessID,
		ApprovalChainID: chain.ID,
	})
	if err != nil {
		return nil, err
	}

	levels := make([]domain.ApprovalChainLevelRecord, 0, len(levelRows))
	for _, lv := range levelRows {
		levels = append(levels, domain.ApprovalChainLevelRecord{
			ID:                 lv.ID,
			OwnerID:            lv.OwnerID,
			BusinessID:         lv.BusinessID,
			ApprovalChainID:    lv.ApprovalChainID,
			LevelOrder:         lv.LevelOrder,
			ApproverType:       lv.ApproverType,
			ApproverRoleID:     lv.ApproverRoleID.Int64,
			ApproverEmployeeID: lv.ApproverEmployeeID.Int64,
			CreatedAt:          lv.CreatedAt,
		})
	}

	return &domain.ApprovalChainRecord{
		ID:          chain.ID,
		OwnerID:     chain.OwnerID,
		BusinessID:  chain.BusinessID,
		RequestType: chain.RequestType,
		Name:        chain.Name,
		Levels:      levels,
		CreatedAt:   chain.CreatedAt,
		UpdatedAt:   chain.UpdatedAt,
	}, nil
}

func (r *Repository) UpsertApprovalChain(ctx context.Context, businessID int64, chain domain.ApprovalChainRecord) (*domain.ApprovalChainRecord, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	result, err := qtx.UpsertApprovalChain(ctx, sqlc.UpsertApprovalChainParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		RequestType: chain.RequestType,
		Name:        chain.Name,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	chainID, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if chainID == 0 {
		existing, lookupErr := qtx.GetApprovalChain(ctx, sqlc.GetApprovalChainParams{
			BusinessID:  businessID,
			RequestType: chain.RequestType,
		})
		if lookupErr != nil {
			_ = tx.Rollback()
			return nil, lookupErr
		}
		chainID = existing.ID
	}

	err = qtx.DeleteApprovalChainLevels(ctx, sqlc.DeleteApprovalChainLevelsParams{
		BusinessID:      businessID,
		ApprovalChainID: chainID,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	for _, level := range chain.Levels {
		err := qtx.CreateApprovalChainLevel(ctx, sqlc.CreateApprovalChainLevelParams{
			OwnerID:            businessID,
			BusinessID:         businessID,
			ApprovalChainID:    chainID,
			LevelOrder:         level.LevelOrder,
			ApproverType:       level.ApproverType,
			ApproverRoleID:     nullInt64(level.ApproverRoleID),
			ApproverEmployeeID: nullInt64(level.ApproverEmployeeID),
		})
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetApprovalChain(ctx, businessID, chain.RequestType)
}

// --- Helpers ---

func employeeFromRow(row sqlc.HrEmployee) domain.EmployeeRecord {
	return domain.EmployeeRecord{
		ID:                    row.ID,
		OwnerID:               row.OwnerID,
		BusinessID:            row.BusinessID,
		UserID:                row.UserID.Int64,
		DepartmentID:          row.DepartmentID.Int64,
		LocationID:            row.LocationID.Int64,
		WorkScheduleID:        row.WorkScheduleID.Int64,
		FirstName:             row.FirstName,
		LastName:              row.LastName,
		Email:                 row.Email,
		Phone:                 row.Phone.String,
		EmploymentType:        row.EmploymentType,
		Status:                row.Status,
		SalaryCents:           nullDecimalStringToCents(row.Salary),
		HourlyRateCents:       nullDecimalStringToCents(row.HourlyRate),
		OvertimeRateCents:     nullDecimalStringToCents(row.OvertimeRate),
		StartDate:             row.StartDate.Time,
		EndDate:               row.EndDate.Time,
		Address:               row.Address.String,
		EmergencyContactName:  row.EmergencyContactName.String,
		EmergencyContactPhone: row.EmergencyContactPhone.String,
		Notes:                 row.Notes.String,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}

func employeeToParams(businessID int64, emp domain.EmployeeRecord) sqlc.CreateEmployeeParams {
	return sqlc.CreateEmployeeParams{
		OwnerID:               businessID,
		BusinessID:            businessID,
		UserID:                nullInt64(emp.UserID),
		DepartmentID:          nullInt64(emp.DepartmentID),
		LocationID:            nullInt64(emp.LocationID),
		WorkScheduleID:        nullInt64(emp.WorkScheduleID),
		FirstName:             emp.FirstName,
		LastName:              emp.LastName,
		Email:                 emp.Email,
		Phone:                 nullString(emp.Phone),
		EmploymentType:        emp.EmploymentType,
		Status:                emp.Status,
		Salary:                centsToNullDecimalString(emp.SalaryCents),
		HourlyRate:            centsToNullDecimalString(emp.HourlyRateCents),
		OvertimeRate:          centsToNullDecimalString(emp.OvertimeRateCents),
		StartDate:             nullTime(emp.StartDate),
		EndDate:               nullTime(emp.EndDate),
		Address:               nullString(emp.Address),
		EmergencyContactName:  nullString(emp.EmergencyContactName),
		EmergencyContactPhone: nullString(emp.EmergencyContactPhone),
		Notes:                 nullString(emp.Notes),
	}
}

func locationFromRow(row sqlc.HrLocation) domain.LocationRecord {
	return domain.LocationRecord{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		Name:       row.Name,
		Address:    row.Address.String,
		City:       row.City.String,
		Country:    row.Country.String,
		Timezone:   row.Timezone.String,
		Latitude:   row.Latitude.Float64,
		Longitude:  row.Longitude.Float64,
		IsDefault:  row.IsDefault,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}

func departmentFromRow(row sqlc.HrDepartment) domain.DepartmentRecord {
	return domain.DepartmentRecord{
		ID:                    row.ID,
		OwnerID:               row.OwnerID,
		BusinessID:            row.BusinessID,
		LocationID:            row.LocationID,
		ParentDepartmentID:    row.ParentDepartmentID.Int64,
		Name:                  row.Name,
		DefaultWorkScheduleID: row.DefaultWorkScheduleID.Int64,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}

func workScheduleFromRow(row sqlc.HrWorkSchedule) domain.WorkScheduleRecord {
	return domain.WorkScheduleRecord{
		ID:           row.ID,
		OwnerID:      row.OwnerID,
		BusinessID:   row.BusinessID,
		Name:         row.Name,
		ScheduleData: string(row.ScheduleData),
		IsDefault:    row.IsDefault,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func roleFromRow(row sqlc.HrRole) domain.RoleRecord {
	return domain.RoleRecord{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		Name:        row.Name,
		Description: row.Description.String,
		Permissions: string(row.Permissions),
		IsSystem:    row.IsSystem,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func setupStepFromRow(row sqlc.HrSetupStep) domain.SetupStepRecord {
	return domain.SetupStepRecord{
		ID:          row.ID,
		OwnerID:     row.OwnerID,
		BusinessID:  row.BusinessID,
		StepKey:     row.StepKey,
		StepLabel:   row.StepLabel,
		IsCompleted: row.IsCompleted,
		CompletedAt: row.CompletedAt.Time,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

func nullString(v string) sql.NullString {
	return sql.NullString{String: v, Valid: v != ""}
}

// centsToNullDecimalString renders an integer cents amount as the DECIMAL(10,2)
// string sqlc expects, using only integer arithmetic so no float64 rounding
// error can be introduced on the way into the database. A zero amount is
// stored as NULL, matching the previous float-based helper's behaviour.
func centsToNullDecimalString(cents int64) sql.NullString {
	if cents == 0 {
		return sql.NullString{}
	}
	return sql.NullString{String: centsToDecimalString(cents), Valid: true}
}

func centsToDecimalString(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}
	whole := cents / 100
	fraction := cents % 100
	s := strconv.FormatInt(whole, 10) + "." + fmt.Sprintf("%02d", fraction)
	if negative {
		s = "-" + s
	}
	return s
}

func nullFloat64(v float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: v, Valid: v != 0}
}

func nullTime(v time.Time) sql.NullTime {
	return sql.NullTime{Time: v, Valid: !v.IsZero()}
}

// nullDecimalStringToCents parses a MySQL DECIMAL(10,2) value read back as a
// string into integer cents. It works on the string digits directly rather
// than through strconv.ParseFloat, so no float64 rounding error can be
// introduced on the way out of the database.
func nullDecimalStringToCents(v sql.NullString) int64 {
	if !v.Valid {
		return 0
	}
	cents, err := decimalStringToCents(v.String)
	if err != nil {
		// DECIMAL values from MySQL should always parse; log corruption and return zero
		logger.Global().Warn("failed to parse nullable decimal from database", logger.String("value", v.String), logger.Error(err))
		return 0
	}
	return cents
}

func decimalStringToCents(s string) (int64, error) {
	s = strings.TrimSpace(s)
	negative := strings.HasPrefix(s, "-")
	if negative {
		s = s[1:]
	}

	whole, fraction, hasFraction := strings.Cut(s, ".")
	if whole == "" {
		whole = "0"
	}
	if len(fraction) > 2 {
		fraction = fraction[:2]
	}
	for len(fraction) < 2 {
		fraction += "0"
	}
	if !hasFraction {
		fraction = "00"
	}

	wholeCents, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, err
	}
	fractionCents, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, err
	}

	total := wholeCents*100 + fractionCents
	if negative {
		total = -total
	}
	return total, nil
}

func parseFloat(v string) float64 {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		// DECIMAL values from MySQL should always parse; log corruption and return zero
		logger.Global().Warn("failed to parse decimal from database", logger.String("value", v), logger.Error(err))
		return 0
	}
	return f
}

// defaultedScheduleData keeps an absent schedule out of the JSON column as an
// empty object rather than an empty string, which MySQL will not accept.
func defaultedScheduleData(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "{}"
	}

	return raw
}
