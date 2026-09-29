package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) ListLeaveBalances(ctx context.Context, ownerID, employeeID int64, year int32) ([]domain.LeaveBalanceRecord, error) {
	rows, err := r.queries.ListLeaveBalances(ctx, sqlc.ListLeaveBalancesParams{
		BusinessID: ownerID,
		EmployeeID: nullInt64Value(employeeID),
		Year:       year,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list leave balances", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}

	balances := make([]domain.LeaveBalanceRecord, 0, len(rows))
	for _, row := range rows {
		balances = append(balances, leaveBalanceFromRow(sqlc.GetLeaveBalanceRow(row)))
	}
	return balances, nil
}

func (r *Repository) CountLeaveBalances(ctx context.Context, ownerID, employeeID int64, year int32) (int64, error) {
	return r.queries.CountLeaveBalances(ctx, sqlc.CountLeaveBalancesParams{
		BusinessID: ownerID,
		EmployeeID: nullInt64Value(employeeID),
		Year:       year,
	})
}

func (r *Repository) GetLeaveBalance(ctx context.Context, ownerID, employeeID, leaveTypeID int64, year int32) (*domain.LeaveBalanceRecord, error) {
	row, err := r.queries.GetLeaveBalance(ctx, sqlc.GetLeaveBalanceParams{
		BusinessID:  ownerID,
		EmployeeID:  employeeID,
		LeaveTypeID: leaveTypeID,
		Year:        year,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.LeaveBalanceRecord{
				OwnerID:     ownerID,
				BusinessID:  ownerID,
				EmployeeID:  employeeID,
				LeaveTypeID: leaveTypeID,
				Year:        year,
			}, nil
		}
		return nil, err
	}
	bal := leaveBalanceFromRow(sqlc.GetLeaveBalanceRow(row))
	return &bal, nil
}

func (r *Repository) UpsertLeaveBalance(ctx context.Context, ownerID int64, bal domain.LeaveBalanceRecord) error {
	_, err := r.queries.UpsertLeaveBalance(ctx, sqlc.UpsertLeaveBalanceParams{
		OwnerID:          ownerID,
		BusinessID:       ownerID,
		EmployeeID:       bal.EmployeeID,
		LeaveTypeID:      bal.LeaveTypeID,
		Year:             bal.Year,
		EntitledDays:     fmt.Sprintf("%.2f", bal.EntitledDays),
		UsedDays:         fmt.Sprintf("%.2f", bal.UsedDays),
		PendingDays:      fmt.Sprintf("%.2f", bal.PendingDays),
		LastAccrualMonth: bal.LastAccrualMonth,
	})
	return err
}

func (r *Repository) AddToUsedDays(ctx context.Context, ownerID, employeeID, leaveTypeID int64, year int32, days float64) error {
	result, err := r.queries.AddToUsedDays(ctx, sqlc.AddToUsedDaysParams{
		UsedDays:    fmt.Sprintf("%.2f", days),
		PendingDays: fmt.Sprintf("%.2f", days),
		BusinessID:  ownerID,
		EmployeeID:  employeeID,
		LeaveTypeID: leaveTypeID,
		Year:        year,
	})
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrLeaveBalanceNotFound
	}
	return nil
}

func (r *Repository) AddToPendingDays(ctx context.Context, ownerID, employeeID, leaveTypeID int64, year int32, days float64) error {
	result, err := r.queries.AddToPendingDays(ctx, sqlc.AddToPendingDaysParams{
		PendingDays: fmt.Sprintf("%.2f", days),
		BusinessID:  ownerID,
		EmployeeID:  employeeID,
		LeaveTypeID: leaveTypeID,
		Year:        year,
	})
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrLeaveBalanceNotFound
	}
	return nil
}

func (r *Repository) SubtractPendingDays(ctx context.Context, ownerID, employeeID, leaveTypeID int64, year int32, days float64) error {
	result, err := r.queries.SubtractPendingDays(ctx, sqlc.SubtractPendingDaysParams{
		PendingDays: fmt.Sprintf("%.2f", days),
		BusinessID:  ownerID,
		EmployeeID:  employeeID,
		LeaveTypeID: leaveTypeID,
		Year:        year,
	})
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrLeaveBalanceNotFound
	}
	return nil
}

func (r *Repository) ListBalancesForAccrual(ctx context.Context, ownerID int64, year, currentMonth int32) ([]domain.LeaveBalanceRecord, error) {
	rows, err := r.queries.ListBalancesForAccrual(ctx, sqlc.ListBalancesForAccrualParams{
		BusinessID:       ownerID,
		Year:             year,
		LastAccrualMonth: currentMonth,
	})
	if err != nil {
		return nil, err
	}
	balances := make([]domain.LeaveBalanceRecord, 0, len(rows))
	for _, row := range rows {
		balances = append(balances, leaveBalanceFromRow(sqlc.GetLeaveBalanceRow(row)))
	}
	return balances, nil
}

func (r *Repository) AddToEntitledDays(ctx context.Context, ownerID, employeeID, leaveTypeID int64, year int32, days float64, month int32) error {
	result, err := r.queries.AddToEntitledDays(ctx, sqlc.AddToEntitledDaysParams{
		EntitledDays:       fmt.Sprintf("%.2f", days),
		LastAccrualMonth:   month,
		BusinessID:         ownerID,
		EmployeeID:         employeeID,
		LeaveTypeID:        leaveTypeID,
		Year:               year,
		LastAccrualMonth_2: month,
	})
	if err != nil {
		return err
	}
	// A zero-row update is an intentional no-op when another worker already
	// applied this month's accrual. The conditional update is the concurrency
	// guard; a missing row is not returned by ListBalancesForAccrual.
	_, err = result.RowsAffected()
	return err
}

func (r *Repository) ApplyYearEndCarryover(ctx context.Context, ownerID int64, previousYear, currentYear int32) error {
	_, err := r.queries.ApplyYearEndCarryover(ctx, sqlc.ApplyYearEndCarryoverParams{
		Year: currentYear, BusinessID: ownerID, Year_2: previousYear,
	})
	return err
}

// --- Helpers ---

func leaveTypeFromRow(row sqlc.HrLeaveType) domain.LeaveTypeRecord {
	return domain.LeaveTypeRecord{
		ID:                             row.ID,
		OwnerID:                        row.OwnerID,
		BusinessID:                     row.BusinessID,
		Name:                           row.Name,
		IsPaid:                         row.IsPaid,
		RequiresApproval:               row.RequiresApproval,
		AccrualRule:                    row.AccrualRule.String,
		AccrualAmount:                  parseFloat(row.AccrualAmount.String),
		MaxCarryoverDays:               row.MaxCarryoverDays,
		MinNoticeDays:                  row.MinNoticeDays,
		RequiresDocumentationAfterDays: row.RequiresDocumentationAfterDays.Int32,
		IsActive:                       row.IsActive,
		CreatedAt:                      row.CreatedAt,
		UpdatedAt:                      row.UpdatedAt,
	}
}

func leaveRequestFromRow(row sqlc.HrLeaveRequest) domain.LeaveRequestRecord {
	return domain.LeaveRequestRecord{
		ID:                   row.ID,
		OwnerID:              row.OwnerID,
		BusinessID:           row.BusinessID,
		EmployeeID:           row.EmployeeID,
		LeaveTypeID:          row.LeaveTypeID,
		StartDate:            row.StartDate.Format("2006-01-02"),
		EndDate:              row.EndDate.Format("2006-01-02"),
		TotalDays:            parseFloat(row.TotalDays),
		Notes:                row.Notes.String,
		Status:               row.Status,
		CurrentApprovalLevel: row.CurrentApprovalLevel,
		RejectionReason:      row.RejectionReason.String,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}

func leaveBalanceFromRow(row sqlc.GetLeaveBalanceRow) domain.LeaveBalanceRecord {
	return domain.LeaveBalanceRecord{
		ID:               row.ID,
		OwnerID:          row.OwnerID,
		BusinessID:       row.BusinessID,
		EmployeeID:       row.EmployeeID,
		LeaveTypeID:      row.LeaveTypeID,
		Year:             row.Year,
		EntitledDays:     parseFloat(row.EntitledDays),
		UsedDays:         parseFloat(row.UsedDays),
		PendingDays:      parseFloat(row.PendingDays),
		LastAccrualMonth: row.LastAccrualMonth,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse leave decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}

func nullInt32(v int32) sql.NullInt32 {
	return sql.NullInt32{Int32: v, Valid: v != 0}
}

func nullStringPtr(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}

func nullBoolPtr(v *bool) sql.NullBool {
	if v == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{Bool: *v, Valid: true}
}

func nullFloatPtr(v *float64) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: fmt.Sprintf("%.2f", *v), Valid: true}
}

func nullInt32Ptr(v *int32) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *v, Valid: true}
}

func nullInt64Value(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

func nullStringFromFloat(f float64) sql.NullString {
	if f == 0 {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: fmt.Sprintf("%.2f", f), Valid: true}
}

// leaveDateArg converts an optional ISO date string to the nullable time
// shape the range filters expect; nil stays open-ended.
func leaveDateArg(v *string) (sql.NullTime, error) {
	if v == nil || *v == "" {
		return sql.NullTime{}, nil
	}
	t, err := time.Parse("2006-01-02", *v)
	if err != nil {
		return sql.NullTime{}, fmt.Errorf("invalid date %q: %w", *v, err)
	}
	return sql.NullTime{Time: t, Valid: true}, nil
}
