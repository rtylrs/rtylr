package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

const dateLayout = "2006-01-02"

// parseDate reads a YYYY-MM-DD contract value into a time.Time.
func parseDate(value string) (time.Time, error) {
	return time.Parse(dateLayout, value)
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v > 0}
}

func int64Value(v sql.NullInt64) int64 {
	if v.Valid {
		return v.Int64
	}
	return 0
}

func stringValue(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

// sumToInt64 narrows the interface{} that sqlc returns for COALESCE(SUM(...)).
// MySQL hands a SUM back as a decimal, which the driver may deliver as []byte,
// int64 or float64 depending on the column and the connection; all three are
// handled rather than assuming one and panicking on the others.
func sumToInt64(value interface{}) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	case []byte:
		parsed, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func timeValue(v sql.NullTime) time.Time {
	if v.Valid {
		return v.Time
	}
	return time.Time{}
}

// The narrowed worktime queries each generate a distinct Row struct even
// though they select identical column lists; these explicit adapters let the
// shared mappers stay single-sourced.

func workDayFromGet(r sqlc.GetWorkDayRow) sqlc.FlowWorkDay {
	return sqlc.FlowWorkDay{
		ID: r.ID, EmployeeID: r.EmployeeID, Date: r.Date, Timezone: r.Timezone,
		WorkMode: r.WorkMode, StartTime: r.StartTime, EndTime: r.EndTime, BreakMinutes: r.BreakMinutes,
		Status: r.Status, LoggedBy: r.LoggedBy, ApprovedBy: r.ApprovedBy, ApprovedAt: r.ApprovedAt,
		ReopenedBy: r.ReopenedBy, ReopenedAt: r.ReopenedAt,
	}
}

func workDayFromGetByID(r sqlc.GetWorkDayByIDRow) sqlc.FlowWorkDay {
	return sqlc.FlowWorkDay{
		ID: r.ID, EmployeeID: r.EmployeeID, Date: r.Date, Timezone: r.Timezone,
		WorkMode: r.WorkMode, StartTime: r.StartTime, EndTime: r.EndTime, BreakMinutes: r.BreakMinutes,
		Status: r.Status, LoggedBy: r.LoggedBy, ApprovedBy: r.ApprovedBy, ApprovedAt: r.ApprovedAt,
		ReopenedBy: r.ReopenedBy, ReopenedAt: r.ReopenedAt,
	}
}

func lineFromList(r sqlc.ListWorkDayLinesRow) sqlc.HrTimeEntry {
	return sqlc.HrTimeEntry{
		ID: r.ID, WorkDayID: r.WorkDayID, ProjectID: r.ProjectID, TaskID: r.TaskID,
		WorkTypeID: r.WorkTypeID, WorkMode: r.WorkMode, Description: r.Description,
		DurationMinutes: r.DurationMinutes, IsBillable: r.IsBillable, Status: r.Status,
		ClientKey: r.ClientKey, LoggedBy: r.LoggedBy, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func workDayFromEmployeeIDs(r sqlc.ListWorkDaysByEmployeeIDsRow) sqlc.FlowWorkDay {
	return sqlc.FlowWorkDay{
		ID: r.ID, EmployeeID: r.EmployeeID, Date: r.Date, Timezone: r.Timezone,
		WorkMode: r.WorkMode, StartTime: r.StartTime, EndTime: r.EndTime, BreakMinutes: r.BreakMinutes,
		Status: r.Status, LoggedBy: r.LoggedBy, ApprovedBy: r.ApprovedBy, ApprovedAt: r.ApprovedAt,
		ReopenedBy: r.ReopenedBy, ReopenedAt: r.ReopenedAt,
	}
}

func workDayFromList(r sqlc.ListWorkDaysRow) sqlc.FlowWorkDay {
	return sqlc.FlowWorkDay{
		ID: r.ID, EmployeeID: r.EmployeeID, Date: r.Date, Timezone: r.Timezone,
		WorkMode: r.WorkMode, StartTime: r.StartTime, EndTime: r.EndTime, BreakMinutes: r.BreakMinutes,
		Status: r.Status, LoggedBy: r.LoggedBy, ApprovedBy: r.ApprovedBy, ApprovedAt: r.ApprovedAt,
		ReopenedBy: r.ReopenedBy, ReopenedAt: r.ReopenedAt,
	}
}

func lineFromGet(r sqlc.GetTimeEntryRow) sqlc.HrTimeEntry {
	return sqlc.HrTimeEntry{
		ID: r.ID, WorkDayID: r.WorkDayID, ProjectID: r.ProjectID, TaskID: r.TaskID,
		WorkTypeID: r.WorkTypeID, WorkMode: r.WorkMode, Description: r.Description,
		DurationMinutes: r.DurationMinutes, IsBillable: r.IsBillable, Status: r.Status,
		ClientKey: r.ClientKey, LoggedBy: r.LoggedBy, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func lineFromClientKey(r sqlc.GetTimeEntryByClientKeyRow) sqlc.HrTimeEntry {
	return sqlc.HrTimeEntry{
		ID: r.ID, WorkDayID: r.WorkDayID, ProjectID: r.ProjectID, TaskID: r.TaskID,
		WorkTypeID: r.WorkTypeID, WorkMode: r.WorkMode, Description: r.Description,
		DurationMinutes: r.DurationMinutes, IsBillable: r.IsBillable, Status: r.Status,
		ClientKey: r.ClientKey, LoggedBy: r.LoggedBy, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func workTypeFromGet(r sqlc.GetWorkTypeRow) sqlc.FlowWorkType {
	return sqlc.FlowWorkType{
		ID: r.ID, ParentID: r.ParentID, Name: r.Name, Description: r.Description,
		HrCategoryID: r.HrCategoryID, SortOrder: r.SortOrder, Archived: r.Archived,
	}
}

func workTypeFromList(r sqlc.ListWorkTypesRow) sqlc.FlowWorkType {
	return sqlc.FlowWorkType{
		ID: r.ID, ParentID: r.ParentID, Name: r.Name, Description: r.Description,
		HrCategoryID: r.HrCategoryID, SortOrder: r.SortOrder, Archived: r.Archived,
	}
}

func toDomainWorkDay(row sqlc.FlowWorkDay) domain.WorkDay {
	day := domain.WorkDay{
		ID:           row.ID,
		EmployeeID:   row.EmployeeID,
		Date:         row.Date,
		Timezone:     row.Timezone,
		WorkMode:     row.WorkMode,
		StartTime:    stringValue(row.StartTime),
		EndTime:      stringValue(row.EndTime),
		BreakMinutes: row.BreakMinutes,
		Status:       row.Status,
		LoggedBy:     int64Value(row.LoggedBy),
		ApprovedBy:   int64Value(row.ApprovedBy),
		ApprovedAt:   timeValue(row.ApprovedAt),
		ReopenedBy:   int64Value(row.ReopenedBy),
		ReopenedAt:   timeValue(row.ReopenedAt),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	day.ClockedMinutes = domain.ClockedMinutes(day.StartTime, day.EndTime, day.BreakMinutes)
	return day
}

func toDomainLine(row sqlc.HrTimeEntry) domain.WorkDayLine {
	return domain.WorkDayLine{
		ID:              row.ID,
		WorkDayID:       int64Value(row.WorkDayID),
		ProjectID:       int64Value(row.ProjectID),
		TaskID:          int64Value(row.TaskID),
		WorkTypeID:      int64Value(row.WorkTypeID),
		WorkMode:        stringValue(row.WorkMode),
		Description:     stringValue(row.Description),
		DurationMinutes: row.DurationMinutes,
		IsBillable:      row.IsBillable,
		Status:          row.Status,
		ClientKey:       stringValue(row.ClientKey),
		LoggedBy:        int64Value(row.LoggedBy),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func toDomainWorkType(row sqlc.FlowWorkType) domain.WorkType {
	return domain.WorkType{
		ID:           row.ID,
		ParentID:     int64Value(row.ParentID),
		Name:         row.Name,
		Description:  stringValue(row.Description),
		HRCategoryID: int64Value(row.HrCategoryID),
		SortOrder:    row.SortOrder,
		Archived:     row.Archived,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

// GetWorkDay reads one person's day. Returns ok=false when nothing has been
// recorded for that date, which is an ordinary state rather than an error.
func (r *Repository) GetWorkDay(ctx context.Context, ownerID, businessID, employeeID int64, date string) (domain.WorkDay, bool, error) {
	parsed, err := parseDate(date)
	if err != nil {
		return domain.WorkDay{}, false, err
	}
	row, err := r.queries.GetWorkDay(ctx, sqlc.GetWorkDayParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		EmployeeID: employeeID,
		Date:       parsed,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.WorkDay{}, false, nil
		}
		logger.Ctx(ctx).Error("GetWorkDay failed", logger.Error(err))
		return domain.WorkDay{}, false, err
	}
	return toDomainWorkDay(workDayFromGet(row)), true, nil
}

// UpsertWorkDay creates the day or updates it in place. A day is identified by
// who and when, not by an id the client holds, so the first save and every later
// correction are the same call.
func (r *Repository) UpsertWorkDay(ctx context.Context, input domain.WorkDayInput) (domain.WorkDay, error) {
	parsed, err := parseDate(input.Date)
	if err != nil {
		return domain.WorkDay{}, err
	}
	if _, err := r.queries.UpsertWorkDay(ctx, sqlc.UpsertWorkDayParams{
		OwnerID:      input.OwnerID,
		BusinessID:   input.BusinessID,
		EmployeeID:   input.EmployeeID,
		Date:         parsed,
		Timezone:     input.Timezone,
		WorkMode:     input.WorkMode,
		StartTime:    sqlutil.NullString(input.StartTime),
		EndTime:      sqlutil.NullString(input.EndTime),
		BreakMinutes: input.BreakMinutes,
		LoggedBy:     nullInt64(input.LoggedBy),
	}); err != nil {
		logger.Ctx(ctx).Error("UpsertWorkDay failed", logger.Error(err))
		return domain.WorkDay{}, err
	}
	// Read back rather than trusting LastInsertId: on the update path it is zero,
	// and the caller needs the row either way.
	day, _, err := r.GetWorkDay(ctx, input.OwnerID, input.BusinessID, input.EmployeeID, input.Date)
	return day, err
}

// ListWorkDayLines returns the day's lines with its logged total. The total is
// summed in SQL so it cannot disagree with a paginated read of the lines.
func (r *Repository) ListWorkDayLines(ctx context.Context, businessID, workDayID int64) ([]domain.WorkDayLine, int32, error) {
	rows, err := r.queries.ListWorkDayLines(ctx, sqlc.ListWorkDayLinesParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		WorkDayID:  nullInt64(workDayID),
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListWorkDayLines failed", logger.Error(err))
		return nil, 0, err
	}
	lines := make([]domain.WorkDayLine, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, toDomainLine(lineFromList(row)))
	}

	total, err := r.queries.SumWorkDayLineMinutes(ctx, sqlc.SumWorkDayLineMinutesParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		WorkDayID:  nullInt64(workDayID),
	})
	if err != nil {
		logger.Ctx(ctx).Error("SumWorkDayLineMinutes failed", logger.Error(err))
		return lines, 0, err
	}
	return lines, int32(sumToInt64(total)), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

// SumWorkDayLineMinutesByDayIDs returns logged totals for a page of work days
// with one grouped query.
func (r *Repository) SumWorkDayLineMinutesByDayIDs(ctx context.Context, ownerID, businessID int64, workDayIDs []int64) (map[int64]int32, error) {
	totals := make(map[int64]int32, len(workDayIDs))
	if len(workDayIDs) == 0 {
		return totals, nil
	}
	workDayIDParams := make([]sql.NullInt64, 0, len(workDayIDs))
	for _, workDayID := range workDayIDs {
		workDayIDParams = append(workDayIDParams, nullInt64(workDayID))
	}

	rows, err := r.queries.SumWorkDayLineMinutesByDayIDs(ctx, sqlc.SumWorkDayLineMinutesByDayIDsParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		WorkDayIds: workDayIDParams,
	})
	if err != nil {
		logger.Ctx(ctx).Error("SumWorkDayLineMinutesByDayIDs failed", logger.Error(err))
		return nil, err
	}
	for _, row := range rows {
		if row.WorkDayID.Valid {
			totals[row.WorkDayID.Int64] = int32(sumToInt64(row.LoggedMinutes)) // #nosec G115 -- Reviewed width conversion matches the domain's minute field and established database bounds.
		}
	}
	return totals, nil
}

// CreateWorkDayLine adds a line. A repeated client key returns the line that
// already exists rather than an error: an offline client retrying a send it
// never saw acknowledged has not made a mistake, and the second call must be a
// no-op rather than a second entry.
func (r *Repository) CreateWorkDayLine(ctx context.Context, input domain.WorkDayLineInput) (domain.WorkDayLine, error) {
	if input.ClientKey != "" {
		existing, err := r.queries.GetTimeEntryByClientKey(ctx, sqlc.GetTimeEntryByClientKeyParams{
			OwnerID:    input.OwnerID,
			BusinessID: input.BusinessID,
			ClientKey:  sqlutil.NullString(input.ClientKey),
		})
		if err == nil {
			return toDomainLine(lineFromClientKey(existing)), nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Ctx(ctx).Error("GetTimeEntryByClientKey failed", logger.Error(err))
			return domain.WorkDayLine{}, err
		}
	}

	parsed, err := parseDate(input.Date)
	if err != nil {
		return domain.WorkDayLine{}, err
	}

	res, err := r.queries.CreateWorkDayLine(ctx, sqlc.CreateWorkDayLineParams{
		OwnerID:         input.OwnerID,
		BusinessID:      input.BusinessID,
		EmployeeID:      input.EmployeeID,
		WorkDayID:       nullInt64(input.WorkDayID),
		Date:            parsed,
		ProjectID:       nullInt64(input.ProjectID),
		TaskID:          nullInt64(input.TaskID),
		WorkTypeID:      nullInt64(input.WorkTypeID),
		WorkMode:        sqlutil.NullString(input.WorkMode),
		Timezone:        sqlutil.NullString(input.Timezone),
		DurationMinutes: input.DurationMinutes,
		Description:     sqlutil.NullString(input.Description),
		IsBillable:      input.IsBillable,
		ClientKey:       sqlutil.NullString(input.ClientKey),
		LoggedBy:        nullInt64(input.LoggedBy),
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateWorkDayLine failed", logger.Error(err))
		return domain.WorkDayLine{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.WorkDayLine{}, err
	}
	row, err := r.queries.GetTimeEntry(ctx, sqlc.GetTimeEntryParams{ID: id, OwnerID: input.OwnerID, BusinessID: input.BusinessID})
	if err != nil {
		return domain.WorkDayLine{}, err
	}
	return toDomainLine(lineFromGet(row)), nil
}

// UpdateWorkDayLine changes a line in place.
func (r *Repository) UpdateWorkDayLine(ctx context.Context, input domain.WorkDayLineInput) (domain.WorkDayLine, error) {
	if _, err := r.queries.UpdateWorkDayLine(ctx, sqlc.UpdateWorkDayLineParams{
		ProjectID:       nullInt64(input.ProjectID),
		TaskID:          nullInt64(input.TaskID),
		WorkTypeID:      nullInt64(input.WorkTypeID),
		WorkMode:        sqlutil.NullString(input.WorkMode),
		DurationMinutes: input.DurationMinutes,
		Description:     sqlutil.NullString(input.Description),
		IsBillable:      input.IsBillable,
		ID:              input.ID,
		OwnerID:         input.OwnerID,
		BusinessID:      input.BusinessID,
	}); err != nil {
		logger.Ctx(ctx).Error("UpdateWorkDayLine failed", logger.Error(err))
		return domain.WorkDayLine{}, err
	}
	row, err := r.queries.GetTimeEntry(ctx, sqlc.GetTimeEntryParams{ID: input.ID, OwnerID: input.OwnerID, BusinessID: input.BusinessID})
	if err != nil {
		return domain.WorkDayLine{}, err
	}
	return toDomainLine(lineFromGet(row)), nil
}

// DeleteWorkDayLine removes a line. Reports whether a row was actually removed,
// so the caller can answer 404 rather than pretending a foreign id succeeded.
func (r *Repository) DeleteWorkDayLine(ctx context.Context, id, ownerID, businessID int64) (bool, error) {
	res, err := r.queries.DeleteWorkDayLine(ctx, sqlc.DeleteWorkDayLineParams{
		ID:         id,
		OwnerID:    ownerID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteWorkDayLine failed", logger.Error(err))
		return false, err
	}
	affected, err := res.RowsAffected()
	return affected > 0, err
}

// GetWorkDayByID reads a day by id, scoped to the tenant.
func (r *Repository) GetWorkDayByID(ctx context.Context, id, ownerID, businessID int64) (domain.WorkDay, bool, error) {
	row, err := r.queries.GetWorkDayByID(ctx, sqlc.GetWorkDayByIDParams{
		ID:         id,
		OwnerID:    ownerID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.WorkDay{}, false, nil
		}
		return domain.WorkDay{}, false, err
	}
	return toDomainWorkDay(workDayFromGetByID(row)), true, nil
}

// ListWorkTypes returns the workspace's classification tree.
func (r *Repository) ListWorkTypes(ctx context.Context, ownerID, businessID int64, includeArchived bool) ([]domain.WorkType, error) {
	rows, err := r.queries.ListWorkTypes(ctx, sqlc.ListWorkTypesParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		Column3:    includeArchived,
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListWorkTypes failed", logger.Error(err))
		return nil, err
	}
	types := make([]domain.WorkType, 0, len(rows))
	for _, row := range rows {
		types = append(types, toDomainWorkType(workTypeFromList(row)))
	}
	return types, nil
}

// CreateWorkType adds a node to the tree.
func (r *Repository) CreateWorkType(ctx context.Context, input domain.WorkTypeInput) (domain.WorkType, error) {
	res, err := r.queries.CreateWorkType(ctx, sqlc.CreateWorkTypeParams{
		OwnerID:      input.OwnerID,
		BusinessID:   input.BusinessID,
		ParentID:     nullInt64(input.ParentID),
		Name:         input.Name,
		Description:  sqlutil.NullString(input.Description),
		HrCategoryID: nullInt64(input.HRCategoryID),
		SortOrder:    input.SortOrder,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateWorkType failed", logger.Error(err))
		return domain.WorkType{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.WorkType{}, err
	}
	return r.GetWorkType(ctx, id, input.OwnerID, input.BusinessID)
}

// GetWorkType reads one node, scoped to the tenant.
func (r *Repository) GetWorkType(ctx context.Context, id, ownerID, businessID int64) (domain.WorkType, error) {
	row, err := r.queries.GetWorkType(ctx, sqlc.GetWorkTypeParams{
		ID:         id,
		OwnerID:    ownerID,
		BusinessID: businessID,
	})
	if err != nil {
		return domain.WorkType{}, err
	}
	return toDomainWorkType(workTypeFromGet(row)), nil
}

// UpdateWorkType changes a node.
func (r *Repository) UpdateWorkType(ctx context.Context, input domain.WorkTypeInput) (domain.WorkType, error) {
	if _, err := r.queries.UpdateWorkType(ctx, sqlc.UpdateWorkTypeParams{
		Name:         input.Name,
		Description:  sqlutil.NullString(input.Description),
		ParentID:     nullInt64(input.ParentID),
		HrCategoryID: nullInt64(input.HRCategoryID),
		SortOrder:    input.SortOrder,
		Archived:     input.Archived,
		ID:           input.ID,
		OwnerID:      input.OwnerID,
		BusinessID:   input.BusinessID,
	}); err != nil {
		logger.Ctx(ctx).Error("UpdateWorkType failed", logger.Error(err))
		return domain.WorkType{}, err
	}
	return r.GetWorkType(ctx, input.ID, input.OwnerID, input.BusinessID)
}

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
