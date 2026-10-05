package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func (r *Repository) AdvanceLeaveApprovalLevel(ctx context.Context, ownerID, requestID int64, currentLevel int32) error {
	result, err := r.queries.AdvanceLeaveApprovalLevel(ctx, sqlc.AdvanceLeaveApprovalLevelParams{ID: requestID, BusinessID: ownerID, CurrentApprovalLevel: currentLevel})
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

func (r *Repository) RejectLeaveRequestWithBalance(ctx context.Context, ownerID, requestID, employeeID, leaveTypeID int64, allocations []domain.LeaveDayAllocation, reason string) (err error) {
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
	result, err := qtx.RejectLeaveRequest(ctx, sqlc.RejectLeaveRequestParams{RejectionReason: sqlutil.NullString(reason), ID: requestID, BusinessID: ownerID})
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
	if err = subtractPendingDaysInTx(ctx, qtx, ownerID, employeeID, leaveTypeID, allocations); err != nil {
		return err
	}
	return tx.Commit()
}

// RejectLeaveRequest reports whether the request was actually still pending
// at the moment of the write (the query's WHERE clause guards on it) - a
// zero result means it was already approved, rejected, or canceled between
// the read and this write, in which case pending days must not be adjusted.
func (r *Repository) RejectLeaveRequest(ctx context.Context, ownerID, requestID int64, reason string) (bool, error) {
	result, err := r.queries.RejectLeaveRequest(ctx, sqlc.RejectLeaveRequestParams{
		RejectionReason: sqlutil.NullString(reason),
		ID:              requestID,
		BusinessID:      ownerID,
	})
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

// CancelLeaveRequest reports whether the request was actually still pending
// at the moment of the write (the query's WHERE clause guards on it) - a
// zero result means it was approved, rejected, or already canceled between
// the caller's read and this write, not that the write failed.
func (r *Repository) CancelLeaveRequest(ctx context.Context, ownerID, requestID int64) (bool, error) {
	result, err := r.queries.CancelLeaveRequest(ctx, sqlc.CancelLeaveRequestParams{
		ID:         requestID,
		BusinessID: ownerID,
	})
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (r *Repository) CancelLeaveRequestWithBalance(ctx context.Context, ownerID, requestID, employeeID, leaveTypeID int64, allocations []domain.LeaveDayAllocation) (err error) {
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
	result, err := qtx.CancelLeaveRequest(ctx, sqlc.CancelLeaveRequestParams{ID: requestID, BusinessID: ownerID})
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
	if err = subtractPendingDaysInTx(ctx, qtx, ownerID, employeeID, leaveTypeID, allocations); err != nil {
		return err
	}
	return tx.Commit()
}

func subtractPendingDaysInTx(ctx context.Context, qtx *sqlc.Queries, ownerID, employeeID, leaveTypeID int64, allocations []domain.LeaveDayAllocation) error {
	for _, allocation := range allocations {
		result, err := qtx.SubtractPendingDays(ctx, sqlc.SubtractPendingDaysParams{
			PendingDays: fmt.Sprintf("%.2f", allocation.Days), BusinessID: ownerID,
			EmployeeID: employeeID, LeaveTypeID: leaveTypeID, Year: allocation.Year,
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
	}
	return nil
}

func (r *Repository) ListOverlappingLeaveRequests(ctx context.Context, ownerID, employeeID int64, startDate, endDate string) ([]domain.LeaveRequestRecord, error) {
	sd, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}
	ed, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListOverlappingLeaveRequests(ctx, sqlc.ListOverlappingLeaveRequestsParams{
		BusinessID: ownerID,
		EmployeeID: employeeID,
		EndDate:    ed,
		StartDate:  sd,
	})
	if err != nil {
		return nil, err
	}

	requests := make([]domain.LeaveRequestRecord, 0, len(rows))
	for _, row := range rows {
		requests = append(requests, leaveRequestFromRow(row))
	}
	return requests, nil
}

func (r *Repository) ListLeaveCalendar(ctx context.Context, businessID, locationID int64, dateFrom, dateTo string) ([]domain.LeaveRequestRecord, error) {
	if locationID < 0 {
		return nil, domain.ErrInvalidCalendarLocation
	}
	df, err := time.Parse("2006-01-02", dateFrom)
	if err != nil {
		return nil, err
	}
	dt, err := time.Parse("2006-01-02", dateTo)
	if err != nil {
		return nil, err
	}
	if dt.Before(df) {
		return nil, domain.ErrInvalidDateRange
	}

	rows, err := r.queries.ListLeaveCalendar(ctx, sqlc.ListLeaveCalendarParams{
		BusinessID: businessID,
		DateTo:     dt,
		DateFrom:   df,
		LocationID: nullInt64Value(locationID),
	})
	if err != nil {
		return nil, err
	}

	requests := make([]domain.LeaveRequestRecord, 0, len(rows))
	for _, row := range rows {
		requests = append(requests, leaveRequestFromRow(row))
	}
	return requests, nil
}

// --- Leave Balances ---
