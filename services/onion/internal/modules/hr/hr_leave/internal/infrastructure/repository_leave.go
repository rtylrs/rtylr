package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func (r *Repository) GetLeaveRequest(ctx context.Context, ownerID, requestID int64) (*domain.LeaveRequestRecord, error) {
	row, err := r.queries.GetLeaveRequest(ctx, sqlc.GetLeaveRequestParams{BusinessID: ownerID, ID: requestID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrLeaveRequestNotFound
		}
		logger.Ctx(ctx).Error("failed to get leave request", logger.Any("owner_id", ownerID), logger.Any("request_id", requestID), logger.Error(err))
		return nil, err
	}
	req := leaveRequestFromRow(row)
	return &req, nil
}

func (r *Repository) CreateLeaveRequest(ctx context.Context, ownerID int64, req domain.LeaveRequestRecord) (*domain.LeaveRequestRecord, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, err
	}

	result, err := r.queries.CreateLeaveRequest(ctx, sqlc.CreateLeaveRequestParams{
		OwnerID:              ownerID,
		BusinessID:           ownerID,
		EmployeeID:           req.EmployeeID,
		LeaveTypeID:          req.LeaveTypeID,
		StartDate:            startDate,
		EndDate:              endDate,
		TotalDays:            fmt.Sprintf("%.2f", req.TotalDays),
		Notes:                sqlutil.NullString(req.Notes),
		Status:               req.Status,
		CurrentApprovalLevel: req.CurrentApprovalLevel,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create leave request", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	req.ID = id
	req.OwnerID = ownerID
	req.BusinessID = ownerID
	return &req, nil
}

// CreateLeaveRequestWithBalance makes overlap validation, balance validation,
// request creation, and pending-day accounting one transaction. The balance
// rows are locked first, giving concurrent requests for the same employee and
// leave type a stable serialization point before the overlap query runs.
func (r *Repository) CreateLeaveRequestWithBalance(ctx context.Context, ownerID int64, req domain.LeaveRequestRecord, allocations []domain.LeaveDayAllocation) (created *domain.LeaveRequestRecord, err error) {
	startDate, parseErr := time.Parse("2006-01-02", req.StartDate)
	if parseErr != nil {
		return nil, parseErr
	}
	endDate, parseErr := time.Parse("2006-01-02", req.EndDate)
	if parseErr != nil {
		return nil, parseErr
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)
	for _, allocation := range allocations {
		row, balanceErr := qtx.GetLeaveBalanceForUpdate(ctx, sqlc.GetLeaveBalanceForUpdateParams{
			BusinessID: ownerID, EmployeeID: req.EmployeeID, LeaveTypeID: req.LeaveTypeID, Year: allocation.Year,
		})
		if errors.Is(balanceErr, sql.ErrNoRows) {
			return nil, domain.ErrLeaveBalanceNotFound
		}
		if balanceErr != nil {
			return nil, balanceErr
		}
		balance := leaveBalanceFromRow(sqlc.GetLeaveBalanceRow(row))
		if balance.EntitledDays-balance.UsedDays-balance.PendingDays < allocation.Days {
			return nil, domain.ErrInsufficientLeaveBalance
		}
	}

	overlapping, err := qtx.ListOverlappingLeaveRequests(ctx, sqlc.ListOverlappingLeaveRequestsParams{
		BusinessID: ownerID,
		EmployeeID: req.EmployeeID,
		EndDate:    endDate,
		StartDate:  startDate,
	})
	if err != nil {
		return nil, err
	}
	if len(overlapping) > 0 {
		return nil, domain.ErrLeaveConflict
	}

	result, err := qtx.CreateLeaveRequest(ctx, sqlc.CreateLeaveRequestParams{
		OwnerID: ownerID, BusinessID: ownerID, EmployeeID: req.EmployeeID, LeaveTypeID: req.LeaveTypeID,
		StartDate: startDate, EndDate: endDate, TotalDays: fmt.Sprintf("%.2f", req.TotalDays),
		Notes: sqlutil.NullString(req.Notes), Status: req.Status, CurrentApprovalLevel: req.CurrentApprovalLevel,
	})
	if err != nil {
		return nil, err
	}
	requestID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	for _, allocation := range allocations {
		result, err = qtx.AddToPendingDays(ctx, sqlc.AddToPendingDaysParams{
			PendingDays: fmt.Sprintf("%.2f", allocation.Days), BusinessID: ownerID,
			EmployeeID: req.EmployeeID, LeaveTypeID: req.LeaveTypeID, Year: allocation.Year,
		})
		if err != nil {
			return nil, err
		}
		affected, affectedErr := result.RowsAffected()
		if affectedErr != nil {
			return nil, affectedErr
		}
		if affected == 0 {
			return nil, domain.ErrLeaveBalanceNotFound
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	req.ID = requestID
	req.OwnerID = ownerID
	req.BusinessID = ownerID
	return &req, nil
}

func (r *Repository) ApproveLeaveRequest(ctx context.Context, ownerID, requestID int64) error {
	result, err := r.queries.ApproveLeaveRequest(ctx, sqlc.ApproveLeaveRequestParams{ID: requestID, BusinessID: ownerID})
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrLeaveRequestNotPending
	}
	return nil
}

func (r *Repository) ApproveLeaveRequestWithBalance(ctx context.Context, ownerID, requestID, employeeID, leaveTypeID int64, year int32, days float64) (err error) {
	return r.ApproveLeaveRequestWithAllocations(ctx, ownerID, requestID, employeeID, leaveTypeID, []domain.LeaveDayAllocation{{Year: year, Days: days}})
}

func (r *Repository) ApproveLeaveRequestWithAllocations(ctx context.Context, ownerID, requestID, employeeID, leaveTypeID int64, allocations []domain.LeaveDayAllocation) (err error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)
	result, err := qtx.ApproveLeaveRequest(ctx, sqlc.ApproveLeaveRequestParams{ID: requestID, BusinessID: ownerID})
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrLeaveRequestNotPending
	}

	for _, allocation := range allocations {
		result, err = qtx.AddToUsedDays(ctx, sqlc.AddToUsedDaysParams{
			UsedDays: fmt.Sprintf("%.2f", allocation.Days), PendingDays: fmt.Sprintf("%.2f", allocation.Days),
			BusinessID: ownerID, EmployeeID: employeeID, LeaveTypeID: leaveTypeID, Year: allocation.Year,
		})
		if err != nil {
			return err
		}
		affected, affectedErr := result.RowsAffected()
		if affectedErr != nil {
			return affectedErr
		}
		if affected == 0 {
			return domain.ErrLeaveBalanceNotFound
		}
	}

	return tx.Commit()
}
