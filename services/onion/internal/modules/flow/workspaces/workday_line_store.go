package workspaces

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// WorkDayLine is Flow's boundary representation of an HR-owned work entry.
type WorkDayLine struct {
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

type WorkDayMemberMinutes struct {
	UserID  int64
	Minutes int64
}

type WorkDayLineInput struct {
	ID              int64
	OwnerID         int64
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

type WorkDayLineOwner struct {
	WorkDayID  int64
	EmployeeID int64
	TaskID     int64
}

type WorkDayLineStore interface {
	ListWorkDayLines(ctx context.Context, ownerID, businessID, workDayID int64) ([]WorkDayLine, int32, error)
	SumWorkDayLineMinutesByDayIDs(ctx context.Context, businessID int64, workDayIDs []int64) (map[int64]int32, error)
	SumWorkDayLineMinutesByMemberAndTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) ([]WorkDayMemberMinutes, error)
	SumWorkDayLineMinutesByTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) (map[int64]int64, error)
	CreateWorkDayLine(ctx context.Context, input WorkDayLineInput) (WorkDayLine, error)
	UpdateWorkDayLine(ctx context.Context, input WorkDayLineInput) (WorkDayLine, error)
	GetWorkDayLineOwner(ctx context.Context, id, ownerID, businessID int64) (WorkDayLineOwner, bool, error)
	DeleteWorkDayLine(ctx context.Context, id, ownerID, businessID int64) (bool, error)
}

type workDayLineStoreAdapter struct {
	store WorkDayLineStore
}

func (a workDayLineStoreAdapter) ListWorkDayLines(ctx context.Context, ownerID, businessID, workDayID int64) ([]domain.WorkDayLine, int32, error) {
	lines, total, err := a.store.ListWorkDayLines(ctx, ownerID, businessID, workDayID)
	if err != nil {
		return nil, 0, err
	}
	out := make([]domain.WorkDayLine, 0, len(lines))
	for _, line := range lines {
		out = append(out, workDayLineToDomain(line))
	}
	return out, total, nil
}

func (a workDayLineStoreAdapter) SumWorkDayLineMinutesByDayIDs(ctx context.Context, businessID int64, ids []int64) (map[int64]int32, error) {
	return a.store.SumWorkDayLineMinutesByDayIDs(ctx, businessID, ids)
}

func (a workDayLineStoreAdapter) SumWorkDayLineMinutesByMemberAndTaskIDs(ctx context.Context, businessID int64, ids []int64) ([]domain.WorkDayMemberMinutes, error) {
	rows, err := a.store.SumWorkDayLineMinutesByMemberAndTaskIDs(ctx, businessID, ids)
	if err != nil {
		return nil, err
	}
	minutes := make([]domain.WorkDayMemberMinutes, 0, len(rows))
	for _, row := range rows {
		minutes = append(minutes, domain.WorkDayMemberMinutes{UserID: row.UserID, Minutes: row.Minutes})
	}
	return minutes, nil
}

func (a workDayLineStoreAdapter) SumWorkDayLineMinutesByTaskIDs(ctx context.Context, businessID int64, ids []int64) (map[int64]int64, error) {
	return a.store.SumWorkDayLineMinutesByTaskIDs(ctx, businessID, ids)
}

func (a workDayLineStoreAdapter) CreateWorkDayLine(ctx context.Context, input domain.WorkDayLineInput) (domain.WorkDayLine, error) {
	line, err := a.store.CreateWorkDayLine(ctx, workDayLineInputFromDomain(input))
	if err != nil {
		return domain.WorkDayLine{}, err
	}
	return workDayLineToDomain(line), nil
}

func (a workDayLineStoreAdapter) UpdateWorkDayLine(ctx context.Context, input domain.WorkDayLineInput) (domain.WorkDayLine, error) {
	line, err := a.store.UpdateWorkDayLine(ctx, workDayLineInputFromDomain(input))
	if err != nil {
		return domain.WorkDayLine{}, err
	}
	return workDayLineToDomain(line), nil
}

func (a workDayLineStoreAdapter) GetWorkDayLineOwner(ctx context.Context, id, ownerID, businessID int64) (domain.WorkDayLineOwner, bool, error) {
	owner, found, err := a.store.GetWorkDayLineOwner(ctx, id, ownerID, businessID)
	return domain.WorkDayLineOwner{WorkDayID: owner.WorkDayID, EmployeeID: owner.EmployeeID, TaskID: owner.TaskID}, found, err
}

func (a workDayLineStoreAdapter) DeleteWorkDayLine(ctx context.Context, id, ownerID, businessID int64) (bool, error) {
	return a.store.DeleteWorkDayLine(ctx, id, ownerID, businessID)
}

func workDayLineToDomain(line WorkDayLine) domain.WorkDayLine {
	return domain.WorkDayLine{
		ID: line.ID, WorkDayID: line.WorkDayID, ProjectID: line.ProjectID,
		TaskID: line.TaskID, WorkTypeID: line.WorkTypeID, WorkMode: line.WorkMode,
		Description: line.Description, DurationMinutes: line.DurationMinutes,
		IsBillable: line.IsBillable, Status: line.Status, ClientKey: line.ClientKey,
		LoggedBy: line.LoggedBy, CreatedAt: line.CreatedAt, UpdatedAt: line.UpdatedAt,
	}
}

func workDayLineInputFromDomain(input domain.WorkDayLineInput) WorkDayLineInput {
	return WorkDayLineInput{
		ID: input.ID, OwnerID: input.OwnerID, BusinessID: input.BusinessID,
		EmployeeID: input.EmployeeID, WorkDayID: input.WorkDayID, Date: input.Date,
		ProjectID: input.ProjectID, TaskID: input.TaskID, WorkTypeID: input.WorkTypeID,
		WorkMode: input.WorkMode, Timezone: input.Timezone, Description: input.Description,
		DurationMinutes: input.DurationMinutes, IsBillable: input.IsBillable,
		ClientKey: input.ClientKey, LoggedBy: input.LoggedBy,
	}
}
