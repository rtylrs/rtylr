package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/voxire/onion/internal/modules/hr/staff/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/staff/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

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

func (r *Repository) EnsureStaffForUser(ctx context.Context, ownerID, businessID, userID int64, role string) error {
	userIDParam := sql.NullInt64{Int64: userID, Valid: true}
	if err := r.queries.LinkStaffEmployeeForUser(ctx, sqlc.LinkStaffEmployeeForUserParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		UserID:     userIDParam,
	}); err != nil {
		return err
	}
	return r.queries.EnsureStaffForUser(ctx, sqlc.EnsureStaffForUserParams{
		Role:       role,
		OwnerID:    ownerID,
		BusinessID: businessID,
		UserID:     userIDParam,
	})
}

func (r *Repository) DeleteStaffForUser(ctx context.Context, ownerID, businessID, userID int64) error {
	return r.queries.DeleteStaffForUser(ctx, sqlc.DeleteStaffForUserParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		UserID:     sql.NullInt64{Int64: userID, Valid: true},
	})
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
