package hr_time

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
)

// FlowWorkdayEntry is the narrow HR Time capability Flow needs to display and
// edit work-day entries. It contains no persistence or transport types.
type FlowWorkdayEntry struct {
	ID              int64
	WorkDayID       int64
	ProjectID       int64
	TaskID          int64
	WorkTypeID      int64
	WorkMode        string
	Description     string
	DurationMinutes int32
	IsBillable      bool
	Status          string
	ClientKey       string
	LoggedBy        int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// FlowWorkdayEntryInput has no owner: an HR time entry's owner_id is its
// business ID, so the business is the only scope Flow passes.
type FlowWorkdayEntryInput struct {
	ID              int64
	BusinessID      int64
	EmployeeID      int64
	WorkDayID       int64
	Date            string
	ProjectID       int64
	TaskID          int64
	WorkTypeID      int64
	WorkMode        string
	Timezone        string
	Description     string
	DurationMinutes int32
	IsBillable      bool
	ClientKey       string
	LoggedBy        int64
}

type FlowWorkdayEntryOwner struct {
	WorkDayID  int64
	EmployeeID int64
	TaskID     int64
}

type FlowWorkdayMemberMinutes struct {
	EmployeeID int64
	Minutes    int64
}

func (m *Module) ListFlowWorkdayEntries(ctx context.Context, businessID, workDayID int64) ([]FlowWorkdayEntry, int32, error) {
	rows, total, err := m.service.ListFlowWorkdayEntries(ctx, businessID, businessID, workDayID)
	if err != nil {
		return nil, 0, err
	}
	entries := make([]FlowWorkdayEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, flowWorkdayEntryFromDomain(row))
	}
	return entries, total, nil
}

func (m *Module) SumFlowWorkdayEntryMinutesByDayIDs(ctx context.Context, businessID int64, workDayIDs []int64) (map[int64]int32, error) {
	return m.service.SumFlowWorkdayEntryMinutesByDayIDs(ctx, businessID, workDayIDs)
}

func (m *Module) SumFlowWorkdayEntryMinutesByTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) (map[int64]int64, error) {
	return m.service.SumFlowWorkdayEntryMinutesByTaskIDs(ctx, businessID, taskIDs)
}

func (m *Module) SumFlowWorkdayEntryMinutesByMemberAndTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) ([]FlowWorkdayMemberMinutes, error) {
	rows, err := m.service.SumFlowWorkdayEntryMinutesByMemberAndTaskIDs(ctx, businessID, taskIDs)
	if err != nil {
		return nil, err
	}
	minutes := make([]FlowWorkdayMemberMinutes, 0, len(rows))
	for _, row := range rows {
		minutes = append(minutes, FlowWorkdayMemberMinutes{EmployeeID: row.EmployeeID, Minutes: row.Minutes})
	}
	return minutes, nil
}

func (m *Module) CreateFlowWorkdayEntry(ctx context.Context, input FlowWorkdayEntryInput) (*FlowWorkdayEntry, error) {
	row, err := m.service.CreateFlowWorkdayEntry(ctx, flowWorkdayEntryInputToDomain(input))
	if err != nil {
		return nil, err
	}
	result := flowWorkdayEntryFromDomain(*row)
	return &result, nil
}

func (m *Module) UpdateFlowWorkdayEntry(ctx context.Context, input FlowWorkdayEntryInput) (*FlowWorkdayEntry, error) {
	row, err := m.service.UpdateFlowWorkdayEntry(ctx, flowWorkdayEntryInputToDomain(input))
	if err != nil {
		return nil, err
	}
	result := flowWorkdayEntryFromDomain(*row)
	return &result, nil
}

func (m *Module) FlowWorkdayEntryOwner(ctx context.Context, id, businessID int64) (FlowWorkdayEntryOwner, bool, error) {
	owner, found, err := m.service.FlowWorkdayEntryOwner(ctx, id, businessID, businessID)
	return FlowWorkdayEntryOwner{WorkDayID: owner.WorkDayID, EmployeeID: owner.EmployeeID, TaskID: owner.TaskID}, found, err
}

func (m *Module) DeleteFlowWorkdayEntry(ctx context.Context, id, businessID int64) (bool, error) {
	return m.service.DeleteFlowWorkdayEntry(ctx, id, businessID, businessID)
}

func flowWorkdayEntryInputToDomain(input FlowWorkdayEntryInput) domain.FlowWorkdayEntryInput {
	return domain.FlowWorkdayEntryInput{
		ID: input.ID, OwnerID: input.BusinessID, BusinessID: input.BusinessID,
		EmployeeID: input.EmployeeID, WorkDayID: input.WorkDayID, Date: input.Date,
		ProjectID: input.ProjectID, TaskID: input.TaskID, WorkTypeID: input.WorkTypeID,
		WorkMode: input.WorkMode, Timezone: input.Timezone, Description: input.Description,
		DurationMinutes: input.DurationMinutes, IsBillable: input.IsBillable,
		ClientKey: input.ClientKey, LoggedBy: input.LoggedBy,
	}
}

func flowWorkdayEntryFromDomain(row domain.FlowWorkdayEntry) FlowWorkdayEntry {
	return FlowWorkdayEntry{
		ID: row.ID, WorkDayID: row.WorkDayID, ProjectID: row.ProjectID,
		TaskID: row.TaskID, WorkTypeID: row.WorkTypeID, WorkMode: row.WorkMode,
		Description: row.Description, DurationMinutes: row.DurationMinutes,
		IsBillable: row.IsBillable, Status: row.Status, ClientKey: row.ClientKey,
		LoggedBy: row.LoggedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}
