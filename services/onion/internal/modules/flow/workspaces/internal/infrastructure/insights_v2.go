package infrastructure

import (
	"context"
	"database/sql"
	"math"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// progressTaskSQL limits a task (alias t, custom status fs) to the tasks that
// make up project progress: unarchived and top-level.
const progressTaskSQL = `t.parent_task_id IS NULL AND COALESCE(t.archived, FALSE) = FALSE`

// GetInsightsV2 reads the delivery series of projectIDs, which the caller has
// already narrowed to the projects the viewer may see. TotalTasks, DoneTasks
// and the member names are left to the caller.
func (r *Repository) GetInsightsV2(ctx context.Context, businessID int64, projectIDs []int64, now time.Time) (domain.Insights, error) {
	ins := domain.Insights{
		Velocity: []domain.InsightsVelocityPoint{}, Burndown: []domain.InsightsBurndownPoint{},
		TimeByMember: []domain.InsightsMemberTime{}, Throughput: []domain.InsightsThroughput{},
	}
	if len(projectIDs) == 0 {
		return ins, nil
	}
	placeholders, idArgs := int64Placeholders(projectIDs)
	scope := func(extra ...any) []any { return append(append([]any{businessID}, idArgs...), extra...) }

	var err error
	if ins.Throughput, err = r.insightsThroughput(ctx, placeholders, scope, now); err != nil {
		return domain.Insights{}, err
	}
	if ins.CycleTime, err = r.insightsCycleTime(ctx, placeholders, scope, now); err != nil {
		return domain.Insights{}, err
	}
	if ins.Velocity, err = r.insightsVelocity(ctx, businessID, placeholders, scope); err != nil {
		return domain.Insights{}, err
	}
	if ins.Burndown, err = r.insightsBurndown(ctx, businessID, placeholders, scope, now, len(projectIDs)); err != nil {
		return domain.Insights{}, err
	}
	if ins.TimeByMember, err = r.insightsTimeLogMinutes(ctx, placeholders, scope); err != nil {
		return domain.Insights{}, err
	}
	return ins, nil
}

func (r *Repository) insightsThroughput(ctx context.Context, placeholders string, scope func(...any) []any, now time.Time) ([]domain.InsightsThroughput, error) {
	since := domain.WeekStart(now).AddDate(0, 0, -7*(domain.InsightsThroughputWeeks-1))
	rows, err := r.pool.QueryContext(ctx, `SELECT DATE_FORMAT(t.completed_at, '%Y-%m-%d'), COUNT(*)
  FROM flow_tasks t
  LEFT JOIN flow_statuses fs ON fs.id = t.status_id AND fs.project_id = t.project_id
 WHERE t.business_id = ? AND t.project_id IN (`+placeholders+`) AND `+progressTaskSQL+`
   AND `+taskBucketSQL+` = 'done' AND t.completed_at >= ?
 GROUP BY DATE_FORMAT(t.completed_at, '%Y-%m-%d')`, scope(since)...)
	if err != nil {
		logger.Ctx(ctx).Error("insights throughput failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	byDay := map[string]int{}
	for rows.Next() {
		var day string
		var count int
		if err := rows.Scan(&day, &count); err != nil {
			return nil, err
		}
		byDay[day] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return domain.WeeklyThroughput(byDay, now, domain.InsightsThroughputWeeks), nil
}

// insightsCycleTime is the average days from creation to completion of the
// tasks finished in the last InsightsCycleTimeDays, to one decimal.
func (r *Repository) insightsCycleTime(ctx context.Context, placeholders string, scope func(...any) []any, now time.Time) (float64, error) {
	var avgSeconds sql.NullFloat64
	err := r.pool.QueryRowContext(ctx, `SELECT AVG(TIMESTAMPDIFF(SECOND, t.created_at, t.completed_at))
  FROM flow_tasks t
  LEFT JOIN flow_statuses fs ON fs.id = t.status_id AND fs.project_id = t.project_id
 WHERE t.business_id = ? AND t.project_id IN (`+placeholders+`) AND `+progressTaskSQL+`
   AND `+taskBucketSQL+` = 'done' AND t.completed_at >= ? AND t.completed_at >= t.created_at`,
		scope(now.AddDate(0, 0, -domain.InsightsCycleTimeDays))...).Scan(&avgSeconds)
	if err != nil {
		logger.Ctx(ctx).Error("insights cycle time failed", logger.Error(err))
		return 0, err
	}
	return math.Round(avgSeconds.Float64/86400*10) / 10, nil
}

// insightsCycles returns the cycles of the projects in one lifecycle status,
// with their totals, and each one's project name by cycle id.
func (r *Repository) insightsCycles(ctx context.Context, businessID int64, placeholders string, scope func(...any) []any, status, order string, limit int) ([]domain.Cycle, map[int64]string, error) {
	rows, err := r.pool.QueryContext(ctx, `SELECT cy.id, cy.project_id, cy.name, p.name, cy.start_date, cy.end_date
  FROM flow_cycles cy
  JOIN flow_projects p ON p.id = cy.project_id AND p.business_id = cy.business_id
 WHERE cy.business_id = ? AND cy.project_id IN (`+placeholders+`) AND cy.status = ?
 ORDER BY `+order+`
 LIMIT ?`, scope(status, limit)...)
	if err != nil {
		logger.Ctx(ctx).Error("insights cycles failed", logger.Error(err))
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	var cycles []domain.Cycle
	projectNames := map[int64]string{}
	for rows.Next() {
		var c domain.Cycle
		var projectID int64
		var projectName string
		var start, end sql.NullTime
		if err := rows.Scan(&c.ID, &projectID, &c.Name, &projectName, &start, &end); err != nil {
			return nil, nil, err
		}
		c.ProjectID, c.Status, c.StartDate, c.EndDate = &projectID, status, start.Time, end.Time
		cycles = append(cycles, c)
		projectNames[c.ID] = projectName
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return cycles, projectNames, r.enrichCycles(ctx, businessID, cycles)
}

// insightsVelocity is, for the most recently completed cycles (oldest first),
// the tasks done in each against the work it held at its close: its scope
// plus the unfinished tasks it carried over or sent back to the backlog, which
// completing the cycle takes out of it. CompletionRate is a percentage. Cycle
// names repeat across projects, so a series spanning several names both.
func (r *Repository) insightsVelocity(ctx context.Context, businessID int64, placeholders string, scope func(...any) []any) ([]domain.InsightsVelocityPoint, error) {
	cycles, projectNames, err := r.insightsCycles(ctx, businessID, placeholders, scope, domain.CycleStatusCompleted,
		"COALESCE(cy.completed_at, cy.end_date) DESC, cy.id DESC", domain.InsightsVelocityCycles)
	if err != nil {
		return nil, err
	}
	leftOpen, err := r.cyclesLeftOpen(ctx, businessID, cycles)
	if err != nil {
		return nil, err
	}
	projects := map[int64]bool{}
	for _, c := range cycles {
		projects[*c.ProjectID] = true
	}
	points := make([]domain.InsightsVelocityPoint, 0, len(cycles))
	for i := len(cycles) - 1; i >= 0; i-- {
		c := cycles[i]
		label := c.Name
		if len(projects) > 1 {
			label = projectNames[c.ID] + " · " + c.Name
		}
		total := c.Scope + leftOpen[c.ID]
		rate := 0.0
		if total > 0 {
			rate = math.Round(float64(c.Completed)/float64(total)*1000) / 10
		}
		points = append(points, domain.InsightsVelocityPoint{Cycle: label, Completed: c.Completed, Total: total, CompletionRate: rate})
	}
	return points, nil
}

// cyclesLeftOpen counts, per completed cycle, the unarchived tasks it moved on
// unfinished when it closed.
func (r *Repository) cyclesLeftOpen(ctx context.Context, businessID int64, cycles []domain.Cycle) (map[int64]int, error) {
	counts := map[int64]int{}
	if len(cycles) == 0 {
		return counts, nil
	}
	ids := make([]int64, 0, len(cycles))
	for _, c := range cycles {
		ids = append(ids, c.ID)
	}
	placeholders, idArgs := int64Placeholders(ids)
	rows, err := r.pool.QueryContext(ctx, `SELECT co.cycle_id, COUNT(DISTINCT co.task_id)
  FROM flow_cycle_carryovers co
  JOIN flow_tasks t ON t.id = co.task_id AND t.business_id = co.business_id
 WHERE co.business_id = ? AND co.cycle_id IN (`+placeholders+`)
   AND co.decision IN ('move', 'backlog') AND COALESCE(t.archived, FALSE) = FALSE
 GROUP BY co.cycle_id`, append([]any{businessID}, idArgs...)...)
	if err != nil {
		logger.Ctx(ctx).Error("insights carry-overs failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var cycleID int64
		var count int
		if err := rows.Scan(&cycleID, &count); err != nil {
			return nil, err
		}
		counts[cycleID] = count
	}
	return counts, rows.Err()
}

// insightsBurndown follows the active cycle with the most work in scope; ties
// go to the one started last.
// A project has at most one active cycle, so projectCount bounds the read.
func (r *Repository) insightsBurndown(ctx context.Context, businessID int64, placeholders string, scope func(...any) []any, now time.Time, projectCount int) ([]domain.InsightsBurndownPoint, error) {
	cycles, _, err := r.insightsCycles(ctx, businessID, placeholders, scope, domain.CycleStatusActive,
		"cy.start_date DESC, cy.id DESC", projectCount)
	if err != nil {
		return nil, err
	}
	var chosen *domain.Cycle
	for i := range cycles {
		if chosen == nil || cycles[i].Scope > chosen.Scope {
			chosen = &cycles[i]
		}
	}
	if chosen == nil {
		return []domain.InsightsBurndownPoint{}, nil
	}
	completions, err := r.cycleScopeCompletions(ctx, businessID, chosen.ID)
	if err != nil {
		return nil, err
	}
	return domain.CycleBurndown(chosen.StartDate, chosen.EndDate, now, completions), nil
}

// cycleScopeCompletions lists the cycle's scope (as in readCycleTotals: no
// archived or cancelled tasks) with each task's completion time, nil while it
// is not done.
func (r *Repository) cycleScopeCompletions(ctx context.Context, businessID, cycleID int64) ([]*time.Time, error) {
	rows, err := r.pool.QueryContext(ctx, `SELECT CASE WHEN `+taskBucketSQL+` = 'done' THEN t.completed_at END
  FROM flow_cycle_tasks ct
  JOIN flow_cycles cy ON cy.id = ct.cycle_id AND cy.business_id = ?
  JOIN flow_tasks t ON t.id = ct.task_id AND t.business_id = cy.business_id
  LEFT JOIN flow_statuses fs ON fs.id = t.status_id AND fs.project_id = t.project_id
 WHERE ct.cycle_id = ? AND COALESCE(t.archived, FALSE) = FALSE
   AND COALESCE(`+taskBucketSQL+`, '') <> 'cancelled'`, businessID, cycleID)
	if err != nil {
		logger.Ctx(ctx).Error("insights burndown tasks failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var completions []*time.Time
	for rows.Next() {
		var at sql.NullTime
		if err := rows.Scan(&at); err != nil {
			return nil, err
		}
		if at.Valid {
			completions = append(completions, &at.Time)
		} else {
			completions = append(completions, nil)
		}
	}
	return completions, rows.Err()
}

// insightsTimeLogMinutes sums Flow's own time logs per member over every task
// of the projects. Logs linked to an HR time entry are HR's time.
func (r *Repository) insightsTimeLogMinutes(ctx context.Context, placeholders string, scope func(...any) []any) ([]domain.InsightsMemberTime, error) {
	rows, err := r.pool.QueryContext(ctx, `SELECT tl.user_id, CAST(COALESCE(SUM(tl.duration_minutes), 0) AS SIGNED)
  FROM flow_time_logs tl
  JOIN flow_tasks t ON t.id = tl.task_id
 WHERE t.business_id = ? AND t.project_id IN (`+placeholders+`) AND tl.time_entry_id IS NULL
 GROUP BY tl.user_id`, scope()...)
	if err != nil {
		logger.Ctx(ctx).Error("insights time by member failed", logger.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	members := []domain.InsightsMemberTime{}
	for rows.Next() {
		var userID, minutes int64
		if err := rows.Scan(&userID, &minutes); err != nil {
			return nil, err
		}
		members = append(members, domain.InsightsMemberTime{ID: strconv.FormatInt(userID, 10), Minutes: minutes})
	}
	return members, rows.Err()
}

// ListTaskIDsInProjects pages through the ids of every task of the projects.
func (r *Repository) ListTaskIDsInProjects(ctx context.Context, businessID int64, projectIDs []int64, afterID int64, limit int32) ([]int64, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	placeholders, idArgs := int64Placeholders(projectIDs)
	args := append(append([]any{businessID}, idArgs...), afterID, limit)
	rows, err := r.pool.QueryContext(ctx, `SELECT id FROM flow_tasks
 WHERE business_id = ? AND project_id IN (`+placeholders+`) AND id > ?
 ORDER BY id ASC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
