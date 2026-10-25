package infrastructure_test

import (
	"context"
	"database/sql"

	flow "github.com/voxire/onion/internal/modules/flow/workspaces"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil/flowhrtime"
)

type testHRTimeStore struct {
	store *flowhrtime.Adapter
}

func newTestHRTimeStore(ctx context.Context, db *sql.DB) testHRTimeStore {
	return testHRTimeStore{store: flowhrtime.New(ctx, db)}
}

func (s testHRTimeStore) ListWorkDayLines(ctx context.Context, businessID, workDayID int64) ([]domain.WorkDayLine, int32, error) {
	lines, total, err := s.store.ListWorkDayLines(ctx, businessID, workDayID)
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.WorkDayLine, 0, len(lines))
	for _, line := range lines {
		out = append(out, testDomainLine(line))
	}
	return out, total, nil
}

func (s testHRTimeStore) SumWorkDayLineMinutesByDayIDs(ctx context.Context, businessID int64, ids []int64) (map[int64]int32, error) {
	return s.store.SumWorkDayLineMinutesByDayIDs(ctx, businessID, ids)
}

func (s testHRTimeStore) SumWorkDayLineMinutesByTaskIDs(ctx context.Context, businessID int64, ids []int64) (map[int64]int64, error) {
	return s.store.SumWorkDayLineMinutesByTaskIDs(ctx, businessID, ids)
}

func (s testHRTimeStore) SumWorkDayLineMinutesByMemberAndTaskIDs(ctx context.Context, businessID int64, ids []int64) ([]domain.WorkDayMemberMinutes, error) {
	rows, err := s.store.SumWorkDayLineMinutesByMemberAndTaskIDs(ctx, businessID, ids)
	if err != nil {
		return nil, err
	}
	minutes := make([]domain.WorkDayMemberMinutes, 0, len(rows))
	for _, row := range rows {
		minutes = append(minutes, domain.WorkDayMemberMinutes{UserID: row.UserID, Minutes: row.Minutes})
	}
	return minutes, nil
}

func (s testHRTimeStore) CreateWorkDayLine(ctx context.Context, input domain.WorkDayLineInput) (domain.WorkDayLine, error) {
	line, err := s.store.CreateWorkDayLine(ctx, testFlowLineInput(input))
	if err != nil {
		return domain.WorkDayLine{}, err
	}
	return testDomainLine(line), nil
}

func (s testHRTimeStore) UpdateWorkDayLine(ctx context.Context, input domain.WorkDayLineInput) (domain.WorkDayLine, error) {
	line, err := s.store.UpdateWorkDayLine(ctx, testFlowLineInput(input))
	if err != nil {
		return domain.WorkDayLine{}, err
	}
	return testDomainLine(line), nil
}

func (s testHRTimeStore) GetWorkDayLineOwner(ctx context.Context, id, businessID int64) (domain.WorkDayLineOwner, bool, error) {
	owner, found, err := s.store.GetWorkDayLineOwner(ctx, id, businessID)
	return domain.WorkDayLineOwner{WorkDayID: owner.WorkDayID, EmployeeID: owner.EmployeeID, TaskID: owner.TaskID}, found, err
}

func (s testHRTimeStore) DeleteWorkDayLine(ctx context.Context, id, businessID int64) (bool, error) {
	return s.store.DeleteWorkDayLine(ctx, id, businessID)
}

func testFlowLineInput(input domain.WorkDayLineInput) flow.WorkDayLineInput {
	return flow.WorkDayLineInput{
		ID: input.ID, BusinessID: input.BusinessID,
		EmployeeID: input.EmployeeID, WorkDayID: input.WorkDayID, Date: input.Date,
		ProjectID: input.ProjectID, TaskID: input.TaskID, WorkTypeID: input.WorkTypeID,
		WorkMode: input.WorkMode, Timezone: input.Timezone, Description: input.Description,
		DurationMinutes: input.DurationMinutes, IsBillable: input.IsBillable,
		ClientKey: input.ClientKey, LoggedBy: input.LoggedBy,
	}
}

func testDomainLine(line flow.WorkDayLine) domain.WorkDayLine {
	return domain.WorkDayLine{
		ID: line.ID, WorkDayID: line.WorkDayID, ProjectID: line.ProjectID,
		TaskID: line.TaskID, WorkTypeID: line.WorkTypeID, WorkMode: line.WorkMode,
		Description: line.Description, DurationMinutes: line.DurationMinutes,
		IsBillable: line.IsBillable, Status: line.Status, ClientKey: line.ClientKey,
		LoggedBy: line.LoggedBy, CreatedAt: line.CreatedAt, UpdatedAt: line.UpdatedAt,
	}
}
