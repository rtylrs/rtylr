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

// --- Leave Types ---

func (r *Repository) ListLeaveTypes(ctx context.Context, ownerID int64, params domain.ListLeaveTypesParams) ([]domain.LeaveTypeRecord, int64, error) {
	total, err := r.queries.CountLeaveTypes(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to count leave types", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListLeaveTypes(ctx, sqlc.ListLeaveTypesParams{
		BusinessID: ownerID,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list leave types", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, 0, err
	}

	types := make([]domain.LeaveTypeRecord, 0, len(rows))
	for _, row := range rows {
		types = append(types, leaveTypeFromRow(row))
	}
	return types, total, nil
}

func (r *Repository) GetLeaveType(ctx context.Context, ownerID, leaveTypeID int64) (*domain.LeaveTypeRecord, error) {
	row, err := r.queries.GetLeaveType(ctx, sqlc.GetLeaveTypeParams{BusinessID: ownerID, ID: leaveTypeID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrLeaveTypeNotFound
		}
		logger.Ctx(ctx).Error("failed to get leave type", logger.Any("owner_id", ownerID), logger.Any("leave_type_id", leaveTypeID), logger.Error(err))
		return nil, err
	}
	lt := leaveTypeFromRow(row)
	return &lt, nil
}

func (r *Repository) CreateLeaveType(ctx context.Context, ownerID int64, lt domain.LeaveTypeRecord) (*domain.LeaveTypeRecord, error) {
	result, err := r.queries.CreateLeaveType(ctx, sqlc.CreateLeaveTypeParams{
		OwnerID:                        ownerID,
		BusinessID:                     ownerID,
		Name:                           lt.Name,
		IsPaid:                         lt.IsPaid,
		RequiresApproval:               lt.RequiresApproval,
		AccrualRule:                    nullString(lt.AccrualRule),
		AccrualAmount:                  nullStringFromFloat(lt.AccrualAmount),
		MaxCarryoverDays:               lt.MaxCarryoverDays,
		MinNoticeDays:                  lt.MinNoticeDays,
		RequiresDocumentationAfterDays: nullInt32(lt.RequiresDocumentationAfterDays),
		IsActive:                       lt.IsActive,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create leave type", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	lt.ID = id
	lt.OwnerID = ownerID
	lt.BusinessID = ownerID
	return &lt, nil
}

func (r *Repository) UpdateLeaveType(ctx context.Context, ownerID int64, lt domain.LeaveTypeRecord) error {
	return r.queries.UpdateLeaveType(ctx, sqlc.UpdateLeaveTypeParams{
		Name:                           lt.Name,
		IsPaid:                         lt.IsPaid,
		RequiresApproval:               lt.RequiresApproval,
		AccrualRule:                    nullString(lt.AccrualRule),
		AccrualAmount:                  nullStringFromFloat(lt.AccrualAmount),
		MaxCarryoverDays:               lt.MaxCarryoverDays,
		MinNoticeDays:                  lt.MinNoticeDays,
		RequiresDocumentationAfterDays: nullInt32(lt.RequiresDocumentationAfterDays),
		IsActive:                       lt.IsActive,
		ID:                             lt.ID,
		BusinessID:                     ownerID,
	})
}

func (r *Repository) ListAllActiveLeaveTypesForOwner(ctx context.Context, ownerID int64) ([]domain.LeaveTypeRecord, error) {
	rows, err := r.queries.ListAllActiveLeaveTypesForOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	types := make([]domain.LeaveTypeRecord, 0, len(rows))
	for _, row := range rows {
		types = append(types, leaveTypeFromRow(row))
	}
	return types, nil
}

func (r *Repository) ListDistinctOwnerIDs(ctx context.Context) ([]int64, error) {
	return r.queries.ListDistinctOwnerIDs(ctx)
}

// --- Leave Requests ---

func (r *Repository) ListLeaveRequests(ctx context.Context, ownerID int64, params domain.ListLeaveRequestsParams) ([]domain.LeaveRequestRecord, int64, error) {
	fromDate, err := leaveDateArg(params.FromDate)
	if err != nil {
		logger.Ctx(ctx).Warn("invalid leave list from_date", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, 0, err
	}
	toDate, err := leaveDateArg(params.ToDate)
	if err != nil {
		logger.Ctx(ctx).Warn("invalid leave list to_date", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, 0, err
	}

	total, err := r.queries.CountLeaveRequests(ctx, sqlc.CountLeaveRequestsParams{
		BusinessID: ownerID,
		EmployeeID: params.EmployeeID,
		Status:     params.Status,
		FromDate:   fromDate,
		ToDate:     toDate,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to count leave requests", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListLeaveRequests(ctx, sqlc.ListLeaveRequestsParams{
		BusinessID: ownerID,
		EmployeeID: params.EmployeeID,
		Status:     params.Status,
		FromDate:   fromDate,
		ToDate:     toDate,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list leave requests", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, 0, err
	}

	requests := make([]domain.LeaveRequestRecord, 0, len(rows))
	for _, row := range rows {
		requests = append(requests, leaveRequestFromRow(row))
	}
	return requests, total, nil
}

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
		Notes:                nullString(req.Notes),
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
		balance := leaveBalanceFromRow(row)
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
		Notes: nullString(req.Notes), Status: req.Status, CurrentApprovalLevel: req.CurrentApprovalLevel,
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
	result, err := qtx.RejectLeaveRequest(ctx, sqlc.RejectLeaveRequestParams{RejectionReason: nullString(reason), ID: requestID, BusinessID: ownerID})
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
		RejectionReason: nullString(reason),
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

func (r *Repository) ListLeaveCalendar(ctx context.Context, ownerID int64, dateFrom, dateTo string) ([]domain.LeaveRequestRecord, error) {
	df, err := time.Parse("2006-01-02", dateFrom)
	if err != nil {
		return nil, err
	}
	dt, err := time.Parse("2006-01-02", dateTo)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListLeaveCalendar(ctx, sqlc.ListLeaveCalendarParams{
		BusinessID: ownerID,
		EndDate:    dt,
		StartDate:  df,
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

func (r *Repository) ListLeaveBalances(ctx context.Context, ownerID, employeeID int64, year int32) ([]domain.LeaveBalanceRecord, error) {
	rows, err := r.queries.ListLeaveBalances(ctx, sqlc.ListLeaveBalancesParams{
		BusinessID: ownerID,
		EmployeeID: employeeID,
		Year:       year,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list leave balances", logger.Any("owner_id", ownerID), logger.Error(err))
		return nil, err
	}

	balances := make([]domain.LeaveBalanceRecord, 0, len(rows))
	for _, row := range rows {
		balances = append(balances, leaveBalanceFromRow(row))
	}
	return balances, nil
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
	bal := leaveBalanceFromRow(row)
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
		balances = append(balances, leaveBalanceFromRow(row))
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

func leaveBalanceFromRow(row sqlc.HrLeaveBalance) domain.LeaveBalanceRecord {
	return domain.LeaveBalanceRecord{
		ID:               row.ID,
		OwnerID:          row.OwnerID,
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

func nullString(v string) sql.NullString {
	return sql.NullString{String: v, Valid: v != ""}
}

func nullInt32(v int32) sql.NullInt32 {
	return sql.NullInt32{Int32: v, Valid: v != 0}
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
