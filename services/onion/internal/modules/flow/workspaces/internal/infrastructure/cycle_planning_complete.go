package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

type unfinishedTask struct {
	TaskID int64
	Status string
}

// CompleteCycle closes a cycle and settles each unfinished task by its
// decision, or sends it to the backlog when it has none. The close, every task
// change and the carry-over records commit together or not at all.
func (r *Repository) CompleteCycle(ctx context.Context, in domain.CompleteCycleInput) ([]domain.CompletedCycleTask, error) {
	var settled []domain.CompletedCycleTask
	err := r.inCycleTx(ctx, "CompleteCycle", func(tx *sql.Tx) error {
		cycle, err := lockCycle(ctx, tx, in.BusinessID, in.ID)
		if err != nil {
			return err
		}
		if cycle.Status == domain.CycleStatusCompleted {
			return domain.ErrCycleCompleted
		}
		unfinished, err := unfinishedCycleTasks(ctx, tx, in.BusinessID, cycle.ID)
		if err != nil {
			return err
		}
		plan, err := planCycleCompletion(unfinished, in.Decisions)
		if err != nil {
			return err
		}
		for _, item := range plan {
			if err := settleCycleTask(ctx, tx, in, cycle, item); err != nil {
				return err
			}
		}
		settled = plan
		_, err = tx.ExecContext(ctx,
			"UPDATE flow_cycles SET status = 'completed', completed_at = NOW() WHERE id = ? AND business_id = ?",
			cycle.ID, in.BusinessID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return settled, nil
}

// unfinishedCycleTasks lists, in rank order, the cycle's tasks whose status
// bucket is backlog or active.
func unfinishedCycleTasks(ctx context.Context, tx *sql.Tx, businessID, cycleID int64) ([]unfinishedTask, error) {
	rows, err := tx.QueryContext(ctx, `SELECT ct.task_id, COALESCE(fs.bucket, ''), COALESCE(t.status, 'todo')
  FROM flow_cycle_tasks ct
  JOIN flow_tasks t ON t.id = ct.task_id AND t.business_id = ?
  LEFT JOIN flow_statuses fs ON fs.id = t.status_id AND fs.project_id = t.project_id
 WHERE ct.cycle_id = ?
 ORDER BY ct.sort_rank, ct.added_at, ct.task_id`, businessID, cycleID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var tasks []unfinishedTask
	for rows.Next() {
		var task unfinishedTask
		var customBucket string
		if err := rows.Scan(&task.TaskID, &customBucket, &task.Status); err != nil {
			return nil, err
		}
		switch bucketOf(customBucket, task.Status) {
		case domain.StatusBucketDone, domain.StatusBucketCancelled:
		default:
			tasks = append(tasks, task)
		}
	}
	return tasks, rows.Err()
}

// planCycleCompletion pairs every unfinished task with its decision. Moves and
// cancels keep rank order; backlog tasks come last and reversed, because each
// is placed at the top of the backlog and the last one placed ends up first.
func planCycleCompletion(unfinished []unfinishedTask, decisions []domain.CycleTaskDecision) ([]domain.CompletedCycleTask, error) {
	open := make(map[int64]bool, len(unfinished))
	for _, task := range unfinished {
		open[task.TaskID] = true
	}
	chosen := make(map[int64]domain.CycleTaskDecision, len(decisions))
	for _, d := range decisions {
		if !open[d.TaskID] {
			return nil, &domain.CycleDecisionError{TaskID: d.TaskID, Reason: "the task is not an unfinished task of this cycle"}
		}
		if _, dup := chosen[d.TaskID]; dup {
			return nil, &domain.CycleDecisionError{TaskID: d.TaskID, Reason: "the task has more than one decision"}
		}
		switch d.Decision {
		case domain.CycleDecisionBacklog, domain.CycleDecisionCancel:
			d.ToCycleID = 0
		case domain.CycleDecisionMove:
			if d.ToCycleID <= 0 {
				return nil, &domain.CycleDecisionError{TaskID: d.TaskID, Reason: "a move decision needs to_cycle_id"}
			}
		default:
			return nil, &domain.CycleDecisionError{TaskID: d.TaskID, Reason: "decision must be move, backlog or cancel"}
		}
		chosen[d.TaskID] = d
	}

	plan := make([]domain.CompletedCycleTask, 0, len(unfinished))
	var toBacklog []domain.CompletedCycleTask
	for _, task := range unfinished {
		item := domain.CompletedCycleTask{TaskID: task.TaskID, Decision: domain.CycleDecisionBacklog, PreviousStatus: task.Status}
		if d, ok := chosen[task.TaskID]; ok {
			item.Decision, item.ToCycleID = d.Decision, d.ToCycleID
		}
		if item.Decision == domain.CycleDecisionBacklog {
			toBacklog = append(toBacklog, item)
			continue
		}
		plan = append(plan, item)
	}
	for i := len(toBacklog) - 1; i >= 0; i-- {
		plan = append(plan, toBacklog[i])
	}
	return plan, nil
}

func settleCycleTask(ctx context.Context, tx *sql.Tx, in domain.CompleteCycleInput, cycle cycleRow, item domain.CompletedCycleTask) error {
	spot := placement{BusinessID: in.BusinessID, ProjectID: cycle.ProjectID, TaskID: item.TaskID}
	var toCycle sql.NullInt64
	switch item.Decision {
	case domain.CycleDecisionMove:
		if err := requireCarryOverTarget(ctx, tx, in.BusinessID, cycle, item); err != nil {
			return err
		}
		spot.CycleID = item.ToCycleID
		toCycle = sql.NullInt64{Int64: item.ToCycleID, Valid: true}
		if _, err := placeTask(ctx, tx, spot); err != nil {
			return err
		}
	case domain.CycleDecisionCancel:
		// The task stays in the cycle it was cancelled in.
		if _, err := tx.ExecContext(ctx,
			"UPDATE flow_tasks SET status = 'cancelled', completed_at = NULL WHERE id = ? AND business_id = ?",
			item.TaskID, in.BusinessID); err != nil {
			return err
		}
		if err := syncTaskFromLegacy(ctx, tx, item.TaskID); err != nil {
			return err
		}
	default:
		if _, err := placeTask(ctx, tx, spot); err != nil {
			return err
		}
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO flow_cycle_carryovers
(business_id, cycle_id, task_id, to_cycle_id, decision, created_by) VALUES (?, ?, ?, ?, ?, ?)`,
		in.BusinessID, cycle.ID, item.TaskID, toCycle, item.Decision, in.ActorID)
	return err
}

func requireCarryOverTarget(ctx context.Context, tx *sql.Tx, businessID int64, cycle cycleRow, item domain.CompletedCycleTask) error {
	if item.ToCycleID == cycle.ID {
		return &domain.CycleDecisionError{TaskID: item.TaskID, Reason: "a task cannot be moved to the cycle being completed"}
	}
	err := requireOpenCycle(ctx, tx, businessID, item.ToCycleID, cycle.ProjectID)
	switch {
	case errors.Is(err, domain.ErrCycleNotFound):
		return &domain.CycleDecisionError{TaskID: item.TaskID, Reason: "the target cycle is not a cycle of this project"}
	case errors.Is(err, domain.ErrCycleCompleted):
		return &domain.CycleDecisionError{TaskID: item.TaskID, Reason: "the target cycle is completed"}
	}
	return err
}
