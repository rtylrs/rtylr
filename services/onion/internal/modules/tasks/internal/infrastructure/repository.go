package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/tasks/internal/domain"
	"github.com/voxire/onion/internal/modules/tasks/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Task, error) {
	rows, err := r.queries.ListTasksByOwner(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list tasks", logger.Error(err))
		return nil, err
	}

	tasks := make([]domain.Task, 0, len(rows))
	for _, row := range rows {
		t := domain.Task{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			Title:       row.Title,
			Description: row.Description.String,
			Status:      string(row.Status.FlowTasksStatus),
			Priority:    string(row.Priority.FlowTasksPriority),
			LabelIDs:    row.Labels,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		}
		if row.ProjectID != 0 {
			t.ProjectID = &row.ProjectID
		}
		if row.WorkspaceID != 0 {
			t.WorkspaceID = &row.WorkspaceID
		}
		if row.AssigneeID.Valid {
			t.AssigneeID = &row.AssigneeID.Int64
		}
		if row.DueDate.Valid {
			tt := row.DueDate.Time
			t.DueDate = &tt
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *Repository) Create(ctx context.Context, input domain.TaskInput) (int64, error) {
	labelIDs := json.RawMessage(input.LabelIDs)
	if labelIDs == nil {
		labelIDs = json.RawMessage("[]")
	}
	priority := input.Priority
	if priority == "" {
		priority = "medium"
	}

	var dueDate sql.NullTime
	if input.DueDate != nil {
		dueDate = sql.NullTime{Time: *input.DueDate, Valid: true}
	}

	result, err := r.queries.CreateTask(ctx, sqlc.CreateTaskParams{
		OwnerID:     input.OwnerID,
		ProjectID:   derefInt64(input.ProjectID),
		WorkspaceID: derefInt64(input.WorkspaceID),
		Title:       input.Title,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Priority:    sqlc.NullFlowTasksPriority{FlowTasksPriority: sqlc.FlowTasksPriority(priority), Valid: true},
		AssigneeID:  nullInt64Ptr(input.AssigneeID),
		Labels:      labelIDs,
		DueDate:     dueDate,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create task", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.TaskUpdateInput) error {
	labelIDs := json.RawMessage(input.LabelIDs)
	if labelIDs == nil {
		labelIDs = json.RawMessage("[]")
	}

	var dueDate sql.NullTime
	if input.DueDate != nil {
		dueDate = sql.NullTime{Time: *input.DueDate, Valid: true}
	}

	result, err := r.queries.UpdateTask(ctx, sqlc.UpdateTaskParams{
		ProjectID:   derefInt64(input.ProjectID),
		WorkspaceID: derefInt64(input.WorkspaceID),
		Title:       input.Title,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Status:      sqlc.NullFlowTasksStatus{FlowTasksStatus: sqlc.FlowTasksStatus(input.Status), Valid: input.Status != ""},
		Priority:    sqlc.NullFlowTasksPriority{FlowTasksPriority: sqlc.FlowTasksPriority(input.Priority), Valid: input.Priority != ""},
		AssigneeID:  nullInt64Ptr(input.AssigneeID),
		Labels:      labelIDs,
		DueDate:     dueDate,
		ID:          input.ID,
		OwnerID:     input.OwnerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update task", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	result, err := r.queries.DeleteTask(ctx, sqlc.DeleteTaskParams{
		ID:      id,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete task", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}

func (r *Repository) ListLabels(ctx context.Context, ownerID int64) ([]domain.Label, error) {
	rows, err := r.queries.ListLabelsByOwner(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list labels", logger.Error(err))
		return nil, err
	}

	labels := make([]domain.Label, 0, len(rows))
	for _, row := range rows {
		labels = append(labels, domain.Label{
			ID:        row.ID,
			OwnerID:   row.WorkspaceID,
			Name:      row.Name,
			Color:     row.Color,
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return labels, nil
}

func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func nullInt64Ptr(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func (r *Repository) CreateLabel(ctx context.Context, input domain.LabelInput) (int64, error) {
	result, err := r.queries.CreateLabel(ctx, sqlc.CreateLabelParams{
		WorkspaceID: input.OwnerID,
		Name:        input.Name,
		Color:       sql.NullString{String: input.Color, Valid: input.Color != ""}.String,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create label", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}
