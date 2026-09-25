package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
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
		AccrualRule:                    sqlutil.NullString(lt.AccrualRule),
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
	result, err := r.queries.UpdateLeaveType(ctx, sqlc.UpdateLeaveTypeParams{
		Name:                           lt.Name,
		IsPaid:                         lt.IsPaid,
		RequiresApproval:               lt.RequiresApproval,
		AccrualRule:                    sqlutil.NullString(lt.AccrualRule),
		AccrualAmount:                  nullStringFromFloat(lt.AccrualAmount),
		MaxCarryoverDays:               lt.MaxCarryoverDays,
		MinNoticeDays:                  lt.MinNoticeDays,
		RequiresDocumentationAfterDays: nullInt32(lt.RequiresDocumentationAfterDays),
		IsActive:                       lt.IsActive,
		ID:                             lt.ID,
		BusinessID:                     ownerID,
	})
	if err != nil {
		return err
	}
	return r.requireLeaveTypeExistsAfterUpdate(ctx, ownerID, lt.ID, result)
}

// UpdateLeaveTypePatch changes only fields explicitly supplied by the caller.
// The SQL uses nullable COALESCE parameters so false and zero remain distinct
// from an omitted field.
func (r *Repository) UpdateLeaveTypePatch(ctx context.Context, ownerID, leaveTypeID int64, patch domain.LeaveTypePatch) error {
	result, err := r.queries.PatchLeaveType(ctx, sqlc.PatchLeaveTypeParams{
		Name:                           nullStringPtr(patch.Name),
		IsPaid:                         nullBoolPtr(patch.IsPaid),
		RequiresApproval:               nullBoolPtr(patch.RequiresApproval),
		AccrualRule:                    nullStringPtr(patch.AccrualRule),
		AccrualAmount:                  nullFloatPtr(patch.AccrualAmount),
		MaxCarryoverDays:               nullInt32Ptr(patch.MaxCarryoverDays),
		MinNoticeDays:                  nullInt32Ptr(patch.MinNoticeDays),
		RequiresDocumentationAfterDays: nullInt32Ptr(patch.RequiresDocumentationAfterDays),
		IsActive:                       nullBoolPtr(patch.IsActive),
		ID:                             leaveTypeID,
		BusinessID:                     ownerID,
	})
	if err != nil {
		return err
	}
	return r.requireLeaveTypeExistsAfterUpdate(ctx, ownerID, leaveTypeID, result)
}

func (r *Repository) requireLeaveTypeExistsAfterUpdate(ctx context.Context, ownerID, leaveTypeID int64, result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows > 0 {
		return nil
	}
	// MySQL may report zero affected rows when every supplied value already
	// matches the stored row. Probe the scoped row before calling that a 404.
	_, err = r.queries.GetLeaveType(ctx, sqlc.GetLeaveTypeParams{BusinessID: ownerID, ID: leaveTypeID})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrLeaveTypeNotFound
	}
	return err
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
