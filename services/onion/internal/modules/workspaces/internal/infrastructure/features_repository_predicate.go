package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// memberPredicate authorizes a write to a project-scoped table by the acting
// user's membership of the project or, for legacy rows, the workspace.
//
// Both halves are needed. A v2 cycle carries workspace_id 0 and a v2 milestone
// carries NULL (v2 no longer creates workspaces), so the workspace branches
// alone could never match and requireAffected turned every v2 update into
// domain.ErrUnauthorized - the same trap SetTaskLabels documents in labels.sql.
// The workspace branches stay so a v1 row keeps resolving exactly as before.
//
// Five placeholders, in order: project member, project owner, project creator,
// workspace owner, workspace member - each bound to the acting user.
func memberPredicate(table string) string {
	return fmt.Sprintf(`(
		EXISTS (SELECT 1 FROM flow_project_members pm WHERE pm.project_id = %s.project_id AND pm.user_id = ?)
		OR EXISTS (SELECT 1 FROM flow_projects p WHERE p.id = %s.project_id AND (p.owner_id = ? OR p.created_by = ?))
		OR EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = %s.workspace_id AND w.owner_id = ?)
		OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = %s.workspace_id AND wm.user_id = ?)
	)`, table, table, table, table)
}

func requireAffected(res sql.Result) error {
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrUnauthorized
	}
	return nil
}

// ─────────────────────────────────────────────
// Insights (computed aggregations)
// ─────────────────────────────────────────────

func (r *Repository) GetInsights(ctx context.Context, workspaceID int64) (domain.Insights, error) {
	var ins domain.Insights

	counts, err := r.queries.GetInsightsTaskCounts(ctx, workspaceID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("GetInsights task counts failed", logger.Error(err))
		return domain.Insights{}, err
	}
	ins.TotalTasks = int(counts.TotalTasks)
	ins.DoneTasks = int(countIfaceToInt64(counts.DoneTasks))

	avgDays, err := r.queries.GetInsightsCycleTime(ctx, workspaceID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("GetInsights cycle time failed", logger.Error(err))
	}
	// MySQL reports AVG(DATEDIFF(...)) as DECIMAL and the driver hands that back
	// as []byte, so a direct float64 assertion never matched and every
	// workspace reported cycleTime 0. countIfaceToFloat64 is this file's
	// converter for the driver's wire types; DoneTasks above uses it too.
	if avgDays != nil {
		ins.CycleTime = math.Round(countIfaceToFloat64(avgDays)*10) / 10
	}

	velRows, err := r.queries.GetInsightsVelocity(ctx, sqlc.GetInsightsVelocityParams{
		WorkspaceID: workspaceID, WorkspaceID_2: workspaceID,
	})
	if err != nil {
		logger.Ctx(ctx).Warn("GetInsights velocity failed", logger.Error(err))
	} else {
		for _, row := range velRows {
			completed := int(row.Completed)
			total := int(row.Total)
			rate := 0.0
			if total > 0 {
				rate = float64(completed) / float64(total)
			}
			ins.Velocity = append(ins.Velocity, domain.InsightsVelocityPoint{
				Cycle:          row.WeekStart,
				Completed:      completed,
				Total:          total,
				CompletionRate: rate,
			})
		}
	}

	ins.Throughput = make([]domain.InsightsThroughput, len(ins.Velocity))
	for i, v := range ins.Velocity {
		ins.Throughput[i] = domain.InsightsThroughput{Week: v.Cycle, Completed: v.Completed}
	}

	// Burndown is measured from the task history rather than extrapolated: the
	// remaining work at the end of a day is whatever was already in flight when
	// the window opened, plus everything created up to that day, minus
	// everything completed up to it. The previous curve was a synthetic line
	// derived from the open-task count.
	activity, err := r.queries.GetInsightsBurndownDaily(ctx, sqlc.GetInsightsBurndownDailyParams{
		WorkspaceID: workspaceID, WorkspaceID_2: workspaceID,
	})
	if err != nil {
		logger.Ctx(ctx).Warn("GetInsights burndown activity failed", logger.Error(err))
	}
	createdByDay := make(map[string]int, len(activity))
	completedByDay := make(map[string]int, len(activity))
	for _, a := range activity {
		day := a.Day.Format("2006-01-02")
		createdByDay[day] += int(a.Created)
		completedByDay[day] += int(a.Completed)
	}
	baseOpen, err := r.queries.GetInsightsBurndownBase(ctx, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Warn("GetInsights burndown base failed", logger.Error(err))
	}

	now := time.Now().UTC()
	windowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -14)
	ins.Burndown = make([]domain.InsightsBurndownPoint, 15)
	remaining := int(baseOpen)
	for d := 0; d <= 14; d++ {
		day := windowStart.AddDate(0, 0, d).Format("2006-01-02")
		remaining += createdByDay[day] - completedByDay[day]
		if remaining < 0 {
			remaining = 0
		}
		ins.Burndown[d].Day = d
		ins.Burndown[d].Actual = float64(remaining)
	}
	// The ideal line burns the same starting scope down to zero across the
	// window, so the two lines share an origin.
	scope := ins.Burndown[0].Actual
	for d := 0; d <= 14; d++ {
		ins.Burndown[d].Ideal = math.Round(scope*(float64(14-d)/14)*10) / 10
	}

	tlRows, err := r.queries.GetInsightsTimeByMember(ctx, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Warn("GetInsights time by member failed", logger.Error(err))
	} else {
		for _, row := range tlRows {
			hours := countIfaceToFloat64(row.Hours)
			ins.TimeByMember = append(ins.TimeByMember, domain.InsightsMemberTime{
				ID:    ifaceToString(row.MemberID),
				Name:  row.Name,
				Hours: math.Round(hours*10) / 10,
			})
		}
	}
	if ins.TimeByMember == nil {
		ins.TimeByMember = []domain.InsightsMemberTime{}
	}
	if ins.Velocity == nil {
		ins.Velocity = []domain.InsightsVelocityPoint{}
	}
	if ins.Throughput == nil {
		ins.Throughput = []domain.InsightsThroughput{}
	}

	return ins, nil
}

// ─────────────────────────────────────────────
// Team Workload
// ─────────────────────────────────────────────

func (r *Repository) GetTeamWorkload(ctx context.Context, workspaceID, limit, offset int64) ([]domain.TeamMemberWorkload, int64, error) {
	total, err := r.queries.CountWorkspaceMembersForWorkload(ctx, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Error("CountWorkspaceMembersForWorkload failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.GetTeamWorkload(ctx, sqlc.GetTeamWorkloadParams{
		WorkspaceID:   workspaceID,
		WorkspaceID_2: workspaceID,
		Limit:         int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:        int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetTeamWorkload failed", logger.Error(err))
		return nil, 0, err
	}

	workloads := make([]domain.TeamMemberWorkload, 0, len(rows))
	for _, row := range rows {
		w := domain.TeamMemberWorkload{
			ID:   ifaceToString(row.MemberID),
			Name: row.Name,
		}
		parts := strings.Fields(w.Name)
		if len(parts) >= 2 {
			w.Initials = strings.ToUpper(string(parts[0][0])) + strings.ToUpper(string(parts[len(parts)-1][0]))
		} else if len(w.Name) > 0 {
			w.Initials = strings.ToUpper(string(w.Name[0]))
		}
		w.Capacity = 8
		w.AssignedHours = float64(row.TaskCount) * 5
		w.Load = math.Min(1.0, float64(row.TaskCount)/float64(w.Capacity))
		workloads = append(workloads, w)
	}

	for i := range workloads {
		memberID, parseErr := strconv.ParseInt(workloads[i].ID, 10, 64)
		if parseErr != nil {
			logger.Ctx(ctx).Warn("GetTeamWorkload member id parse failed", logger.Error(parseErr))
			workloads[i].Tasks = []domain.Task{}
			continue
		}
		taskRows, taskErr := r.queries.GetTeamWorkloadMemberTasks(ctx, sqlc.GetTeamWorkloadMemberTasksParams{
			WorkspaceID: workspaceID,
			AssigneeID:  sql.NullInt64{Int64: memberID, Valid: true},
		})
		if taskErr != nil {
			logger.Ctx(ctx).Warn("GetTeamWorkload member tasks failed", logger.Error(taskErr))
			workloads[i].Tasks = []domain.Task{}
			continue
		}
		workloads[i].Tasks = mapWorkloadMemberTasks(taskRows)
	}

	return workloads, total, nil
}
