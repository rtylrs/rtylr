package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// dynamic-query: squirrel
func (r *Repository) UpdateTask(ctx context.Context, input domain.UpdateTaskInput) error {
	var statusParam interface{}
	if input.Status != "" {
		statusParam = input.Status
	}
	var priorityParam interface{}
	if input.Priority != "" {
		priorityParam = input.Priority
	}
	// Guard on the input slice, not on the marshalled bytes. marshalLabels turns
	// a nil slice into "[]", which is never nil, so testing its output made
	// labelsParam non-NULL on every call and overwrote labels on updates that
	// never mentioned them. Only a nil slice means "leave the column alone"; an
	// explicit empty slice means "clear every label" and must still be written.
	// marshalLabels is left as-is because CreateTask depends on it defaulting a
	// nil slice to "[]" for the INSERT.
	var labelsParam interface{}
	if input.Labels != nil {
		labelsJSON, err := marshalLabels(input.Labels)
		if err != nil {
			return err
		}
		labelsParam = labelsJSON
	}
	var titleParam interface{}
	if input.Title != "" {
		titleParam = input.Title
	}
	var assigneeParam interface{}
	if input.AssigneeID != nil {
		assigneeParam = *input.AssigneeID
	}
	var dueDateParam interface{}
	if input.DueDate != nil {
		dueDateParam = *input.DueDate
	}
	var startDateParam interface{}
	if input.StartDate != nil {
		startDateParam = *input.StartDate
	}
	var archivedParam interface{}
	if input.Archived != nil {
		archivedParam = *input.Archived
	}

	// Conditional setters distinguish an explicit clear from an omitted field.
	builder := sq.Update("flow_tasks")
	if input.SectionID != nil {
		if *input.SectionID == 0 {
			builder = builder.Set("section_id", nil)
		} else {
			builder = builder.Set("section_id", *input.SectionID)
		}
	}
	if input.EffortMinutes != nil {
		if *input.EffortMinutes == 0 {
			builder = builder.Set("effort_minutes", nil)
		} else {
			builder = builder.Set("effort_minutes", *input.EffortMinutes)
		}
	}
	if input.EstimatedHours != nil {
		if *input.EstimatedHours == 0 {
			builder = builder.Set("estimated_hours", nil)
		} else {
			// Format identically to CreateTask, deliberately: estimated_hours is a
			// DECIMAL(6,2) column, and leaving this as a raw float64 would give the
			// column two different marshalling paths that could drift from each
			// other depending on the driver's float formatting.
			builder = builder.Set("estimated_hours", strconv.FormatFloat(*input.EstimatedHours, 'f', 2, 64))
		}
	}
	if input.WaitingOn != nil {
		if *input.WaitingOn == "" {
			builder = builder.Set("waiting_on", nil)
		} else {
			builder = builder.Set("waiting_on", *input.WaitingOn)
		}
	}
	if input.DecisionNote != nil {
		if *input.DecisionNote == "" {
			builder = builder.Set("decision_note", nil)
		} else {
			builder = builder.Set("decision_note", *input.DecisionNote)
		}
	}

	// Partial updates (the client sends only the fields it changed, e.g. just
	// priority or assignee) must not overwrite the fields it omitted. COALESCE(?, col)
	// preserves the existing value when the param is NULL, and every nullable param
	// above is left NULL when the request omits that field. The completed_at CASE only
	// fires when a non-NULL status is explicitly supplied.
	if input.ClearDescription {
		builder = builder.Set("description", "")
	} else {
		builder = builder.Set("description", sq.Expr("COALESCE(?, description)", sql.NullString{String: input.Description, Valid: input.Description != ""}))
	}
	if input.AssigneeID != nil && *input.AssigneeID == 0 {
		builder = builder.Set("assignee_id", nil)
	} else {
		builder = builder.Set("assignee_id", sq.Expr("COALESCE(?, assignee_id)", assigneeParam))
	}
	if input.ClearDueDate {
		builder = builder.Set("due_date", nil)
	} else {
		builder = builder.Set("due_date", sq.Expr("COALESCE(?, due_date)", dueDateParam))
	}
	if input.ClearStartDate {
		builder = builder.Set("start_date", nil)
	} else {
		builder = builder.Set("start_date", sq.Expr("COALESCE(?, start_date)", startDateParam))
	}
	q, args, err := builder.
		Set("title", sq.Expr("COALESCE(?, title)", titleParam)).
		Set("status", sq.Expr("COALESCE(?, status)", statusParam)).
		Set("priority", sq.Expr("COALESCE(?, priority)", priorityParam)).
		Set("labels", sq.Expr("COALESCE(?, labels)", labelsParam)).
		Set("archived", sq.Expr("COALESCE(?, archived)", archivedParam)).
		Set("completed_at", sq.Expr(
			`CASE
		        WHEN ? IS NOT NULL AND ? = 'done'  AND completed_at IS NULL THEN NOW()
		        WHEN ? IS NOT NULL AND ? != 'done' THEN NULL
		        ELSE completed_at
		    END`,
			statusParam, statusParam, // CASE WHEN ? IS NOT NULL AND ? = 'done'
			statusParam, statusParam, // CASE WHEN ? IS NOT NULL AND ? != 'done'
		)).
		Where(sq.Eq{"id": input.ID, "business_id": input.BusinessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("UpdateTask build failed", logger.Error(err))
		return err
	}

	_, err = r.pool.ExecContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateTask failed", logger.Error(err))
		return err
	}
	return nil
}

// GetFocusPins returns the caller's pins for the supplied tasks. The task
// business predicate is repeated on the join so a stale pin row can never turn
// into a response for a different tenant.
func (r *Repository) GetFocusPins(ctx context.Context, userID, businessID int64, taskIDs []int64) (map[int64]bool, error) {
	pins := make(map[int64]bool, len(taskIDs))
	if len(taskIDs) == 0 {
		return pins, nil
	}

	q, args, err := sq.
		Select("fp.task_id").
		From("flow_task_focus_pins fp").
		Join("flow_tasks t ON t.id = fp.task_id").
		Where(sq.Eq{
			"fp.user_id":     userID,
			"fp.business_id": businessID,
			"t.business_id":  businessID,
			"fp.task_id":     taskIDs,
		}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var taskID int64
		if err := rows.Scan(&taskID); err != nil {
			return nil, err
		}
		pins[taskID] = true
	}
	return pins, rows.Err()
}

// SetTaskFocusPin pins or unpins a task in one user's Today view. The pin is
// per-user by design: Today is a personal view, so a column on flow_tasks would
// put one member's planning in another member's list.
//
// The pin row is still guarded by the task's business on insert; callers do
// not get to attach a pin to a task outside the resolved tenant.
func (r *Repository) SetTaskFocusPin(ctx context.Context, taskID, userID, businessID int64, pinned bool) error {
	if !pinned {
		_, err := r.pool.ExecContext(ctx,
			"DELETE FROM flow_task_focus_pins WHERE task_id = ? AND user_id = ? AND business_id = ?",
			taskID, userID, businessID,
		)
		if err != nil {
			logger.Ctx(ctx).Error("SetTaskFocusPin delete failed", logger.Error(err))
		}
		return err
	}
	// INSERT ... SELECT keeps the business boundary in the write itself. The
	// service also checks the task before calling this, but a repository method
	// must not rely solely on that caller convention.
	_, err := r.pool.ExecContext(ctx,
		`INSERT IGNORE INTO flow_task_focus_pins (task_id, user_id, business_id)
		 SELECT ?, ?, ? FROM flow_tasks WHERE id = ? AND business_id = ?`,
		taskID, userID, businessID, taskID, businessID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("SetTaskFocusPin insert failed", logger.Error(err))
	}
	return err
}

// AddTaskDependency records that taskID is blocked by dependsOnTaskID. The
// unique key on (task_id, depends_on_task_id) makes a repeat add a no-op.
//
// input.BusinessID is written as a plain column value here, not a WHERE
// guard, so this method trusts the caller to have already confirmed both
// tasks belong to that business (the same trust CreateTask already places in
// caller-supplied ids). The API handler enforces that via RequireTaskRole
// before calling in.
func (r *Repository) AddTaskDependency(ctx context.Context, input domain.AddTaskDependencyInput) error {
	_, err := r.pool.ExecContext(ctx,
		"INSERT IGNORE INTO flow_task_dependencies (task_id, depends_on_task_id, business_id) VALUES (?, ?, ?)",
		input.TaskID, input.DependsOnTaskID, input.BusinessID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("AddTaskDependency failed", logger.Error(err))
	}
	return err
}

func (r *Repository) RemoveTaskDependency(ctx context.Context, input domain.RemoveTaskDependencyInput) error {
	_, err := r.pool.ExecContext(ctx,
		"DELETE FROM flow_task_dependencies WHERE task_id = ? AND depends_on_task_id = ? AND business_id = ?",
		input.TaskID, input.DependsOnTaskID, input.BusinessID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("RemoveTaskDependency failed", logger.Error(err))
	}
	return err
}

// ListTaskDependencies returns dependency ids keyed by task id. It takes a slice
// so list views resolve every row in one query instead of N.
func (r *Repository) ListTaskDependencies(ctx context.Context, taskIDs []int64, businessID int64) (map[int64][]int64, error) {
	out := make(map[int64][]int64, len(taskIDs))
	if len(taskIDs) == 0 {
		return out, nil
	}
	q, args, err := sq.
		Select("task_id", "depends_on_task_id").
		From("flow_task_dependencies").
		Where(sq.Eq{"task_id": taskIDs, "business_id": businessID}).
		OrderBy("id").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("ListTaskDependencies failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var taskID, dependsOn int64
		if err := rows.Scan(&taskID, &dependsOn); err != nil {
			return nil, err
		}
		out[taskID] = append(out[taskID], dependsOn)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateTaskStatus(ctx context.Context, input domain.UpdateTaskStatusInput) error {
	err := r.queries.UpdateTaskStatus(ctx, sqlc.UpdateTaskStatusParams{
		Status:     sqlc.NullFlowTasksStatus{FlowTasksStatus: sqlc.FlowTasksStatus(input.Status), Valid: true},
		Column2:    input.Status,
		Column3:    input.Status,
		ID:         input.ID,
		BusinessID: input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateTaskStatus failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteTask(ctx context.Context, taskID, businessID int64) error {
	err := r.queries.DeleteTask(ctx, sqlc.DeleteTaskParams{
		ID:         taskID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteTask failed", logger.Error(err))
		return err
	}
	return nil
}

// ─────────────────────────────────────────────
// Dashboard
// ─────────────────────────────────────────────
