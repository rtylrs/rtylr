package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// taskFinishedSQL is true for a task (alias t) whose status bucket is done or
// cancelled: the bucket of its custom status when it has one in its project,
// otherwise the built-in status.
const taskFinishedSQL = `(EXISTS (SELECT 1 FROM flow_statuses fsb
          WHERE fsb.id = t.status_id AND fsb.project_id = t.project_id AND fsb.bucket IN ('done', 'cancelled'))
   OR (NOT EXISTS (SELECT 1 FROM flow_statuses fsb WHERE fsb.id = t.status_id AND fsb.project_id = t.project_id)
       AND t.status IN ('done', 'cancelled')))`

// taskCycleMembershipSQL is the membership that gives a task (alias t) its
// cycle: a row in a cycle that is not completed, or in a completed cycle the
// task was finished in. An unfinished task left in a completed cycle is back in
// the unplanned backlog.
const taskCycleMembershipSQL = `(COALESCE(cy.status, 'upcoming') <> 'completed' OR ` + taskFinishedSQL + `)`

// taskUnplannedSQL is true for a task (alias t) in the unplanned backlog.
const taskUnplannedSQL = `NOT EXISTS (SELECT 1 FROM flow_cycle_tasks ct
       JOIN flow_cycles cy ON cy.id = ct.cycle_id
      WHERE ct.task_id = t.id AND ` + taskCycleMembershipSQL + `)`

// taskEstimateMinutesSQL is a task's estimate (alias t) in minutes: the effort
// chips value, or estimated_hours for a task that only has hours.
const taskEstimateMinutesSQL = `COALESCE(t.effort_minutes, ROUND(t.estimated_hours * 60), 0)`

// backlogOrderSQL orders the unplanned backlog. Tasks that were never placed
// follow the ranked ones, oldest first.
const backlogOrderSQL = `(t.backlog_rank IS NULL) ASC, t.backlog_rank ASC, t.created_at ASC, t.id ASC`

func int64Placeholders(ids []int64) (string, []any) {
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	return strings.TrimSuffix(strings.Repeat("?,", len(ids)), ","), args
}

func bucketOf(customBucket, legacyStatus string) string {
	if customBucket != "" {
		return customBucket
	}
	return domain.BucketForTaskStatus(legacyStatus)
}

// AttachTaskCycles sets CycleID, CycleName and PlanRank on every task: the
// cycle membership and its rank, or the backlog rank for a task in no cycle.
func (r *Repository) AttachTaskCycles(ctx context.Context, tasks []domain.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	index := make(map[int64]int, len(tasks))
	ids := make([]int64, 0, len(tasks))
	for i := range tasks {
		index[tasks[i].ID] = i
		ids = append(ids, tasks[i].ID)
		tasks[i].CycleID, tasks[i].CycleName, tasks[i].PlanRank = nil, "", nil
	}
	placeholders, args := int64Placeholders(ids)
	// A cycle that is not completed sorts last, so it wins over the completed
	// cycle a reopened task was once finished in.
	rows, err := r.pool.QueryContext(ctx, `SELECT ct.task_id, cy.id, cy.name, ct.sort_rank
  FROM flow_cycle_tasks ct
  JOIN flow_cycles cy ON cy.id = ct.cycle_id
  JOIN flow_tasks t ON t.id = ct.task_id AND t.business_id = cy.business_id
 WHERE ct.task_id IN (`+placeholders+`) AND `+taskCycleMembershipSQL+`
 ORDER BY (COALESCE(cy.status, 'upcoming') <> 'completed') ASC, cy.id ASC`, args...)
	if err != nil {
		logger.Ctx(ctx).Error("AttachTaskCycles failed", logger.Error(err))
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var taskID, cycleID, rank int64
		var name string
		if err := rows.Scan(&taskID, &cycleID, &name, &rank); err != nil {
			return err
		}
		task := &tasks[index[taskID]]
		task.CycleID, task.CycleName, task.PlanRank = &cycleID, name, &rank
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return r.attachBacklogRanks(ctx, tasks, placeholders, args)
}

func (r *Repository) attachBacklogRanks(ctx context.Context, tasks []domain.Task, placeholders string, args []any) error {
	index := make(map[int64]int, len(tasks))
	for i := range tasks {
		index[tasks[i].ID] = i
	}
	rows, err := r.pool.QueryContext(ctx, `SELECT id, backlog_rank FROM flow_tasks
 WHERE id IN (`+placeholders+`) AND backlog_rank IS NOT NULL`, args...)
	if err != nil {
		logger.Ctx(ctx).Error("attachBacklogRanks failed", logger.Error(err))
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var taskID, rank int64
		if err := rows.Scan(&taskID, &rank); err != nil {
			return err
		}
		if task := &tasks[index[taskID]]; task.CycleID == nil {
			task.PlanRank = &rank
		}
	}
	return rows.Err()
}

// GetCycleV2 returns one cycle of the business with its planning figures.
func (r *Repository) GetCycleV2(ctx context.Context, businessID, cycleID int64) (domain.Cycle, error) {
	var c domain.Cycle
	var projectID sql.NullInt64
	var start, end sql.NullTime
	err := r.pool.QueryRowContext(ctx, `SELECT id, project_id, name, COALESCE(description, ''),
       COALESCE(status, 'upcoming'), start_date, end_date
  FROM flow_cycles WHERE id = ? AND business_id = ?`, cycleID, businessID).
		Scan(&c.ID, &projectID, &c.Name, &c.Description, &c.Status, &start, &end)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Cycle{}, domain.ErrCycleNotFound
	}
	if err != nil {
		logger.Ctx(ctx).Error("GetCycleV2 failed", logger.Error(err))
		return domain.Cycle{}, err
	}
	if projectID.Valid {
		c.ProjectID = &projectID.Int64
	}
	c.StartDate, c.EndDate = start.Time, end.Time
	cycles := []domain.Cycle{c}
	if err := r.enrichCycles(ctx, businessID, cycles); err != nil {
		return domain.Cycle{}, err
	}
	return cycles[0], nil
}

// enrichCycles fills the planning figures of cycles that all belong to
// businessID: goal, lifecycle stamps, totals by bucket, estimate and the tasks
// carried over into each.
func (r *Repository) enrichCycles(ctx context.Context, businessID int64, cycles []domain.Cycle) error {
	if len(cycles) == 0 {
		return nil
	}
	index := make(map[int64]*domain.Cycle, len(cycles))
	ids := make([]int64, 0, len(cycles))
	for i := range cycles {
		index[cycles[i].ID] = &cycles[i]
		ids = append(ids, cycles[i].ID)
	}
	placeholders, idArgs := int64Placeholders(ids)
	args := append([]any{businessID}, idArgs...)

	for _, c := range index {
		c.BusinessID = businessID
	}
	if err := r.readCycleFields(ctx, placeholders, args, index); err != nil {
		return err
	}
	if err := r.readCycleTotals(ctx, placeholders, args, index); err != nil {
		return err
	}
	return r.readCycleCarryOvers(ctx, placeholders, args, index)
}

func (r *Repository) readCycleFields(ctx context.Context, placeholders string, args []any, index map[int64]*domain.Cycle) error {
	rows, err := r.pool.QueryContext(ctx, `SELECT id, COALESCE(goal, ''), started_at, completed_at
  FROM flow_cycles WHERE business_id = ? AND id IN (`+placeholders+`)`, args...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id int64
		var goal string
		var startedAt, completedAt sql.NullTime
		if err := rows.Scan(&id, &goal, &startedAt, &completedAt); err != nil {
			return err
		}
		c := index[id]
		c.Goal = goal
		if startedAt.Valid {
			c.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			c.CompletedAt = &completedAt.Time
		}
	}
	return rows.Err()
}

func (r *Repository) readCycleTotals(ctx context.Context, placeholders string, args []any, index map[int64]*domain.Cycle) error {
	rows, err := r.pool.QueryContext(ctx, `SELECT ct.cycle_id, COALESCE(fs.bucket, ''), COALESCE(t.status, 'todo'),
       COUNT(*), COALESCE(SUM(`+taskEstimateMinutesSQL+`), 0)
  FROM flow_cycle_tasks ct
  JOIN flow_cycles cy ON cy.id = ct.cycle_id AND cy.business_id = ?
  JOIN flow_tasks t ON t.id = ct.task_id AND t.business_id = cy.business_id
  LEFT JOIN flow_statuses fs ON fs.id = t.status_id AND fs.project_id = t.project_id
 WHERE ct.cycle_id IN (`+placeholders+`)
 GROUP BY ct.cycle_id, COALESCE(fs.bucket, ''), COALESCE(t.status, 'todo')`, args...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for _, c := range index {
		c.Totals, c.EstimateMinutes, c.Scope, c.Completed = domain.CycleTotals{}, 0, 0, 0
	}
	for rows.Next() {
		var cycleID int64
		var customBucket, legacyStatus string
		var count int
		var minutes float64
		if err := rows.Scan(&cycleID, &customBucket, &legacyStatus, &count, &minutes); err != nil {
			return err
		}
		c := index[cycleID]
		c.Scope += count
		c.EstimateMinutes += int64(minutes)
		switch bucketOf(customBucket, legacyStatus) {
		case domain.StatusBucketActive:
			c.Totals.Active += count
		case domain.StatusBucketDone:
			c.Totals.Done += count
			c.Completed += count
		case domain.StatusBucketCancelled:
			c.Totals.Cancelled += count
		default:
			c.Totals.Backlog += count
		}
	}
	return rows.Err()
}

func (r *Repository) readCycleCarryOvers(ctx context.Context, placeholders string, args []any, index map[int64]*domain.Cycle) error {
	rows, err := r.pool.QueryContext(ctx, `SELECT co.to_cycle_id, co.cycle_id, src.name, co.task_id
  FROM flow_cycle_carryovers co
  JOIN flow_cycles src ON src.id = co.cycle_id AND src.business_id = co.business_id
  JOIN flow_cycle_tasks ct ON ct.cycle_id = co.to_cycle_id AND ct.task_id = co.task_id
 WHERE co.business_id = ? AND co.decision = 'move' AND co.to_cycle_id IN (`+placeholders+`)
 ORDER BY co.cycle_id, co.id`, args...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var toCycleID, fromCycleID, taskID int64
		var fromName string
		if err := rows.Scan(&toCycleID, &fromCycleID, &fromName, &taskID); err != nil {
			return err
		}
		c := index[toCycleID]
		last := len(c.CarriedOver) - 1
		if last < 0 || c.CarriedOver[last].FromCycleID != fromCycleID {
			c.CarriedOver = append(c.CarriedOver, domain.CycleCarryOver{FromCycleID: fromCycleID, FromCycleName: fromName})
			last++
		}
		c.CarriedOver[last].TaskIDs = append(c.CarriedOver[last].TaskIDs, taskID)
	}
	return rows.Err()
}

// CycleMemberLoads sums the estimate of a cycle's tasks per assignee.
func (r *Repository) CycleMemberLoads(ctx context.Context, businessID, cycleID int64) ([]domain.CycleMemberLoad, error) {
	rows, err := r.pool.QueryContext(ctx, `SELECT COALESCE(t.assignee_id, 0), COUNT(*), COALESCE(SUM(`+taskEstimateMinutesSQL+`), 0)
  FROM flow_cycle_tasks ct
  JOIN flow_cycles cy ON cy.id = ct.cycle_id AND cy.business_id = ?
  JOIN flow_tasks t ON t.id = ct.task_id AND t.business_id = cy.business_id
 WHERE ct.cycle_id = ?
 GROUP BY COALESCE(t.assignee_id, 0)`, businessID, cycleID)
	if err != nil {
		logger.Ctx(ctx).Error("CycleMemberLoads failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var loads []domain.CycleMemberLoad
	for rows.Next() {
		var load domain.CycleMemberLoad
		var minutes float64
		if err := rows.Scan(&load.UserID, &load.TaskCount, &minutes); err != nil {
			return nil, err
		}
		load.Minutes = int64(minutes)
		loads = append(loads, load)
	}
	return loads, rows.Err()
}
