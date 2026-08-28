package infrastructure

import (
	"context"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	searchutil "github.com/voxire/onion/internal/platform/searchutil"
)

// projectVisibilityExpr is the per-row gate that makes task listings honour
// the privacy model: the parent project must be one the caller created, owns,
// or belongs to. includeAll (hub bypass) skips the clause entirely.
func projectVisibilityExpr(callerID int64, includeAll bool) squirrel.Sqlizer {
	if includeAll {
		return nil
	}
	return squirrel.Expr(
		`EXISTS (SELECT 1 FROM flow_projects p
		  WHERE p.id = t.project_id
		    AND (p.created_by = ? OR p.owner_id = ?
		         OR EXISTS (SELECT 1 FROM flow_project_members pm
		                    WHERE pm.project_id = p.id AND pm.user_id = ?)))`,
		callerID, callerID, callerID,
	)
}

// GetTasksByBusinessV2 lists business-wide tasks under the visibility model.
// With a ProjectID filter it behaves like the v1 project listing; without one
// it spans every visible project in the business, replacing v1's
// workspace-scoped mode.
func (r *Repository) GetTasksByBusinessV2(ctx context.Context, filter domain.ListTasksFilter, callerID int64, includeAll bool) ([]domain.Task, int64, error) {
	conds := taskListConditions("t.business_id", filter.BusinessID, filter)
	if vis := projectVisibilityExpr(callerID, includeAll); vis != nil {
		conds = append(conds, vis)
	}

	total, err := r.countTasks(ctx, conds)
	if err != nil {
		return nil, 0, err
	}

	builder := enrichedTaskSelectForList(filter.Compact).
		Where(conds).
		OrderBy("t.sort_order ASC", "t.created_at ASC").
		Limit(uint64(filter.Limit)).  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset(uint64(filter.Offset)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.

	tasks, err := r.listEnrichedTasks(ctx, builder)
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

// SearchTasksV2 is business-wide fulltext search gated by the same visibility
// clause. Mirrors SearchTasks' raw-SQL shape because MATCH AGAINST cannot be
// expressed through squirrel.
func (r *Repository) SearchTasksV2(ctx context.Context, businessID, callerID int64, includeAll bool, q string, limit, offset int64) ([]domain.Task, int64, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []domain.Task{}, 0, nil
	}
	searchQuery := searchutil.BooleanPrefixQuery(q)
	if searchQuery == "" {
		return []domain.Task{}, 0, nil
	}

	visibilityJoin := ""
	visArgs := []interface{}{}
	if !includeAll {
		visibilityJoin = ` AND EXISTS (SELECT 1 FROM flow_projects p
		      WHERE p.id = flow_tasks.project_id
		        AND (p.created_by = ? OR p.owner_id = ?
		             OR EXISTS (SELECT 1 FROM flow_project_members pm
		                        WHERE pm.project_id = p.id AND pm.user_id = ?)))`
		visArgs = append(visArgs, callerID, callerID, callerID)
	}

	countQuery := `SELECT COUNT(*)
	          FROM flow_tasks
	          WHERE business_id = ?
	            AND archived = FALSE
	            AND MATCH(title, description) AGAINST (? IN BOOLEAN MODE)` + visibilityJoin
	args := append([]interface{}{businessID, searchQuery}, visArgs...)
	var total int64
	if err := r.pool.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("SearchTasksV2 count failed", logger.Error(err))
		return nil, 0, err
	}

	query := `SELECT id, project_id, workspace_id, owner_id, parent_task_id, section_id, title, description,
	           status, priority, assignee_id, reporter_id, due_date, start_date, completed_at,
	           estimated_hours, actual_hours, sort_order, labels, COALESCE(archived, FALSE) AS archived,
	           created_at, updated_at
	          FROM flow_tasks
	          WHERE business_id = ?
	            AND archived = FALSE
	            AND MATCH(title, description) AGAINST (? IN BOOLEAN MODE)` + visibilityJoin + `
	          ORDER BY updated_at DESC
	          LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("SearchTasksV2 failed", logger.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	tasks, err := scanTasksRaw(ctx, rows)
	if err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}
