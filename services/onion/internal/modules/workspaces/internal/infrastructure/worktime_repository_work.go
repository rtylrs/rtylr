package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

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

// GetWorkDayLineOwner reads the employee that owns a line, scoped to the
// current tenant. The application uses this before line mutations so a caller
// cannot mutate a colleague's work day by presenting only a line id.
func (r *Repository) GetWorkDayLineOwner(ctx context.Context, id, ownerID, businessID int64) (int64, int64, bool, error) {
	row, err := r.queries.GetWorkDayLineOwner(ctx, sqlc.GetWorkDayLineOwnerParams{
		ID:         id,
		OwnerID:    ownerID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, false, nil
		}
		logger.Ctx(ctx).Error("GetWorkDayLineOwner failed", logger.Error(err))
		return 0, 0, false, err
	}
	return int64Value(row.WorkDayID), row.EmployeeID, true, nil
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
