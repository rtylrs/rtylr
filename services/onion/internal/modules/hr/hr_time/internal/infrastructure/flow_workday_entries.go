package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func flowWorkdayEntryFromRow(
	id int64,
	workDayID, projectID, taskID, workTypeID sql.NullInt64,
	workMode, description sql.NullString,
	duration int32,
	billable bool,
	status string,
	clientKey sql.NullString,
	loggedBy sql.NullInt64,
	createdAt, updatedAt time.Time,
) domain.FlowWorkdayEntry {
	return domain.FlowWorkdayEntry{
		ID: id, WorkDayID: workDayID.Int64, ProjectID: projectID.Int64,
		TaskID: taskID.Int64, WorkTypeID: workTypeID.Int64,
		WorkMode: workMode.String, Description: description.String,
		DurationMinutes: duration, IsBillable: billable, Status: status,
		ClientKey: clientKey.String, LoggedBy: loggedBy.Int64,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}
}

func flowWorkdayEntryFromList(row sqlc.ListFlowWorkdayEntriesRow) domain.FlowWorkdayEntry {
	return flowWorkdayEntryFromRow(row.ID, row.WorkDayID, row.ProjectID, row.TaskID,
		row.WorkTypeID, row.WorkMode, row.Description, row.DurationMinutes,
		row.IsBillable, row.Status, row.ClientKey, row.LoggedBy, row.CreatedAt, row.UpdatedAt)
}

func flowWorkdayEntryFromGet(row sqlc.GetFlowWorkdayEntryRow) domain.FlowWorkdayEntry {
	return flowWorkdayEntryFromRow(row.ID, row.WorkDayID, row.ProjectID, row.TaskID,
		row.WorkTypeID, row.WorkMode, row.Description, row.DurationMinutes,
		row.IsBillable, row.Status, row.ClientKey, row.LoggedBy, row.CreatedAt, row.UpdatedAt)
}

func (r *Repository) ListFlowWorkdayEntries(ctx context.Context, ownerID, businessID, workDayID int64) ([]domain.FlowWorkdayEntry, int32, error) {
	rows, err := r.queries.ListFlowWorkdayEntries(ctx, sqlc.ListFlowWorkdayEntriesParams{
		OwnerID: ownerID, BusinessID: businessID, WorkDayID: sqlutil.NullInt64(workDayID),
	})
	if err != nil {
		return nil, 0, err
	}
	entries := make([]domain.FlowWorkdayEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, flowWorkdayEntryFromList(row))
	}
	total, err := r.queries.SumFlowWorkdayEntryMinutes(ctx, sqlc.SumFlowWorkdayEntryMinutesParams{
		OwnerID: ownerID, BusinessID: businessID, WorkDayID: sqlutil.NullInt64(workDayID),
	})
	if err != nil {
		return entries, 0, err
	}
	return entries, int32(flowMinutesValue(total)), nil // #nosec G115 -- The stored duration is an int32 per entry; this aggregate follows the existing API field width.
}

func (r *Repository) SumFlowWorkdayEntryMinutesByDayIDs(ctx context.Context, businessID int64, workDayIDs []int64) (map[int64]int32, error) {
	totals := make(map[int64]int32, len(workDayIDs))
	if len(workDayIDs) == 0 {
		return totals, nil
	}
	ids := make([]sql.NullInt64, 0, len(workDayIDs))
	for _, id := range workDayIDs {
		ids = append(ids, sqlutil.NullInt64(id))
	}
	rows, err := r.queries.SumFlowWorkdayEntryMinutesByDayIDs(ctx, sqlc.SumFlowWorkdayEntryMinutesByDayIDsParams{
		BusinessID: businessID, WorkDayIds: ids,
	})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.WorkDayID.Valid {
			totals[row.WorkDayID.Int64] = int32(flowMinutesValue(row.LoggedMinutes)) // #nosec G115 -- The stored duration is an int32 per entry; this aggregate follows the existing API field width.
		}
	}
	return totals, nil
}

func (r *Repository) SumFlowWorkdayEntryMinutesByTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) (map[int64]int64, error) {
	totals := make(map[int64]int64, len(taskIDs))
	if len(taskIDs) == 0 {
		return totals, nil
	}
	ids := make([]sql.NullInt64, 0, len(taskIDs))
	for _, id := range taskIDs {
		ids = append(ids, sqlutil.NullInt64(id))
	}
	rows, err := r.queries.SumFlowWorkdayEntryMinutesByTaskIDs(ctx, sqlc.SumFlowWorkdayEntryMinutesByTaskIDsParams{
		BusinessID: businessID, TaskIds: ids,
	})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.TaskID.Valid {
			totals[row.TaskID.Int64] = flowMinutesValue(row.LoggedMinutes)
		}
	}
	return totals, nil
}

func (r *Repository) SumFlowWorkdayEntryMinutesByMemberAndTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) ([]domain.FlowWorkdayMemberMinutes, error) {
	if len(taskIDs) == 0 {
		return []domain.FlowWorkdayMemberMinutes{}, nil
	}
	ids := make([]sql.NullInt64, 0, len(taskIDs))
	for _, id := range taskIDs {
		ids = append(ids, sqlutil.NullInt64(id))
	}
	rows, err := r.queries.SumFlowWorkdayEntryMinutesByMemberAndTaskIDs(ctx, sqlc.SumFlowWorkdayEntryMinutesByMemberAndTaskIDsParams{
		BusinessID: businessID,
		TaskIds:    ids,
	})
	if err != nil {
		return nil, err
	}
	minutes := make([]domain.FlowWorkdayMemberMinutes, 0, len(rows))
	for _, row := range rows {
		minutes = append(minutes, domain.FlowWorkdayMemberMinutes{EmployeeID: row.EmployeeID, Minutes: flowMinutesValue(row.LoggedMinutes)})
	}
	return minutes, nil
}

func (r *Repository) CreateFlowWorkdayEntry(ctx context.Context, input domain.FlowWorkdayEntryInput) (*domain.FlowWorkdayEntry, error) {
	date, err := time.Parse(time.DateOnly, input.Date)
	if err != nil {
		return nil, err
	}
	result, err := r.queries.CreateFlowWorkdayEntry(ctx, sqlc.CreateFlowWorkdayEntryParams{
		OwnerID: input.OwnerID, BusinessID: input.BusinessID, EmployeeID: input.EmployeeID,
		WorkDayID: sqlutil.NullInt64(input.WorkDayID), Date: date,
		ProjectID: sqlutil.NullInt64(input.ProjectID), TaskID: sqlutil.NullInt64(input.TaskID),
		WorkTypeID: sqlutil.NullInt64(input.WorkTypeID), WorkMode: sqlutil.NullString(input.WorkMode),
		Timezone: sqlutil.NullString(input.Timezone), DurationMinutes: input.DurationMinutes,
		Description: sqlutil.NullString(input.Description), IsBillable: input.IsBillable,
		ClientKey: sqlutil.NullString(input.ClientKey), LoggedBy: sqlutil.NullInt64(input.LoggedBy),
	})
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	row, err := r.queries.GetFlowWorkdayEntry(ctx, sqlc.GetFlowWorkdayEntryParams{
		ID: id, OwnerID: input.OwnerID, BusinessID: input.BusinessID,
	})
	if err != nil {
		return nil, err
	}
	entry := flowWorkdayEntryFromGet(row)
	return &entry, nil
}

func (r *Repository) UpdateFlowWorkdayEntry(ctx context.Context, input domain.FlowWorkdayEntryInput) (*domain.FlowWorkdayEntry, error) {
	_, err := r.queries.UpdateFlowWorkdayEntry(ctx, sqlc.UpdateFlowWorkdayEntryParams{
		ProjectID: sqlutil.NullInt64(input.ProjectID), TaskID: sqlutil.NullInt64(input.TaskID),
		WorkTypeID: sqlutil.NullInt64(input.WorkTypeID), WorkMode: sqlutil.NullString(input.WorkMode),
		DurationMinutes: input.DurationMinutes, Description: sqlutil.NullString(input.Description),
		IsBillable: input.IsBillable, ID: input.ID, OwnerID: input.OwnerID, BusinessID: input.BusinessID,
	})
	if err != nil {
		return nil, err
	}
	row, err := r.queries.GetFlowWorkdayEntry(ctx, sqlc.GetFlowWorkdayEntryParams{
		ID: input.ID, OwnerID: input.OwnerID, BusinessID: input.BusinessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTimeEntryNotFound
		}
		return nil, err
	}
	entry := flowWorkdayEntryFromGet(row)
	return &entry, nil
}

func (r *Repository) FlowWorkdayEntryOwner(ctx context.Context, id, ownerID, businessID int64) (domain.FlowWorkdayEntryOwner, bool, error) {
	row, err := r.queries.GetFlowWorkdayEntryOwner(ctx, sqlc.GetFlowWorkdayEntryOwnerParams{
		ID: id, OwnerID: ownerID, BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.FlowWorkdayEntryOwner{}, false, nil
		}
		return domain.FlowWorkdayEntryOwner{}, false, err
	}
	return domain.FlowWorkdayEntryOwner{WorkDayID: row.WorkDayID.Int64, EmployeeID: row.EmployeeID, TaskID: row.TaskID.Int64}, true, nil
}

func (r *Repository) DeleteFlowWorkdayEntry(ctx context.Context, id, ownerID, businessID int64) (bool, error) {
	result, err := r.queries.DeleteFlowWorkdayEntry(ctx, sqlc.DeleteFlowWorkdayEntryParams{
		ID: id, OwnerID: ownerID, BusinessID: businessID,
	})
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows > 0, err
}

func flowMinutesValue(value interface{}) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	case []byte:
		parsed, err := strconv.ParseInt(string(v), 10, 64)
		if err == nil {
			return parsed
		}
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return parsed
		}
	}
	return 0
}
