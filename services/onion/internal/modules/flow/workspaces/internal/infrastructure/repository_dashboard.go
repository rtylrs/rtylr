package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application/identity"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/times"
)

func (r *Repository) GetDashboardStats(ctx context.Context, workspaceID, businessID int64) (domain.DashboardStats, error) {
	var stats domain.DashboardStats

	weekStart, _ := times.GetWeekRangeStartAndEnd()
	weekEnd := weekStart.AddDate(0, 0, 7)

	row, err := r.queries.GetDashboardStats(ctx, sqlc.GetDashboardStatsParams{
		CompletedAt:   sql.NullTime{Time: weekStart, Valid: true},
		CompletedAt_2: sql.NullTime{Time: weekEnd, Valid: true},
		WorkspaceID:   workspaceID,
		BusinessID:    businessID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("GetDashboardStats failed", logger.Error(err))
		return domain.DashboardStats{}, err
	}

	stats.OpenTasks = row.OpenTasks
	stats.InProgressTasks = row.InProgressTasks
	stats.CompletedThisWeek = row.CompletedThisWeek
	stats.OverdueTasks = row.OverdueTasks

	return stats, nil
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

func nullInt64Ptr(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func nullTimePtrToSql(v *time.Time) sql.NullTime {
	if v == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *v, Valid: true}
}

func marshalLabels(labels []int64) (json.RawMessage, error) {
	if labels == nil {
		return json.RawMessage("[]"), nil
	}
	return json.Marshal(labels)
}

// enrichedTaskColumns is the column list used by every task read path.
// It selects the raw flow_tasks row plus the JOINed display projections
// (project name/color and subtask counts). Assignee identities are read in a batch.
//
// Keep this list, scanEnrichedTask, and the JOINs in
// enrichedTaskSelect in sync.
func enrichedTaskColumns(includeDescription bool) []string {
	descriptionColumn := "t.description"
	if !includeDescription {
		descriptionColumn = "'' AS description"
	}

	return []string{
		"t.id", "t.project_id", "t.workspace_id", "t.owner_id", "t.business_id", "t.parent_task_id", "t.section_id",
		"t.title", descriptionColumn, "COALESCE(t.status, 'todo') AS status", "COALESCE(t.priority, 'none') AS priority",
		"t.assignee_id", "t.reporter_id", "t.due_date", "t.start_date", "t.completed_at",
		"t.estimated_hours", "t.actual_hours", "t.effort_minutes", "t.sort_order", "t.labels",
		"COALESCE(t.archived, FALSE) AS archived",
		"t.created_at", "t.updated_at",
		"COALESCE(p.name, '')  AS project_name",
		"COALESCE(p.color, '') AS project_color",
		"'' AS assignee_name",
		"COALESCE((SELECT COUNT(*) FROM flow_tasks st WHERE st.parent_task_id = t.id), 0) AS subtask_count",
		"(SELECT COUNT(*) FROM flow_tasks st WHERE st.parent_task_id = t.id AND st.status = 'done') AS completed_subtasks",
		"(SELECT COUNT(*) FROM flow_task_comments c WHERE c.task_id = t.id) AS comment_count",
		"(SELECT COUNT(*) FROM flow_task_attachments a WHERE a.task_id = t.id) AS attachment_count",
		"t.status_id", "COALESCE(t.waiting_on, '')", "COALESCE(t.decision_note, '')",
		"COALESCE(fs.name, '') AS status_name",
		"COALESCE(fs.color, '') AS status_color",
	}
}

// enrichedTaskSelect returns a squirrel builder with the enriched task
// SELECT + JOINs preconfigured. Callers add WHERE/ORDER/LIMIT clauses.
// All tenant-scope filtering is the caller's responsibility.
//
// dynamic-query: squirrel
func enrichedTaskSelect() sq.SelectBuilder {
	return enrichedTaskSelectForList(false)
}

// dynamic-query: squirrel
func enrichedTaskSelectForList(compact bool) sq.SelectBuilder {
	return sq.
		Select(enrichedTaskColumns(!compact)...).
		From("flow_tasks t").
		LeftJoin("flow_projects p ON p.id = t.project_id").
		LeftJoin("flow_statuses fs ON fs.id = t.status_id").
		PlaceholderFormat(sq.Question)
}

// listEnrichedTasks runs an enriched task query built by the caller.
// dynamic-query: squirrel
func (r *Repository) listEnrichedTasks(ctx context.Context, builder sq.SelectBuilder) ([]domain.Task, error) {
	q, args, err := builder.ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("listEnrichedTasks build failed", logger.Error(err))
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			// Request cancellation is expected when a client navigates away or
			// disconnects; it is not an application/database failure.
			logger.Ctx(ctx).Debug("listEnrichedTasks canceled", logger.Error(err))
		} else {
			logger.Ctx(ctx).Error("listEnrichedTasks failed", logger.Error(err))
		}
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("listEnrichedTasks row close failed", logger.Error(closeErr))
		}
	}()

	var tasks []domain.Task
	for rows.Next() {
		t, scanErr := scanEnrichedTask(ctx, rows)
		if scanErr != nil {
			return nil, scanErr
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	keys := make([]identity.IdentityKey, 0, len(tasks))
	for _, task := range tasks {
		if task.AssigneeID != nil {
			keys = append(keys, identity.IdentityKey{BusinessID: task.BusinessID, UserID: *task.AssigneeID})
		}
	}
	profiles, err := r.readIdentities(ctx, keys)
	if err != nil {
		return nil, err
	}
	for i := range tasks {
		if tasks[i].AssigneeID != nil {
			tasks[i].AssigneeName = profiles[identity.IdentityKey{BusinessID: tasks[i].BusinessID, UserID: *tasks[i].AssigneeID}].Name
			tasks[i].AssigneeInitials = initialsOf(tasks[i].AssigneeName)
		}
	}
	return tasks, nil
}

// scanEnrichedTask scans one row from a query using enrichedTaskColumns.
// Returns a fully-populated domain.Task including display projections.
func scanEnrichedTask(ctx context.Context, rows *sql.Rows) (domain.Task, error) {
	var t domain.Task
	var desc sql.NullString
	var parentID, sectionID, assigneeID, statusID sql.NullInt64
	var dueDate, startDate, completedAt sql.NullTime
	var estHours, actHours sql.NullFloat64
	var effortMinutes sql.NullInt64
	var labelsRaw []byte

	err := rows.Scan(
		&t.ID, &t.ProjectID, &t.WorkspaceID, &t.OwnerID, &t.BusinessID, &parentID, &sectionID,
		&t.Title, &desc, &t.Status, &t.Priority,
		&assigneeID, &t.ReporterID,
		&dueDate, &startDate, &completedAt,
		&estHours, &actHours, &effortMinutes, &t.SortOrder, &labelsRaw, &t.Archived,
		&t.CreatedAt, &t.UpdatedAt,
		&t.ProjectName, &t.ProjectColor, &t.AssigneeName, &t.SubtaskCount,
		&t.CompletedSubtasks, &t.CommentCount, &t.AttachmentCount,
		&statusID, &t.WaitingOn, &t.DecisionNote, &t.StatusName, &t.StatusColor,
	)
	if err != nil {
		logger.Ctx(ctx).Error("scanEnrichedTask row scan failed", logger.Error(err))
		return domain.Task{}, err
	}

	t.Description = desc.String
	if parentID.Valid {
		t.ParentTaskID = &parentID.Int64
	}
	if sectionID.Valid {
		t.SectionID = &sectionID.Int64
	}
	if assigneeID.Valid {
		t.AssigneeID = &assigneeID.Int64
	}
	if statusID.Valid {
		t.StatusID = &statusID.Int64
	}
	if dueDate.Valid {
		t.DueDate = &dueDate.Time
	}
	if startDate.Valid {
		t.StartDate = &startDate.Time
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	if estHours.Valid {
		t.EstimatedHours = &estHours.Float64
	}
	if actHours.Valid {
		t.ActualHours = &actHours.Float64
	}
	if effortMinutes.Valid {
		value := int32(effortMinutes.Int64) // #nosec G115 -- effort_minutes is an INT column.
		t.EffortMinutes = &value
	}
	if len(labelsRaw) > 0 {
		if unErr := json.Unmarshal(labelsRaw, &t.Labels); unErr != nil {
			logger.Ctx(ctx).Warn("failed to unmarshal task labels", logger.Error(unErr))
		}
	}
	t.AssigneeInitials = initialsOf(t.AssigneeName)
	if t.Identifier == "" {
		t.Identifier = "T-" + strconv.FormatInt(t.ID, 10)
	}
	return t, nil
}

// initialsOf returns uppercase initials for a display name (up to 2 chars).
func initialsOf(name string) string {
	if name == "" {
		return ""
	}
	out := make([]rune, 0, 2)
	prev := ' '
	for _, r := range name {
		if (prev == ' ' || prev == '\t') && r != ' ' && r != '\t' {
			if r >= 'a' && r <= 'z' {
				r -= 32
			}
			out = append(out, r)
			if len(out) == 2 {
				break
			}
		}
		prev = r
	}
	return string(out)
}

// scanTasksRaw — legacy bare-row scanner kept for code paths not yet
// migrated to listEnrichedTasks (e.g. direct-write callbacks returning
// the inserted row). New callers should use listEnrichedTasks + scanEnrichedTask.
func scanTasksRaw(ctx context.Context, rows *sql.Rows) ([]domain.Task, error) {
	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		var desc sql.NullString
		var parentID, sectionID, assigneeID sql.NullInt64
		var dueDate, startDate, completedAt sql.NullTime
		var estHours, actHours sql.NullFloat64
		var labelsRaw []byte

		err := rows.Scan(
			&t.ID, &t.ProjectID, &t.WorkspaceID, &t.OwnerID, &parentID, &sectionID,
			&t.Title, &desc, &t.Status, &t.Priority,
			&assigneeID, &t.ReporterID,
			&dueDate, &startDate, &completedAt,
			&estHours, &actHours, &t.SortOrder, &labelsRaw, &t.Archived,
			&t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			logger.Ctx(ctx).Error("scanTasksRaw row scan failed", logger.Error(err))
			return nil, err
		}

		t.Description = desc.String
		if parentID.Valid {
			t.ParentTaskID = &parentID.Int64
		}
		if sectionID.Valid {
			t.SectionID = &sectionID.Int64
		}
		if assigneeID.Valid {
			t.AssigneeID = &assigneeID.Int64
		}
		if dueDate.Valid {
			t.DueDate = &dueDate.Time
		}
		if startDate.Valid {
			t.StartDate = &startDate.Time
		}
		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}
		if estHours.Valid {
			t.EstimatedHours = &estHours.Float64
		}
		if actHours.Valid {
			t.ActualHours = &actHours.Float64
		}
		if len(labelsRaw) > 0 {
			if unErr := json.Unmarshal(labelsRaw, &t.Labels); unErr != nil {
				logger.Ctx(ctx).Warn("failed to unmarshal task labels", logger.Error(unErr))
			}
		}
		t.Identifier = "T-" + strconv.FormatInt(t.ID, 10)
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// ─────────────────────────────────────────────
// Workspace Members (extended)
// ─────────────────────────────────────────────
