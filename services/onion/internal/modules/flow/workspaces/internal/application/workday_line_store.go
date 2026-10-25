package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

var ErrWorkDayLineStoreUnavailable = domain.ErrWorkDayLineStoreUnavailable

// WorkDayLineStore is the Flow-owned port for persisted work-day lines. The
// HR Time module implements it at composition; Flow retains work-day policy.
type WorkDayLineStore interface {
	ListWorkDayLines(ctx context.Context, businessID, workDayID int64) ([]domain.WorkDayLine, int32, error)
	SumWorkDayLineMinutesByDayIDs(ctx context.Context, businessID int64, workDayIDs []int64) (map[int64]int32, error)
	SumWorkDayLineMinutesByMemberAndTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) ([]domain.WorkDayMemberMinutes, error)
	SumWorkDayLineMinutesByTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) (map[int64]int64, error)
	CreateWorkDayLine(ctx context.Context, input domain.WorkDayLineInput) (domain.WorkDayLine, error)
	UpdateWorkDayLine(ctx context.Context, input domain.WorkDayLineInput) (domain.WorkDayLine, error)
	GetWorkDayLineOwner(ctx context.Context, id, businessID int64) (domain.WorkDayLineOwner, bool, error)
	DeleteWorkDayLine(ctx context.Context, id, businessID int64) (bool, error)
}
