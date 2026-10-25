package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

// GetWorkDay reads one person's day. Returns ok=false when nothing has been
// recorded for that date, which is an ordinary state rather than an error.
func (r *Repository) GetWorkDay(ctx context.Context, businessID, employeeID int64, date string) (domain.WorkDay, bool, error) {
	parsed, err := parseDate(date)
	if err != nil {
		return domain.WorkDay{}, false, err
	}
	row, err := r.queries.GetWorkDay(ctx, sqlc.GetWorkDayParams{
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
	day, _, err := r.GetWorkDay(ctx, input.BusinessID, input.EmployeeID, input.Date)
	return day, err
}

// Work-day line persistence belongs to HR Time. Flow's repository retains the
// port as an adapter so its use cases still combine Flow-owned day policy with
// HR-owned entry state without querying the HR schema.
func (r *Repository) ListWorkDayLines(ctx context.Context, businessID, workDayID int64) ([]domain.WorkDayLine, int32, error) {
	if r.workDayLineStore == nil {
		return nil, 0, domain.ErrWorkDayLineStoreUnavailable
	}
	return r.workDayLineStore.ListWorkDayLines(ctx, businessID, workDayID)
}

func (r *Repository) SumWorkDayLineMinutesByDayIDs(ctx context.Context, businessID int64, workDayIDs []int64) (map[int64]int32, error) {
	if r.workDayLineStore == nil {
		return nil, domain.ErrWorkDayLineStoreUnavailable
	}
	return r.workDayLineStore.SumWorkDayLineMinutesByDayIDs(ctx, businessID, workDayIDs)
}

func (r *Repository) SumWorkDayLineMinutesByTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) (map[int64]int64, error) {
	if r.workDayLineStore == nil {
		return nil, domain.ErrWorkDayLineStoreUnavailable
	}
	return r.workDayLineStore.SumWorkDayLineMinutesByTaskIDs(ctx, businessID, taskIDs)
}

func (r *Repository) CreateWorkDayLine(ctx context.Context, input domain.WorkDayLineInput) (domain.WorkDayLine, error) {
	if r.workDayLineStore == nil {
		return domain.WorkDayLine{}, domain.ErrWorkDayLineStoreUnavailable
	}
	return r.workDayLineStore.CreateWorkDayLine(ctx, input)
}

func (r *Repository) UpdateWorkDayLine(ctx context.Context, input domain.WorkDayLineInput) (domain.WorkDayLine, error) {
	if r.workDayLineStore == nil {
		return domain.WorkDayLine{}, domain.ErrWorkDayLineStoreUnavailable
	}
	return r.workDayLineStore.UpdateWorkDayLine(ctx, input)
}

func (r *Repository) GetWorkDayLineOwner(ctx context.Context, id, businessID int64) (domain.WorkDayLineOwner, bool, error) {
	if r.workDayLineStore == nil {
		return domain.WorkDayLineOwner{}, false, domain.ErrWorkDayLineStoreUnavailable
	}
	return r.workDayLineStore.GetWorkDayLineOwner(ctx, id, businessID)
}

func (r *Repository) DeleteWorkDayLine(ctx context.Context, id, businessID int64) (bool, error) {
	if r.workDayLineStore == nil {
		return false, domain.ErrWorkDayLineStoreUnavailable
	}
	return r.workDayLineStore.DeleteWorkDayLine(ctx, id, businessID)
}

// GetWorkDayByID reads a day by id, scoped to the tenant.
func (r *Repository) GetWorkDayByID(ctx context.Context, id, businessID int64) (domain.WorkDay, bool, error) {
	row, err := r.queries.GetWorkDayByID(ctx, sqlc.GetWorkDayByIDParams{
		ID:         id,
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
func (r *Repository) ListWorkTypes(ctx context.Context, businessID int64, includeArchived bool) ([]domain.WorkType, error) {
	rows, err := r.queries.ListWorkTypes(ctx, sqlc.ListWorkTypesParams{
		BusinessID: businessID,
		Column2:    includeArchived,
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
	return r.GetWorkType(ctx, id, input.BusinessID)
}

// GetWorkType reads one node, scoped to the tenant.
func (r *Repository) GetWorkType(ctx context.Context, id, businessID int64) (domain.WorkType, error) {
	row, err := r.queries.GetWorkType(ctx, sqlc.GetWorkTypeParams{
		ID:         id,
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
		BusinessID:   input.BusinessID,
	}); err != nil {
		logger.Ctx(ctx).Error("UpdateWorkType failed", logger.Error(err))
		return domain.WorkType{}, err
	}
	return r.GetWorkType(ctx, input.ID, input.BusinessID)
}
