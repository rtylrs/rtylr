package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// DeleteWorkType removes a node. Reports whether a row was removed.
func (r *Repository) DeleteWorkType(ctx context.Context, id, businessID int64) (bool, error) {
	res, err := r.queries.DeleteWorkType(ctx, sqlc.DeleteWorkTypeParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteWorkType failed", logger.Error(err))
		return false, err
	}
	affected, err := res.RowsAffected()
	return affected > 0, err
}

// CountWorkTypeChildren reports how many nodes hang off a parent, which is what
// keeps the tree to two levels and stops a parent with children being nested.
func (r *Repository) CountWorkTypeChildren(ctx context.Context, businessID, parentID int64) (int64, error) {
	return r.queries.CountWorkTypeChildren(ctx, sqlc.CountWorkTypeChildrenParams{
		BusinessID: businessID,
		ParentID:   nullInt64(parentID),
	})
}

// ListWorkDays returns a person's days across a date range, newest first, with
// the total for pagination. Paginated because a year of days is 250 rows and no
// list endpoint here reads unbounded.
func (r *Repository) ListWorkDays(ctx context.Context, businessID, employeeID int64, from, to string, limit int32, offset int64) ([]domain.WorkDay, int64, error) {
	fromDate, err := parseDate(from)
	if err != nil {
		return nil, 0, err
	}
	toDate, err := parseDate(to)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListWorkDays(ctx, sqlc.ListWorkDaysParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		Date:       fromDate,
		Date_2:     toDate,
		Limit:      limit,
		Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListWorkDays failed", logger.Error(err))
		return nil, 0, err
	}

	days := make([]domain.WorkDay, 0, len(rows))
	for _, row := range rows {
		days = append(days, toDomainWorkDay(workDayFromList(row)))
	}

	total, err := r.queries.CountWorkDays(ctx, sqlc.CountWorkDaysParams{
		BusinessID: businessID,
		EmployeeID: employeeID,
		Date:       fromDate,
		Date_2:     toDate,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CountWorkDays failed", logger.Error(err))
		return days, 0, err
	}
	return days, total, nil
}

// ListWorkDaysByEmployeeIDs is ListWorkDays across a resolved set of
// employees rather than one, for the workspace team view. Ordered by
// employee then date so the caller can group rows into a roster without a
// second pass.
func (r *Repository) ListWorkDaysByEmployeeIDs(ctx context.Context, businessID int64, employeeIDs []int64, from, to string, limit int32, offset int64) ([]domain.WorkDay, int64, error) {
	fromDate, err := parseDate(from)
	if err != nil {
		return nil, 0, err
	}
	toDate, err := parseDate(to)
	if err != nil {
		return nil, 0, err
	}
	if len(employeeIDs) == 0 {
		return []domain.WorkDay{}, 0, nil
	}

	rows, err := r.queries.ListWorkDaysByEmployeeIDs(ctx, sqlc.ListWorkDaysByEmployeeIDsParams{
		BusinessID:  businessID,
		EmployeeIds: employeeIDs,
		Date:        fromDate,
		Date_2:      toDate,
		Limit:       limit,
		Offset:      int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListWorkDaysByEmployeeIDs failed", logger.Error(err))
		return nil, 0, err
	}

	days := make([]domain.WorkDay, 0, len(rows))
	for _, row := range rows {
		days = append(days, toDomainWorkDay(workDayFromEmployeeIDs(row)))
	}

	total, err := r.queries.CountWorkDaysByEmployeeIDs(ctx, sqlc.CountWorkDaysByEmployeeIDsParams{
		BusinessID:  businessID,
		EmployeeIds: employeeIDs,
		Date:        fromDate,
		Date_2:      toDate,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CountWorkDaysByEmployeeIDs failed", logger.Error(err))
		return days, 0, err
	}
	return days, total, nil
}

// ListTeamWorkDays reads Flow-owned days for IDs resolved by the roster port.
func (r *Repository) ListTeamWorkDays(ctx context.Context, businessID int64, employeeIDs, hiddenIDs, userIDs []int64, from, to string, limit int32, offset int64) ([]domain.WorkDay, int64, error) {
	fromDate, err := parseDate(from)
	if err != nil {
		return nil, 0, err
	}
	toDate, err := parseDate(to)
	if err != nil {
		return nil, 0, err
	}
	if len(employeeIDs) == 0 && len(userIDs) == 0 {
		return []domain.WorkDay{}, 0, nil
	}
	if len(employeeIDs) == 0 {
		employeeIDs = []int64{-1}
	}
	if len(userIDs) == 0 {
		userIDs = []int64{-1}
	}
	if len(hiddenIDs) == 0 {
		hiddenIDs = []int64{-1}
	}
	rows, err := r.queries.ListTeamWorkDaysByRoster(ctx, sqlc.ListTeamWorkDaysByRosterParams{
		BusinessID: businessID, EmployeeIds: employeeIDs, HiddenIds: hiddenIDs, UserIds: userIDs,
		Date: fromDate, Date_2: toDate, Limit: limit, Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	days := make([]domain.WorkDay, 0, len(rows))
	for _, row := range rows {
		days = append(days, toDomainWorkDay(sqlc.FlowWorkDay{
			ID: row.ID, EmployeeID: row.EmployeeID, Date: row.Date, Timezone: row.Timezone,
			WorkMode: row.WorkMode, StartTime: row.StartTime, EndTime: row.EndTime, BreakMinutes: row.BreakMinutes,
			Status: row.Status, LoggedBy: row.LoggedBy, ApprovedBy: row.ApprovedBy, ApprovedAt: row.ApprovedAt,
			ReopenedBy: row.ReopenedBy, ReopenedAt: row.ReopenedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}))
	}
	total, err := r.queries.CountTeamWorkDaysByRoster(ctx, sqlc.CountTeamWorkDaysByRosterParams{
		BusinessID: businessID, EmployeeIds: employeeIDs, HiddenIds: hiddenIDs, UserIds: userIDs,
		Date: fromDate, Date_2: toDate,
	})
	return days, total, err
}
