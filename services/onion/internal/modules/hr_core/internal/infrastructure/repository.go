package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_core/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

// updateNotFound classifies a zero-affected UPDATE. MySQL reports the number of
// rows changed, not the number matched, so an update that stores the values the
// row already holds (an idempotent PUT) also affects zero rows. Only a probe
// that comes back sql.ErrNoRows means the record is genuinely absent; any other
// probe failure is returned unchanged.
func updateNotFound(probeErr, notFound error) error {
	if errors.Is(probeErr, sql.ErrNoRows) {
		return notFound
	}
	return probeErr
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
	departmentID := sql.NullInt64{}
	if params.DepartmentID != nil {
		departmentID = sqlutil.NullInt64(*params.DepartmentID)
	}
	locationID := sql.NullInt64{}
	if params.LocationID != nil {
		locationID = sqlutil.NullInt64(*params.LocationID)
	}
	status := sql.NullString{}
	if params.Status != nil {
		status = sqlutil.NullString(*params.Status)
	}

	filterParams := sqlc.CountEmployeesParams{
		BusinessID:   businessID,
		DepartmentID: departmentID,
		LocationID:   locationID,
		Status:       status,
	}
	total, err := r.queries.CountEmployees(ctx, filterParams)
	if err != nil {
		logger.Ctx(ctx).Error("failed to count employees", logger.Any("business_id", businessID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListEmployees(ctx, sqlc.ListEmployeesParams{
		BusinessID:   businessID,
		DepartmentID: departmentID,
		LocationID:   locationID,
		Status:       status,
		Limit:        params.Limit,
		Offset:       params.Offset,
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
	affected, err := r.queries.UpdateEmployee(ctx, sqlc.UpdateEmployeeParams{
		UserID:                sqlutil.NullInt64(emp.UserID),
		DepartmentID:          sqlutil.NullInt64(emp.DepartmentID),
		LocationID:            sqlutil.NullInt64(emp.LocationID),
		WorkScheduleID:        sqlutil.NullInt64(emp.WorkScheduleID),
		FirstName:             emp.FirstName,
		LastName:              emp.LastName,
		Email:                 emp.Email,
		Phone:                 sqlutil.NullString(emp.Phone),
		EmploymentType:        emp.EmploymentType,
		Status:                emp.Status,
		Salary:                centsToNullDecimalString(emp.SalaryCents),
		HourlyRate:            centsToNullDecimalString(emp.HourlyRateCents),
		OvertimeRate:          centsToNullDecimalString(emp.OvertimeRateCents),
		StartDate:             nullTime(emp.StartDate),
		EndDate:               nullTime(emp.EndDate),
		Address:               sqlutil.NullString(emp.Address),
		EmergencyContactName:  sqlutil.NullString(emp.EmergencyContactName),
		EmergencyContactPhone: sqlutil.NullString(emp.EmergencyContactPhone),
		Notes:                 sqlutil.NullString(emp.Notes),
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
	if affected == 0 {
		return nil, domain.ErrEmployeeNotFound
	}
	emp.OwnerID = businessID
	emp.BusinessID = businessID
	return &emp, nil
}

func (r *Repository) DeactivateEmployee(ctx context.Context, businessID, employeeID int64, endDate time.Time) error {
	result, err := r.queries.DeactivateEmployee(ctx, sqlc.DeactivateEmployeeParams{
		EndDate:    nullTime(endDate),
		ID:         employeeID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to deactivate employee", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected == 0 {
		return domain.ErrEmployeeNotFound
	}
	return nil
}
