package workspaceadapters

import (
	"context"
	"errors"

	flowworkspaces "github.com/voxire/onion/internal/modules/flow/workspaces"
	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
	hrtime "github.com/voxire/onion/internal/modules/hr/hr_time"
)

type flowHRTimeAdapter struct {
	module *hrtime.Module
	core   *hrcore.Module
}

func (a flowHRTimeAdapter) ListWorkDayLines(ctx context.Context, businessID, workDayID int64) ([]flowworkspaces.WorkDayLine, int32, error) {
	rows, total, err := a.module.ListFlowWorkdayEntries(ctx, businessID, workDayID)
	if err != nil {
		return nil, 0, err
	}
	lines := make([]flowworkspaces.WorkDayLine, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, flowWorkdayLine(row))
	}
	return lines, total, nil
}

func (a flowHRTimeAdapter) SumWorkDayLineMinutesByDayIDs(ctx context.Context, businessID int64, workDayIDs []int64) (map[int64]int32, error) {
	return a.module.SumFlowWorkdayEntryMinutesByDayIDs(ctx, businessID, workDayIDs)
}

func (a flowHRTimeAdapter) SumWorkDayLineMinutesByTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) (map[int64]int64, error) {
	return a.module.SumFlowWorkdayEntryMinutesByTaskIDs(ctx, businessID, taskIDs)
}

func (a flowHRTimeAdapter) SumWorkDayLineMinutesByMemberAndTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) ([]flowworkspaces.WorkDayMemberMinutes, error) {
	rows, err := a.module.SumFlowWorkdayEntryMinutesByMemberAndTaskIDs(ctx, businessID, taskIDs)
	if err != nil {
		return nil, err
	}
	minutes := make([]flowworkspaces.WorkDayMemberMinutes, 0, len(rows))
	for _, row := range rows {
		userID := row.EmployeeID
		if a.core != nil {
			linkedUserID, resolveErr := a.core.GetEmployeeUserID(ctx, businessID, row.EmployeeID)
			if resolveErr == nil && linkedUserID > 0 {
				userID = linkedUserID
			} else if resolveErr != nil && !errors.Is(resolveErr, hrcore.ErrEmployeeNotFound) {
				return nil, resolveErr
			}
		}
		minutes = append(minutes, flowworkspaces.WorkDayMemberMinutes{UserID: userID, Minutes: row.Minutes})
	}
	return minutes, nil
}

func (a flowHRTimeAdapter) CreateWorkDayLine(ctx context.Context, input flowworkspaces.WorkDayLineInput) (flowworkspaces.WorkDayLine, error) {
	row, err := a.module.CreateFlowWorkdayEntry(ctx, hrtime.FlowWorkdayEntryInput{
		ID: input.ID, BusinessID: input.BusinessID,
		EmployeeID: input.EmployeeID, WorkDayID: input.WorkDayID, Date: input.Date,
		ProjectID: input.ProjectID, TaskID: input.TaskID, WorkTypeID: input.WorkTypeID,
		WorkMode: input.WorkMode, Timezone: input.Timezone, Description: input.Description,
		DurationMinutes: input.DurationMinutes, IsBillable: input.IsBillable,
		ClientKey: input.ClientKey, LoggedBy: input.LoggedBy,
	})
	if err != nil {
		return flowworkspaces.WorkDayLine{}, err
	}
	return flowWorkdayLine(*row), nil
}

func (a flowHRTimeAdapter) UpdateWorkDayLine(ctx context.Context, input flowworkspaces.WorkDayLineInput) (flowworkspaces.WorkDayLine, error) {
	row, err := a.module.UpdateFlowWorkdayEntry(ctx, hrtime.FlowWorkdayEntryInput{
		ID: input.ID, BusinessID: input.BusinessID,
		EmployeeID: input.EmployeeID, WorkDayID: input.WorkDayID, Date: input.Date,
		ProjectID: input.ProjectID, TaskID: input.TaskID, WorkTypeID: input.WorkTypeID,
		WorkMode: input.WorkMode, Timezone: input.Timezone, Description: input.Description,
		DurationMinutes: input.DurationMinutes, IsBillable: input.IsBillable,
		ClientKey: input.ClientKey, LoggedBy: input.LoggedBy,
	})
	if err != nil {
		return flowworkspaces.WorkDayLine{}, err
	}
	return flowWorkdayLine(*row), nil
}

func (a flowHRTimeAdapter) GetWorkDayLineOwner(ctx context.Context, id, businessID int64) (flowworkspaces.WorkDayLineOwner, bool, error) {
	owner, found, err := a.module.FlowWorkdayEntryOwner(ctx, id, businessID)
	return flowworkspaces.WorkDayLineOwner{WorkDayID: owner.WorkDayID, EmployeeID: owner.EmployeeID, TaskID: owner.TaskID}, found, err
}

func (a flowHRTimeAdapter) DeleteWorkDayLine(ctx context.Context, id, businessID int64) (bool, error) {
	return a.module.DeleteFlowWorkdayEntry(ctx, id, businessID)
}

func flowWorkdayLine(row hrtime.FlowWorkdayEntry) flowworkspaces.WorkDayLine {
	return flowworkspaces.WorkDayLine{
		ID: row.ID, WorkDayID: row.WorkDayID, ProjectID: row.ProjectID,
		TaskID: row.TaskID, WorkTypeID: row.WorkTypeID, WorkMode: row.WorkMode,
		Description: row.Description, DurationMinutes: row.DurationMinutes,
		IsBillable: row.IsBillable, Status: row.Status, ClientKey: row.ClientKey,
		LoggedBy: row.LoggedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func NewHRTimeWorkDayLineStore(module *hrtime.Module, core *hrcore.Module) flowworkspaces.WorkDayLineStore {
	return flowHRTimeAdapter{module: module, core: core}
}
