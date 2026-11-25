package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// Every planning write runs in one transaction that first locks the project
// row. That serialises start, complete and move within a project, which is what
// keeps "one active cycle" and the rank order correct under concurrent calls.

type cycleRow struct {
	ID        int64
	ProjectID int64
	Status    string
	StartDate time.Time
	EndDate   time.Time
}

// lockCycle locks the cycle's project, then the cycle, and returns the cycle.
// A cycle of another business is reported as not found.
func lockCycle(ctx context.Context, tx *sql.Tx, businessID, cycleID int64) (cycleRow, error) {
	var projectID sql.NullInt64
	err := tx.QueryRowContext(ctx, "SELECT project_id FROM flow_cycles WHERE id = ? AND business_id = ?",
		cycleID, businessID).Scan(&projectID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !projectID.Valid) {
		return cycleRow{}, domain.ErrCycleNotFound
	}
	if err != nil {
		return cycleRow{}, err
	}
	if err := lockProject(ctx, tx, businessID, projectID.Int64); err != nil {
		return cycleRow{}, err
	}
	return readCycleForUpdate(ctx, tx, businessID, cycleID)
}

func lockProject(ctx context.Context, tx *sql.Tx, businessID, projectID int64) error {
	var id int64
	err := tx.QueryRowContext(ctx, "SELECT id FROM flow_projects WHERE id = ? AND business_id = ? FOR UPDATE",
		projectID, businessID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrProjectNotFound
	}
	return err
}

func readCycleForUpdate(ctx context.Context, tx *sql.Tx, businessID, cycleID int64) (cycleRow, error) {
	var c cycleRow
	var start, end sql.NullTime
	err := tx.QueryRowContext(ctx, `SELECT id, COALESCE(project_id, 0), COALESCE(status, 'upcoming'), start_date, end_date
  FROM flow_cycles WHERE id = ? AND business_id = ? FOR UPDATE`, cycleID, businessID).
		Scan(&c.ID, &c.ProjectID, &c.Status, &start, &end)
	if errors.Is(err, sql.ErrNoRows) {
		return cycleRow{}, domain.ErrCycleNotFound
	}
	c.StartDate, c.EndDate = start.Time, end.Time
	return c, err
}

func (r *Repository) inCycleTx(ctx context.Context, operation string, run func(tx *sql.Tx) error) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := run(tx); err != nil {
		if !isCyclePlanningError(err) {
			logger.Ctx(ctx).Error(operation+" failed", logger.Error(err))
		}
		return err
	}
	return tx.Commit()
}

func isCyclePlanningError(err error) bool {
	var decision *domain.CycleDecisionError
	if errors.As(err, &decision) {
		return true
	}
	for _, known := range []error{
		domain.ErrCycleNotFound, domain.ErrCycleAlreadyActive, domain.ErrCycleNotUpcoming,
		domain.ErrCycleCompleted, domain.ErrCycleTaskNotFound, domain.ErrCycleRankAnchor,
		domain.ErrInvalidCycleMove, domain.ErrInvalidCycleRequest, domain.ErrProjectNotFound,
	} {
		if errors.Is(err, known) {
			return true
		}
	}
	return false
}

// UpdateCycleV2 writes the fields the caller supplied. An empty name and zero
// dates are left alone; a nil goal is left alone and an empty one clears it.
func (r *Repository) UpdateCycleV2(ctx context.Context, in domain.UpdateCycleV2Input) error {
	var parts []string
	var args []any
	if in.Name != "" {
		parts, args = append(parts, "name = ?"), append(args, in.Name)
	}
	if !in.StartDate.IsZero() {
		parts, args = append(parts, "start_date = ?"), append(args, in.StartDate.Format("2006-01-02"))
	}
	if !in.EndDate.IsZero() {
		parts, args = append(parts, "end_date = ?"), append(args, in.EndDate.Format("2006-01-02"))
	}
	if in.Goal != nil {
		parts, args = append(parts, "goal = ?"), append(args, sql.NullString{String: *in.Goal, Valid: *in.Goal != ""})
	}
	if len(parts) == 0 {
		return nil
	}
	args = append(args, in.ID, in.BusinessID)
	_, err := r.pool.ExecContext(ctx,
		"UPDATE flow_cycles SET "+strings.Join(parts, ", ")+" WHERE id = ? AND business_id = ?", args...)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateCycleV2 failed", logger.Error(err))
	}
	return err
}

// DeleteCycle removes the cycle and its memberships, so its tasks fall back
// to the unplanned backlog. Carry-over records written when it closed go with
// it, and records that pointed at it as the destination keep the decision but
// lose the destination.
func (r *Repository) DeleteCycle(ctx context.Context, businessID, cycleID int64) error {
	return r.inCycleTx(ctx, "DeleteCycle", func(tx *sql.Tx) error {
		if _, err := lockCycle(ctx, tx, businessID, cycleID); err != nil {
			return err
		}
		for _, stmt := range []string{
			"DELETE FROM flow_cycle_tasks WHERE cycle_id = ?",
			"DELETE FROM flow_cycle_carryovers WHERE cycle_id = ? AND business_id = ?",
			"UPDATE flow_cycle_carryovers SET to_cycle_id = NULL WHERE to_cycle_id = ? AND business_id = ?",
			"DELETE FROM flow_cycles WHERE id = ? AND business_id = ?",
		} {
			args := []any{cycleID}
			if strings.Count(stmt, "?") == 2 {
				args = append(args, businessID)
			}
			if _, err := tx.ExecContext(ctx, stmt, args...); err != nil {
				return err
			}
		}
		return nil
	})
}

// StartCycle makes an upcoming cycle the project's active one. It fails with
// ErrCycleAlreadyActive when another cycle of the project is active.
func (r *Repository) StartCycle(ctx context.Context, in domain.StartCycleInput) error {
	return r.inCycleTx(ctx, "StartCycle", func(tx *sql.Tx) error {
		cycle, err := lockCycle(ctx, tx, in.BusinessID, in.ID)
		if err != nil {
			return err
		}
		if cycle.Status != domain.CycleStatusUpcoming {
			return domain.ErrCycleNotUpcoming
		}
		var activeID int64
		err = tx.QueryRowContext(ctx, `SELECT id FROM flow_cycles
 WHERE project_id = ? AND business_id = ? AND status = 'active' AND id <> ? LIMIT 1`,
			cycle.ProjectID, in.BusinessID, cycle.ID).Scan(&activeID)
		if err == nil {
			return domain.ErrCycleAlreadyActive
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		start, end, err := domain.ResolveCycleWindow(in, cycle.StartDate, cycle.EndDate)
		if err != nil {
			return err
		}
		query := "UPDATE flow_cycles SET status = 'active', started_at = NOW(), start_date = ?, end_date = ?"
		args := []any{start.Format("2006-01-02"), end.Format("2006-01-02")}
		if in.Goal != nil {
			query += ", goal = ?"
			args = append(args, sql.NullString{String: *in.Goal, Valid: *in.Goal != ""})
		}
		_, err = tx.ExecContext(ctx, query+" WHERE id = ? AND business_id = ?", append(args, cycle.ID, in.BusinessID)...)
		return err
	})
}

// MoveCycleTask places a task in a cycle, or in the backlog when CycleID is
// zero, at the position the anchors describe.
func (r *Repository) MoveCycleTask(ctx context.Context, in domain.MoveCycleTaskInput) (domain.MovedCycleTask, error) {
	var moved domain.MovedCycleTask
	err := r.inCycleTx(ctx, "MoveCycleTask", func(tx *sql.Tx) error {
		projectID, err := taskProject(ctx, tx, in.BusinessID, in.TaskID)
		if err != nil {
			return err
		}
		if err := lockProject(ctx, tx, in.BusinessID, projectID); err != nil {
			return err
		}
		if in.CycleID != 0 {
			if err := requireOpenCycle(ctx, tx, in.BusinessID, in.CycleID, projectID); err != nil {
				return err
			}
		}
		placed, err := placeTask(ctx, tx, placement{
			BusinessID: in.BusinessID, ProjectID: projectID, TaskID: in.TaskID,
			CycleID: in.CycleID, BeforeTaskID: in.BeforeTaskID, AfterTaskID: in.AfterTaskID,
		})
		moved = domain.MovedCycleTask{ProjectID: projectID, Rank: placed}
		return err
	})
	return moved, err
}

func taskProject(ctx context.Context, tx *sql.Tx, businessID, taskID int64) (int64, error) {
	var projectID int64
	err := tx.QueryRowContext(ctx, "SELECT project_id FROM flow_tasks WHERE id = ? AND business_id = ?",
		taskID, businessID).Scan(&projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrCycleTaskNotFound
	}
	return projectID, err
}

// requireOpenCycle checks that cycleID is a cycle of projectID in the business
// that still accepts tasks.
func requireOpenCycle(ctx context.Context, tx *sql.Tx, businessID, cycleID, projectID int64) error {
	cycle, err := readCycleForUpdate(ctx, tx, businessID, cycleID)
	if err != nil {
		return err
	}
	if cycle.ProjectID != projectID {
		return domain.ErrCycleNotFound
	}
	if cycle.Status == domain.CycleStatusCompleted {
		return domain.ErrCycleCompleted
	}
	return nil
}

type placement struct {
	BusinessID   int64
	ProjectID    int64
	TaskID       int64
	CycleID      int64 // zero is the backlog
	BeforeTaskID int64
	AfterTaskID  int64
}

// placeTask takes the task out of any other open cycle and ranks it in the
// target list. The caller holds the project lock.
func placeTask(ctx context.Context, tx *sql.Tx, p placement) (int64, error) {
	if _, err := tx.ExecContext(ctx, `DELETE ct FROM flow_cycle_tasks ct
  JOIN flow_cycles cy ON cy.id = ct.cycle_id
 WHERE ct.task_id = ? AND cy.business_id = ? AND ct.cycle_id <> ?
   AND COALESCE(cy.status, 'upcoming') <> 'completed'`, p.TaskID, p.BusinessID, p.CycleID); err != nil {
		return 0, err
	}
	list, err := rankedList(ctx, tx, p)
	if err != nil {
		return 0, err
	}
	spot, err := domain.PlaceRank(list, p.TaskID, p.BeforeTaskID, p.AfterTaskID, p.CycleID == 0)
	if err != nil {
		return 0, err
	}
	if err := writeRenumber(ctx, tx, p, spot.Renumber); err != nil {
		return 0, err
	}
	if p.CycleID == 0 {
		_, err = tx.ExecContext(ctx,
			"UPDATE flow_tasks SET backlog_rank = ?, updated_at = updated_at WHERE id = ? AND business_id = ?",
			spot.Rank, p.TaskID, p.BusinessID)
		return spot.Rank, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO flow_cycle_tasks (cycle_id, task_id, sort_rank) VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE sort_rank = VALUES(sort_rank)`, p.CycleID, p.TaskID, spot.Rank)
	return spot.Rank, err
}

// rankedList reads the target list in display order, without the moved task.
func rankedList(ctx context.Context, tx *sql.Tx, p placement) ([]domain.RankedTask, error) {
	query := `SELECT ct.task_id, ct.sort_rank FROM flow_cycle_tasks ct
 WHERE ct.cycle_id = ? AND ct.task_id <> ?
 ORDER BY ct.sort_rank, ct.added_at, ct.task_id`
	args := []any{p.CycleID, p.TaskID}
	if p.CycleID == 0 {
		query = `SELECT t.id, t.backlog_rank FROM flow_tasks t
 WHERE t.project_id = ? AND t.business_id = ? AND t.id <> ? AND ` + taskUnplannedSQL + `
 ORDER BY ` + backlogOrderSQL
		args = []any{p.ProjectID, p.BusinessID, p.TaskID}
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var list []domain.RankedTask
	for rows.Next() {
		var id int64
		var rank sql.NullInt64
		if err := rows.Scan(&id, &rank); err != nil {
			return nil, err
		}
		list = append(list, domain.RankedTask{TaskID: id, Rank: rank.Int64, Ranked: rank.Valid})
	}
	return list, rows.Err()
}

const renumberChunk = 400

// writeRenumber stores fresh ranks for the tasks already in the list. The moved
// task is written by the caller.
func writeRenumber(ctx context.Context, tx *sql.Tx, p placement, ranks map[int64]int64) error {
	ids := make([]int64, 0, len(ranks))
	for id := range ranks {
		if id != p.TaskID {
			ids = append(ids, id)
		}
	}
	for start := 0; start < len(ids); start += renumberChunk {
		chunk := ids[start:min(start+renumberChunk, len(ids))]
		var cases strings.Builder
		args := make([]any, 0, 3*len(chunk)+1)
		for _, id := range chunk {
			cases.WriteString(" WHEN ? THEN ?")
			args = append(args, id, ranks[id])
		}
		placeholders, idArgs := int64Placeholders(chunk)
		query := "UPDATE flow_cycle_tasks SET sort_rank = CASE task_id" + cases.String() +
			" END WHERE cycle_id = ? AND task_id IN (" + placeholders + ")"
		scope := p.CycleID
		if p.CycleID == 0 {
			query = "UPDATE flow_tasks SET backlog_rank = CASE id" + cases.String() +
				" END, updated_at = updated_at WHERE project_id = ? AND id IN (" + placeholders + ")"
			scope = p.ProjectID
		}
		args = append(append(args, scope), idArgs...)
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}
	return nil
}
