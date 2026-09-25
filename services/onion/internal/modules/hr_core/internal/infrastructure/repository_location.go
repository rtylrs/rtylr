package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_core/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func (r *Repository) CreateLocation(ctx context.Context, businessID int64, loc domain.LocationRecord) (*domain.LocationRecord, error) {
	result, err := r.queries.CreateLocation(ctx, sqlc.CreateLocationParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		Name:       loc.Name,
		Address:    sqlutil.NullString(loc.Address),
		City:       sqlutil.NullString(loc.City),
		Country:    sqlutil.NullString(loc.Country),
		Timezone:   sqlutil.NullString(loc.Timezone),
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
	result, err := r.queries.UpdateLocation(ctx, sqlc.UpdateLocationParams{
		Name:       loc.Name,
		Address:    sqlutil.NullString(loc.Address),
		City:       sqlutil.NullString(loc.City),
		Country:    sqlutil.NullString(loc.Country),
		Timezone:   sqlutil.NullString(loc.Timezone),
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
	if affected, err := result.RowsAffected(); err != nil {
		return nil, err
	} else if affected == 0 {
		if _, getErr := r.queries.GetLocation(ctx, sqlc.GetLocationParams{BusinessID: businessID, ID: loc.ID}); getErr != nil {
			return nil, updateNotFound(getErr, domain.ErrLocationNotFound)
		}
	}
	loc.OwnerID = businessID
	loc.BusinessID = businessID
	return &loc, nil
}

// --- Departments ---

func (r *Repository) ListDepartments(ctx context.Context, businessID, locationID int64) ([]domain.DepartmentRecord, error) {
	rows, err := r.queries.ListHRDepartments(ctx, sqlc.ListHRDepartmentsParams{
		BusinessID: businessID,
		LocationID: locationID,
	})
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
		ParentDepartmentID:    sqlutil.NullInt64(dept.ParentDepartmentID),
		Name:                  dept.Name,
		DefaultWorkScheduleID: sqlutil.NullInt64(dept.DefaultWorkScheduleID),
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
	result, err := r.queries.UpdateHRDepartment(ctx, sqlc.UpdateHRDepartmentParams{
		LocationID:            dept.LocationID,
		ParentDepartmentID:    sqlutil.NullInt64(dept.ParentDepartmentID),
		Name:                  dept.Name,
		DefaultWorkScheduleID: sqlutil.NullInt64(dept.DefaultWorkScheduleID),
		ID:                    dept.ID,
		BusinessID:            businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to update department", logger.Any("business_id", businessID), logger.Any("department_id", dept.ID), logger.Error(err))
		return nil, err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return nil, err
	} else if affected == 0 {
		if _, getErr := r.queries.GetHRDepartment(ctx, sqlc.GetHRDepartmentParams{BusinessID: businessID, ID: dept.ID}); getErr != nil {
			return nil, updateNotFound(getErr, domain.ErrDepartmentNotFound)
		}
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
