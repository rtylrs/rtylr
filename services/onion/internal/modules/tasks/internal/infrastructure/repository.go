package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/tasks/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Task, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, project_id, workspace_id, title, description, status, priority, assignee_id, label_ids, due_date, created_at, updated_at
		 FROM flow_tasks WHERE owner_id = ? ORDER BY created_at DESC`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list tasks", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	tasks := make([]domain.Task, 0)
	for rows.Next() {
		t, scanErr := scanTask(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *Repository) Create(ctx context.Context, input domain.TaskInput) (int64, error) {
	labelIDs := input.LabelIDs
	if labelIDs == nil {
		labelIDs = []byte("[]")
	}
	priority := input.Priority
	if priority == "" {
		priority = "medium"
	}
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_tasks (owner_id, project_id, workspace_id, title, description, status, priority, assignee_id, label_ids, due_date)
		 VALUES (?, ?, ?, ?, ?, 'todo', ?, ?, ?, ?)`,
		input.OwnerID, input.ProjectID, input.WorkspaceID, input.Title,
		nullableString(input.Description), priority, input.AssigneeID, labelIDs,
		nullableTimePtrToNull(input.DueDate),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create task", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.TaskUpdateInput) error {
	labelIDs := input.LabelIDs
	if labelIDs == nil {
		labelIDs = []byte("[]")
	}
	result, err := r.pool.ExecContext(ctx,
		`UPDATE flow_tasks SET project_id = ?, workspace_id = ?, title = ?, description = ?, status = ?, priority = ?, assignee_id = ?, label_ids = ?, due_date = ?
		 WHERE id = ? AND owner_id = ?`,
		input.ProjectID, input.WorkspaceID, input.Title, nullableString(input.Description),
		input.Status, input.Priority, input.AssigneeID, labelIDs,
		nullableTimePtrToNull(input.DueDate), input.ID, input.OwnerID,
	)
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
	result, err := r.pool.ExecContext(ctx,
		"DELETE FROM flow_tasks WHERE id = ? AND owner_id = ?", id, ownerID,
	)
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
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, name, color, created_at FROM flow_labels WHERE owner_id = ? ORDER BY name`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list labels", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	labels := make([]domain.Label, 0)
	for rows.Next() {
		var l domain.Label
		var color sql.NullString
		if scanErr := rows.Scan(&l.ID, &l.OwnerID, &l.Name, &color, &l.CreatedAt); scanErr != nil {
			return nil, scanErr
		}
		l.Color = color.String
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

func (r *Repository) CreateLabel(ctx context.Context, input domain.LabelInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_labels (owner_id, name, color) VALUES (?, ?, ?)`,
		input.OwnerID, input.Name, nullableString(input.Color),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create label", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func scanTask(rows *sql.Rows) (domain.Task, error) {
	var t domain.Task
	var projID, wsID, assigneeID sql.NullInt64
	var desc sql.NullString
	var labelIDs []byte
	var dueDate sql.NullTime

	if err := rows.Scan(&t.ID, &t.OwnerID, &projID, &wsID, &t.Title, &desc, &t.Status, &t.Priority, &assigneeID, &labelIDs, &dueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return domain.Task{}, err
	}
	if projID.Valid {
		t.ProjectID = &projID.Int64
	}
	if wsID.Valid {
		t.WorkspaceID = &wsID.Int64
	}
	if assigneeID.Valid {
		t.AssigneeID = &assigneeID.Int64
	}
	t.Description = desc.String
	t.LabelIDs = labelIDs
	if dueDate.Valid {
		tt := dueDate.Time
		t.DueDate = &tt
	}
	return t, nil
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullableTimePtrToNull(t interface{}) interface{} {
	// Return nil to let DB use NULL for nil *time.Time
	if t == nil {
		return nil
	}
	return t
}
