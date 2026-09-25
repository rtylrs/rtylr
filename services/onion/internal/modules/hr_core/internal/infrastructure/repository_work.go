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

func (r *Repository) UpdateWorkSchedule(ctx context.Context, businessID int64, ws domain.WorkScheduleRecord) (*domain.WorkScheduleRecord, error) {
	ws.ScheduleData = defaultedScheduleData(ws.ScheduleData)

	result, err := r.queries.UpdateWorkSchedule(ctx, sqlc.UpdateWorkScheduleParams{
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
	if affected, err := result.RowsAffected(); err != nil {
		return nil, err
	} else if affected == 0 {
		if _, getErr := r.queries.GetWorkSchedule(ctx, sqlc.GetWorkScheduleParams{BusinessID: businessID, ID: ws.ID}); getErr != nil {
			return nil, updateNotFound(getErr, domain.ErrWorkScheduleNotFound)
		}
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
		Description: sqlutil.NullString(role.Description),
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
	result, err := r.queries.UpdateRole(ctx, sqlc.UpdateRoleParams{
		Name:        role.Name,
		Description: sqlutil.NullString(role.Description),
		Permissions: json.RawMessage(role.Permissions),
		ID:          role.ID,
		BusinessID:  businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to update role", logger.Any("business_id", businessID), logger.Any("role_id", role.ID), logger.Error(err))
		return nil, err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return nil, err
	} else if affected == 0 {
		if _, getErr := r.queries.GetRole(ctx, sqlc.GetRoleParams{BusinessID: businessID, ID: role.ID}); getErr != nil {
			return nil, updateNotFound(getErr, domain.ErrRoleNotFound)
		}
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

func (r *Repository) AssignDefaultOwnerRole(ctx context.Context, businessID, userID int64) error {
	if err := r.queries.AssignDefaultOwnerRole(ctx, sqlc.AssignDefaultOwnerRoleParams{
		BusinessID: businessID,
		UserID:     sqlutil.NullInt64(userID),
	}); err != nil {
		logger.Ctx(ctx).Error("failed to assign default HR owner role",
			logger.Any("business_id", businessID), logger.Any("user_id", userID), logger.Error(err))
		return err
	}
	return nil
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSetupStepNotFound
		}
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
