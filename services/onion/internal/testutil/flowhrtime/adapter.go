package flowhrtime

import (
	"context"
	"database/sql"

	flow "github.com/voxire/onion/internal/modules/flow/workspaces"
	hrtime "github.com/voxire/onion/internal/modules/hr/hr_time"
)

type Adapter struct {
	module *hrtime.Module
}

func New(ctx context.Context, db *sql.DB) *Adapter {
	return &Adapter{module: hrtime.NewModule(ctx, db, nil)}
}

func (a *Adapter) ListWorkDayLines(ctx context.Context, businessID, workDayID int64) ([]flow.WorkDayLine, int32, error) {
	entries, total, err := a.module.ListFlowWorkdayEntries(ctx, businessID, workDayID)
	if err != nil {
		return nil, 0, err
	}
	lines := make([]flow.WorkDayLine, 0, len(entries))
	for _, entry := range entries {
		lines = append(lines, toFlow(entry))
	}
	return lines, total, nil
}

func (a *Adapter) SumWorkDayLineMinutesByDayIDs(ctx context.Context, businessID int64, workDayIDs []int64) (map[int64]int32, error) {
	return a.module.SumFlowWorkdayEntryMinutesByDayIDs(ctx, businessID, workDayIDs)
}

func (a *Adapter) SumWorkDayLineMinutesByTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) (map[int64]int64, error) {
	return a.module.SumFlowWorkdayEntryMinutesByTaskIDs(ctx, businessID, taskIDs)
}

func (a *Adapter) SumWorkDayLineMinutesByMemberAndTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) ([]flow.WorkDayMemberMinutes, error) {
	rows, err := a.module.SumFlowWorkdayEntryMinutesByMemberAndTaskIDs(ctx, businessID, taskIDs)
	if err != nil {
		return nil, err
	}
	minutes := make([]flow.WorkDayMemberMinutes, 0, len(rows))
	for _, row := range rows {
		minutes = append(minutes, flow.WorkDayMemberMinutes{UserID: row.EmployeeID, Minutes: row.Minutes})
	}
	return minutes, nil
}

func (a *Adapter) CreateWorkDayLine(ctx context.Context, input flow.WorkDayLineInput) (flow.WorkDayLine, error) {
	entry, err := a.module.CreateFlowWorkdayEntry(ctx, toHRInput(input))
	if err != nil {
		return flow.WorkDayLine{}, err
	}
	return toFlow(*entry), nil
}

func (a *Adapter) UpdateWorkDayLine(ctx context.Context, input flow.WorkDayLineInput) (flow.WorkDayLine, error) {
	entry, err := a.module.UpdateFlowWorkdayEntry(ctx, toHRInput(input))
	if err != nil {
		return flow.WorkDayLine{}, err
	}
	return toFlow(*entry), nil
}

func (a *Adapter) GetWorkDayLineOwner(ctx context.Context, id, businessID int64) (flow.WorkDayLineOwner, bool, error) {
	owner, found, err := a.module.FlowWorkdayEntryOwner(ctx, id, businessID)
	return flow.WorkDayLineOwner{WorkDayID: owner.WorkDayID, EmployeeID: owner.EmployeeID, TaskID: owner.TaskID}, found, err
}

func (a *Adapter) DeleteWorkDayLine(ctx context.Context, id, businessID int64) (bool, error) {
	return a.module.DeleteFlowWorkdayEntry(ctx, id, businessID)
}

func toHRInput(input flow.WorkDayLineInput) hrtime.FlowWorkdayEntryInput {
	return hrtime.FlowWorkdayEntryInput{
		ID: input.ID, BusinessID: input.BusinessID,
		EmployeeID: input.EmployeeID, WorkDayID: input.WorkDayID, Date: input.Date,
		ProjectID: input.ProjectID, TaskID: input.TaskID, WorkTypeID: input.WorkTypeID,
		WorkMode: input.WorkMode, Timezone: input.Timezone, Description: input.Description,
		DurationMinutes: input.DurationMinutes, IsBillable: input.IsBillable,
		ClientKey: input.ClientKey, LoggedBy: input.LoggedBy,
	}
}

func toFlow(entry hrtime.FlowWorkdayEntry) flow.WorkDayLine {
	return flow.WorkDayLine{
		ID: entry.ID, WorkDayID: entry.WorkDayID, ProjectID: entry.ProjectID,
		TaskID: entry.TaskID, WorkTypeID: entry.WorkTypeID, WorkMode: entry.WorkMode,
		Description: entry.Description, DurationMinutes: entry.DurationMinutes,
		IsBillable: entry.IsBillable, Status: entry.Status, ClientKey: entry.ClientKey,
		LoggedBy: entry.LoggedBy, CreatedAt: entry.CreatedAt, UpdatedAt: entry.UpdatedAt,
	}
}
