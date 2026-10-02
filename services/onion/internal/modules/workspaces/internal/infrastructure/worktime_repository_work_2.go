package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// DeleteWorkType removes a node. Reports whether a row was removed.
func (r *Repository) DeleteWorkType(ctx context.Context, id, ownerID, businessID int64) (bool, error) {
	res, err := r.queries.DeleteWorkType(ctx, sqlc.DeleteWorkTypeParams{
		ID:         id,
		OwnerID:    ownerID,
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
func (r *Repository) CountWorkTypeChildren(ctx context.Context, ownerID, businessID, parentID int64) (int64, error) {
	return r.queries.CountWorkTypeChildren(ctx, sqlc.CountWorkTypeChildrenParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		ParentID:   nullInt64(parentID),
	})
}

// ListWorkDays returns a person's days across a date range, newest first, with
// the total for pagination. Paginated because a year of days is 250 rows and no
// list endpoint here reads unbounded.
func (r *Repository) ListWorkDays(ctx context.Context, ownerID, businessID, employeeID int64, from, to string, limit int32, offset int64) ([]domain.WorkDay, int64, error) {
	fromDate, err := parseDate(from)
	if err != nil {
		return nil, 0, err
	}
	toDate, err := parseDate(to)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListWorkDays(ctx, sqlc.ListWorkDaysParams{
		OwnerID:    ownerID,
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
		OwnerID:    ownerID,
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

// EmployeeIDsByUserIDs resolves workspace members (auth user ids) to the HR
// employee ids work days are keyed on. hrOwnerID is the HR tenant key, which
// callers pass as businessID because hr_core stores that value in both
// owner_id and business_id. A user with no hr_employees row (or none matching
// this tenant) is simply absent from the result - there is no day to show for
// someone who was never onboarded into HR, which is a normal state, not an
// error.
func (r *Repository) EmployeeIDsByUserIDs(ctx context.Context, hrOwnerID, businessID int64, userIDs []int64) (map[int64]int64, error) {
	if len(userIDs) == 0 {
		return map[int64]int64{}, nil
	}
	args := make([]sql.NullInt64, 0, len(userIDs))
	for _, id := range userIDs {
		args = append(args, sql.NullInt64{Int64: id, Valid: true})
	}

	rows, err := r.queries.GetEmployeeIDsByUserIDs(ctx, sqlc.GetEmployeeIDsByUserIDsParams{
		OwnerID:    hrOwnerID,
		BusinessID: businessID,
		UserIds:    args,
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetEmployeeIDsByUserIDs failed", logger.Error(err))
		return nil, err
	}

	// Keyed by user_id -> employee_id, the direction the caller resolves in:
	// it already has workspace members (user ids) and wants their work days.
	byUserID := make(map[int64]int64, len(rows))
	for _, row := range rows {
		if row.UserID.Valid {
			byUserID[row.UserID.Int64] = row.ID
		}
	}
	return byUserID, nil
}

// ListWorkDaysByEmployeeIDs is ListWorkDays across a resolved set of
// employees rather than one, for the workspace team view. Ordered by
// employee then date so the caller can group rows into a roster without a
// second pass.
func (r *Repository) ListWorkDaysByEmployeeIDs(ctx context.Context, ownerID, businessID int64, employeeIDs []int64, from, to string, limit int32, offset int64) ([]domain.WorkDay, int64, error) {
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
		OwnerID:     ownerID,
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
		OwnerID:     ownerID,
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

// ListTeamWorkDays resolves user-keyed and HR-keyed days within one business.
// A nil roster is reserved for the application's business-wide visibility gate.
func (r *Repository) ListTeamWorkDays(ctx context.Context, businessID int64, userIDs []int64, from, to string, limit int32, offset int64) ([]domain.WorkDay, int64, error) {
	fromDate, err := parseDate(from)
	if err != nil {
		return nil, 0, err
	}
	toDate, err := parseDate(to)
	if err != nil {
		return nil, 0, err
	}
	includeAll := userIDs == nil
	if len(userIDs) == 0 {
		if !includeAll {
			return []domain.WorkDay{}, 0, nil
		}
		userIDs = []int64{-1}
	}
	hrUserIDs := make([]sql.NullInt64, 0, len(userIDs))
	for _, id := range userIDs {
		hrUserIDs = append(hrUserIDs, sql.NullInt64{Int64: id, Valid: true})
	}
	rows, err := r.queries.ListTeamWorkDaysV2(ctx, sqlc.ListTeamWorkDaysV2Params{
		BusinessID: businessID, IncludeAll: includeAll, UserIds: userIDs, HrUserIds: hrUserIDs,
		FromDate: fromDate, ToDate: toDate, Limit: limit, Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}
	days := make([]domain.WorkDay, 0, len(rows))
	for _, row := range rows {
		day := toDomainWorkDay(sqlc.FlowWorkDay{
			ID: row.ID, EmployeeID: row.EmployeeID, Date: row.Date, Timezone: row.Timezone,
			WorkMode: row.WorkMode, StartTime: row.StartTime, EndTime: row.EndTime, BreakMinutes: row.BreakMinutes,
			Status: row.Status, LoggedBy: row.LoggedBy, ApprovedBy: row.ApprovedBy, ApprovedAt: row.ApprovedAt,
			ReopenedBy: row.ReopenedBy, ReopenedAt: row.ReopenedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		})
		day.UserID = sumToInt64(row.UserID)
		days = append(days, day)
	}
	total, err := r.queries.CountTeamWorkDaysV2(ctx, sqlc.CountTeamWorkDaysV2Params{
		BusinessID: businessID, IncludeAll: includeAll, UserIds: userIDs, HrUserIds: hrUserIDs,
		FromDate: fromDate, ToDate: toDate,
	})
	return days, total, err
}
