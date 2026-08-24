package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// F-A: per-project custom statuses.
//
// Every method passes businessID into the query. The sqlc params make that hard
// to forget, but not impossible: onion#929 shipped a write that built its input
// without BusinessID, and because the column defaults to 0 on the older tables
// the row was stored and then never matched a read. flow_statuses has no such
// default, so the same mistake here fails loudly.

func statusFromRow(id, businessID, projectID int64, name string, bucket sqlc.FlowStatusesBucket,
	legacy sqlc.FlowStatusesLegacyStatus, color string, sortOrder int32, isDefault bool,
	createdBy int64, createdAt sql.NullTime) domain.Status {
	s := domain.Status{
		ID:           id,
		BusinessID:   businessID,
		ProjectID:    projectID,
		Name:         name,
		Bucket:       string(bucket),
		LegacyStatus: string(legacy),
		Color:        color,
		SortOrder:    sortOrder,
		IsDefault:    isDefault,
		CreatedBy:    createdBy,
	}
	if createdAt.Valid {
		s.CreatedAt = createdAt.Time.UTC().Format("2006-01-02T15:04:05Z")
	}
	return s
}

// ListStatusesByProject returns a project's set in sort order.
//
// An empty slice is the expected result for a project that has not configured
// any, and means "fall back to the built-in string statuses". It is not an
// error and must never be reported as one.
func (r *Repository) ListStatusesByProject(ctx context.Context, businessID, projectID int64) ([]domain.Status, error) {
	rows, err := r.queries.ListStatusesByProject(ctx, sqlc.ListStatusesByProjectParams{
		BusinessID: businessID,
		ProjectID:  projectID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListStatusesByProject failed", logger.Error(err))
		return nil, err
	}
	out := make([]domain.Status, 0, len(rows))
	for _, row := range rows {
		out = append(out, statusFromRow(row.ID, row.BusinessID, row.ProjectID, row.Name,
			row.Bucket, row.LegacyStatus, row.Color, row.SortOrder, row.IsDefault,
			row.CreatedBy, row.CreatedAt))
	}
	return out, nil
}

// GetStatusByID returns one status, scoped to the caller's business.
func (r *Repository) GetStatusByID(ctx context.Context, id, businessID int64) (domain.Status, error) {
	row, err := r.queries.GetStatusByID(ctx, sqlc.GetStatusByIDParams{ID: id, BusinessID: businessID})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Status{}, domain.ErrStatusNotFound
	}
	if err != nil {
		logger.Ctx(ctx).Error("GetStatusByID failed", logger.Error(err))
		return domain.Status{}, err
	}
	return statusFromRow(row.ID, row.BusinessID, row.ProjectID, row.Name, row.Bucket,
		row.LegacyStatus, row.Color, row.SortOrder, row.IsDefault, row.CreatedBy, row.CreatedAt), nil
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

// CreateStatus inserts a status and returns it.
//
// Runs in a transaction because is_default is at most one per project and MySQL
// cannot express that as a partial unique index: clearing the flag and setting
// it have to be one atomic step, or two concurrent creates leave a project with
// two defaults and a new task landing on whichever the read happens to find.
func (r *Repository) CreateStatus(ctx context.Context, businessID, createdBy int64, in domain.StatusInput) (domain.Status, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return domain.Status{}, err
	}
	defer func() { _ = tx.Rollback() }()
	q := r.queries.WithTx(tx)

	sortOrder := in.SortOrder
	if !in.SortOrderSet {
		next, nextErr := q.NextStatusSortOrder(ctx, sqlc.NextStatusSortOrderParams{
			BusinessID: businessID, ProjectID: in.ProjectID,
		})
		if nextErr != nil {
			logger.Ctx(ctx).Error("NextStatusSortOrder failed", logger.Error(nextErr))
			return domain.Status{}, nextErr
		}
		sortOrder = int32(next)
	}

	if in.IsDefault {
		if _, clearErr := q.ClearProjectDefaultStatus(ctx, sqlc.ClearProjectDefaultStatusParams{
			BusinessID: businessID, ProjectID: in.ProjectID,
		}); clearErr != nil {
			logger.Ctx(ctx).Error("ClearProjectDefaultStatus failed", logger.Error(clearErr))
			return domain.Status{}, clearErr
		}
	}

	res, err := q.CreateStatus(ctx, sqlc.CreateStatusParams{
		BusinessID:   businessID,
		ProjectID:    in.ProjectID,
		Name:         in.Name,
		Bucket:       sqlc.FlowStatusesBucket(in.Bucket),
		LegacyStatus: sqlc.FlowStatusesLegacyStatus(in.LegacyStatus),
		Color:        in.Color,
		SortOrder:    sortOrder,
		IsDefault:    in.IsDefault,
		CreatedBy:    createdBy,
	})
	if err != nil {
		if isDuplicateKey(err) {
			return domain.Status{}, domain.ErrStatusNameTaken
		}
		logger.Ctx(ctx).Error("CreateStatus failed", logger.Error(err))
		return domain.Status{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Status{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Status{}, err
	}

	return r.GetStatusByID(ctx, id, businessID)
}

// UpdateStatus replaces a status. Same transaction reasoning as CreateStatus.
func (r *Repository) UpdateStatus(ctx context.Context, id, businessID, projectID int64, in domain.StatusInput) (domain.Status, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return domain.Status{}, err
	}
	defer func() { _ = tx.Rollback() }()
	q := r.queries.WithTx(tx)

	if in.IsDefault {
		if _, clearErr := q.ClearProjectDefaultStatus(ctx, sqlc.ClearProjectDefaultStatusParams{
			BusinessID: businessID, ProjectID: projectID,
		}); clearErr != nil {
			logger.Ctx(ctx).Error("ClearProjectDefaultStatus failed", logger.Error(clearErr))
			return domain.Status{}, clearErr
		}
	}

	res, err := q.UpdateStatus(ctx, sqlc.UpdateStatusParams{
		Name:         in.Name,
		Bucket:       sqlc.FlowStatusesBucket(in.Bucket),
		LegacyStatus: sqlc.FlowStatusesLegacyStatus(in.LegacyStatus),
		Color:        in.Color,
		SortOrder:    in.SortOrder,
		IsDefault:    in.IsDefault,
		ID:           id,
		BusinessID:   businessID,
	})
	if err != nil {
		if isDuplicateKey(err) {
			return domain.Status{}, domain.ErrStatusNameTaken
		}
		logger.Ctx(ctx).Error("UpdateStatus failed", logger.Error(err))
		return domain.Status{}, err
	}
	// A zero-row update means the id does not belong to this business. Reported
	// as not-found rather than success, because returning success here is how a
	// caller concludes a write landed when it never did.
	affected, err := res.RowsAffected()
	if err != nil {
		return domain.Status{}, err
	}
	if affected == 0 {
		return domain.Status{}, domain.ErrStatusNotFound
	}
	if err := tx.Commit(); err != nil {
		return domain.Status{}, err
	}
	return r.GetStatusByID(ctx, id, businessID)
}

// DeleteStatus removes a status and moves its tasks back to their built-in
// string status. Returns how many tasks moved.
//
// Never blocks on the tasks and never deletes them: the string status they
// already carry keeps them exactly where the user left them.
func (r *Repository) DeleteStatus(ctx context.Context, id, businessID int64) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	q := r.queries.WithTx(tx)

	cleared, err := q.ClearTasksOfStatus(ctx, sqlc.ClearTasksOfStatusParams{
		BusinessID: businessID, StatusID: sql.NullInt64{Int64: id, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("ClearTasksOfStatus failed", logger.Error(err))
		return 0, err
	}
	moved, err := cleared.RowsAffected()
	if err != nil {
		return 0, err
	}

	res, err := q.DeleteStatus(ctx, sqlc.DeleteStatusParams{ID: id, BusinessID: businessID})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteStatus failed", logger.Error(err))
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected == 0 {
		return 0, domain.ErrStatusNotFound
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return moved, nil
}

// SetTaskStatusID applies a custom status to a task, or clears it when statusID
// is zero, and writes the matching built-in string status in the same statement.
//
// legacyStatus is resolved by the caller from the status row, so the two columns
// cannot drift. Letting them drift is how a task shows one status on the board
// and another in a report.
func (r *Repository) SetTaskStatusID(ctx context.Context, taskID, businessID, statusID int64, legacyStatus string) error {
	params := sqlc.SetTaskStatusIDParams{
		StatusID:   sql.NullInt64{Int64: statusID, Valid: statusID != 0},
		Status:     sqlc.NullFlowTasksStatus{FlowTasksStatus: sqlc.FlowTasksStatus(legacyStatus), Valid: legacyStatus != ""},
		ID:         taskID,
		BusinessID: businessID,
	}
	res, err := r.queries.SetTaskStatusID(ctx, params)
	if err != nil {
		logger.Ctx(ctx).Error("SetTaskStatusID failed", logger.Error(err))
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrStatusNotFound
	}
	return nil
}
