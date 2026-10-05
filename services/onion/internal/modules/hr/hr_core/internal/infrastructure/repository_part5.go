package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func (r *Repository) UpsertHRConfig(ctx context.Context, businessID int64, cfg domain.HRCompanyConfigRecord) (*domain.HRCompanyConfigRecord, error) {
	_, err := r.queries.UpsertHRConfig(ctx, sqlc.UpsertHRConfigParams{
		OwnerID:               businessID,
		BusinessID:            businessID,
		FiscalYearStartMonth:  cfg.FiscalYearStartMonth,
		DefaultCurrency:       cfg.DefaultCurrency,
		DefaultTimezone:       cfg.DefaultTimezone,
		DefaultWorkScheduleID: sqlutil.NullInt64(cfg.DefaultWorkScheduleID),
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
			Description: sqlutil.NullString(role.Description),
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
			ApproverRoleID:     sqlutil.NullInt64(level.ApproverRoleID),
			ApproverEmployeeID: sqlutil.NullInt64(level.ApproverEmployeeID),
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
