package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
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

// ─────────────────────────────────────────────
// Cycles
// ─────────────────────────────────────────────

func (r *Repository) ListCycles(ctx context.Context, workspaceID, limit, offset int64) ([]domain.Cycle, int64, error) {
	total, err := r.queries.CountCycles(ctx, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Error("CountCycles failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListCycles(ctx, sqlc.ListCyclesParams{
		WorkspaceID: workspaceID,
		Limit:       int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:      int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListCycles failed", logger.Error(err))
		return nil, 0, err
	}

	cycles := make([]domain.Cycle, 0, len(rows))
	now := time.Now().UTC()
	for _, row := range rows {
		c := domain.Cycle{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			Name:        row.Name,
			Description: row.Description.String,
			Status:      string(row.Status.FlowCyclesStatus),
			StartDate:   row.StartDate.Time,
			EndDate:     row.EndDate.Time,
			Scope:       int(row.Scope),
			Completed:   int(row.Completed),
			CreatedBy:   row.CreatedBy.Int64,
			CreatedAt:   row.CreatedAt.Time,
		}
		if row.ProjectID.Valid {
			c.ProjectID = &row.ProjectID.Int64
		}
		if c.Status != "completed" {
			if now.Before(c.StartDate) {
				c.Status = "upcoming"
			} else if now.After(c.EndDate) {
				c.Status = "completed"
			} else {
				c.Status = "active"
			}
		}
		cycles = append(cycles, c)
	}
	return cycles, total, nil
}

func (r *Repository) CreateCycle(ctx context.Context, input domain.CycleInput) (int64, error) {
	res, err := r.queries.CreateCycle(ctx, sqlc.CreateCycleParams{
		OwnerID:     input.BusinessID,
		BusinessID:  input.BusinessID,
		WorkspaceID: input.WorkspaceID,
		ProjectID:   nullInt64PtrFromPtr(input.ProjectID),
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		StartDate:   sql.NullTime{Time: input.StartDate, Valid: true},
		EndDate:     sql.NullTime{Time: input.EndDate, Valid: true},
		CreatedBy:   sql.NullInt64{Int64: input.CreatedBy, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateCycle failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateCycle(ctx context.Context, input domain.UpdateCycleInput) error {
	parts := []string{"name = ?"}
	args := []interface{}{input.Name}

	if input.Status != "" {
		parts = append(parts, "status = ?")
		args = append(args, input.Status)
	}
	if !input.EndDate.IsZero() {
		parts = append(parts, "end_date = ?")
		args = append(args, input.EndDate.Format("2006-01-02"))
	}
	// Trailing binds: id, then the two placeholders inside workspaceMemberPredicate.
	args = append(args, input.ID, input.BusinessID, input.ActorID)

	res, err := r.pool.ExecContext(ctx,
		fmt.Sprintf("UPDATE flow_cycles SET %s WHERE id = ? AND %s", strings.Join(parts, ", "), workspaceMemberPredicate("flow_cycles")),
		args...,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateCycle failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func nullInt64PtrFromPtr(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

// ─────────────────────────────────────────────
// Milestones
// ─────────────────────────────────────────────

func (r *Repository) ListMilestones(ctx context.Context, workspaceID int64, projectID *int64, limit, offset int64) ([]domain.Milestone, int64, error) {
	if projectID != nil {
		total, countErr := r.queries.CountMilestonesByProject(ctx, sqlc.CountMilestonesByProjectParams{
			WorkspaceID: sql.NullInt64{Int64: workspaceID, Valid: true},
			ProjectID:   *projectID,
		})
		if countErr != nil {
			logger.Ctx(ctx).Error("CountMilestonesByProject failed", logger.Error(countErr))
			return nil, 0, countErr
		}
		projectRows, listErr := r.queries.ListMilestonesByProject(ctx, sqlc.ListMilestonesByProjectParams{
			WorkspaceID: sql.NullInt64{Int64: workspaceID, Valid: true},
			ProjectID:   *projectID,
			Limit:       int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			Offset:      int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		})
		if listErr != nil {
			logger.Ctx(ctx).Error("ListMilestones failed", logger.Error(listErr))
			return nil, 0, listErr
		}
		return mapMilestoneProjectRows(projectRows), total, nil
	}

	total, err := r.queries.CountMilestonesByWorkspace(ctx, sql.NullInt64{Int64: workspaceID, Valid: true})
	if err != nil {
		logger.Ctx(ctx).Error("CountMilestonesByWorkspace failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListMilestonesByWorkspace(ctx, sqlc.ListMilestonesByWorkspaceParams{
		WorkspaceID: sql.NullInt64{Int64: workspaceID, Valid: true},
		Limit:       int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:      int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListMilestones failed", logger.Error(err))
		return nil, 0, err
	}
	return mapMilestoneWorkspaceRows(rows), total, nil
}

func mapMilestoneWorkspaceRows(rows []sqlc.ListMilestonesByWorkspaceRow) []domain.Milestone {
	milestones := make([]domain.Milestone, 0, len(rows))
	now := time.Now().UTC()
	for _, row := range rows {
		m := domain.Milestone{
			ID:          row.ID,
			ProjectID:   row.ProjectID,
			WorkspaceID: row.WorkspaceID.Int64,
			Name:        row.Name,
			Description: row.Description.String,
			Status:      string(row.Status.FlowMilestonesStatus),
			DueDate:     row.DueDate.Time,
			CreatedBy:   row.CreatedBy.Int64,
			CreatedAt:   row.CreatedAt.Time,
			TaskCount:   int(row.TaskCount),
			DoneCount:   int(row.DoneCount),
		}
		if row.CompletedAt.Valid {
			m.CompletedAt = &row.CompletedAt.Time
		}
		if row.ProjectName.Valid {
			m.Project = &domain.MilestoneProject{Name: row.ProjectName.String, Color: row.ProjectColor.String}
		}
		if m.TaskCount > 0 {
			m.Progress = int(float64(m.DoneCount) / float64(m.TaskCount) * 100)
		}
		if m.Status != "completed" && m.DueDate.Before(now) {
			m.Status = "at_risk"
		}
		milestones = append(milestones, m)
	}
	return milestones
}

func mapMilestoneProjectRows(rows []sqlc.ListMilestonesByProjectRow) []domain.Milestone {
	milestones := make([]domain.Milestone, 0, len(rows))
	now := time.Now().UTC()
	for _, row := range rows {
		m := domain.Milestone{
			ID:          row.ID,
			ProjectID:   row.ProjectID,
			WorkspaceID: row.WorkspaceID.Int64,
			Name:        row.Name,
			Description: row.Description.String,
			Status:      string(row.Status.FlowMilestonesStatus),
			DueDate:     row.DueDate.Time,
			CreatedBy:   row.CreatedBy.Int64,
			CreatedAt:   row.CreatedAt.Time,
			TaskCount:   int(row.TaskCount),
			DoneCount:   int(row.DoneCount),
		}
		if row.CompletedAt.Valid {
			m.CompletedAt = &row.CompletedAt.Time
		}
		if row.ProjectName.Valid {
			m.Project = &domain.MilestoneProject{Name: row.ProjectName.String, Color: row.ProjectColor.String}
		}
		if m.TaskCount > 0 {
			m.Progress = int(float64(m.DoneCount) / float64(m.TaskCount) * 100)
		}
		if m.Status != "completed" && m.DueDate.Before(now) {
			m.Status = "at_risk"
		}
		milestones = append(milestones, m)
	}
	return milestones
}

func (r *Repository) CreateMilestone(ctx context.Context, input domain.MilestoneInput) (int64, error) {
	res, err := r.queries.CreateMilestone(ctx, sqlc.CreateMilestoneParams{
		OwnerID:     input.CreatedBy,
		BusinessID:  input.BusinessID,
		ProjectID:   input.ProjectID,
		WorkspaceID: sql.NullInt64{Int64: input.WorkspaceID, Valid: true},
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		DueDate:     sql.NullTime{Time: input.DueDate, Valid: true},
		CreatedBy:   sql.NullInt64{Int64: input.CreatedBy, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateMilestone failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateMilestone(ctx context.Context, input domain.UpdateMilestoneInput) error {
	parts := []string{"name = ?"}
	args := []interface{}{input.Name}

	if input.Description != "" {
		parts = append(parts, "description = ?")
		args = append(args, input.Description)
	}
	if input.Status != "" {
		parts = append(parts, "status = ?")
		args = append(args, input.Status)
		if input.Status == "completed" {
			parts = append(parts, "completed_at = NOW()")
		}
	}
	if !input.DueDate.IsZero() {
		parts = append(parts, "due_date = ?")
		args = append(args, input.DueDate.Format("2006-01-02"))
	}
	// Trailing binds: id, then the two placeholders inside workspaceMemberPredicate.
	args = append(args, input.ID, input.BusinessID, input.ActorID)

	res, err := r.pool.ExecContext(ctx,
		fmt.Sprintf("UPDATE flow_milestones SET %s WHERE id = ? AND %s", strings.Join(parts, ", "), workspaceMemberPredicate("flow_milestones")),
		args...,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateMilestone failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

// ─────────────────────────────────────────────
// Views
// ─────────────────────────────────────────────

// ListViews returns the shared views plus the caller's own private ones.
func (r *Repository) ListViews(ctx context.Context, businessID, workspaceID, callerID, limit, offset int64) ([]domain.View, int64, error) {
	total, err := r.queries.CountViewsByWorkspace(ctx, sqlc.CountViewsByWorkspaceParams{
		BusinessID:  businessID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CountViewsByWorkspace failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListViews(ctx, sqlc.ListViewsParams{
		BusinessID:  businessID,
		WorkspaceID: workspaceID,
		// Drives the private-view filter. Valid: true unconditionally, because a
		// NULL here would compare against created_by as NULL and hide every
		// private view from its own creator too.
		CreatedBy: sql.NullInt64{Int64: callerID, Valid: true},
		Limit:     int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:    int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListViews failed", logger.Error(err))
		return nil, 0, err
	}

	views := make([]domain.View, 0, len(rows))
	for _, row := range rows {
		sortDir := ""
		if row.SortDir.Valid {
			sortDir = string(row.SortDir.FlowViewsSortDir)
		}
		v := domain.View{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			Name:        row.Name,
			Icon:        row.Icon.String,
			GroupBy:     row.GroupBy.String,
			SortBy:      row.SortBy.String,
			SortDir:     sortDir,
			Visibility:  string(row.Visibility),
			CreatedBy:   row.CreatedBy.Int64,
			CreatedAt:   row.CreatedAt.Time,
		}
		if len(row.Filters) > 0 {
			if unmarshalErr := json.Unmarshal(row.Filters, &v.Filters); unmarshalErr != nil {
				logger.Ctx(ctx).Warn("failed to unmarshal view filters", logger.Error(unmarshalErr))
			}
		}
		if v.Filters == nil {
			v.Filters = map[string]interface{}{}
		}
		views = append(views, v)
	}
	return views, total, nil
}

// visibilityOrShared normalises an unset visibility to "shared". Sending "" to
// an ENUM column is rejected, and defaulting to private would hide views that
// used to be visible to everyone.
func visibilityOrShared(v string) string {
	if v == "" {
		return "shared"
	}
	return v
}

func (r *Repository) CreateView(ctx context.Context, input domain.ViewInput) (int64, error) {
	filtersJSON, err := json.Marshal(input.Filters)
	if err != nil {
		return 0, err
	}
	res, err := r.queries.CreateView(ctx, sqlc.CreateViewParams{
		OwnerID:     input.BusinessID,
		BusinessID:  input.BusinessID,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Icon:        sql.NullString{String: input.Icon, Valid: input.Icon != ""},
		Filters:     filtersJSON,
		GroupBy:     sql.NullString{String: input.GroupBy, Valid: input.GroupBy != ""},
		SortBy:      sql.NullString{String: input.SortBy, Valid: input.SortBy != ""},
		SortDir:     sqlc.NullFlowViewsSortDir{FlowViewsSortDir: sqlc.FlowViewsSortDir(input.SortDir), Valid: input.SortDir != ""},
		// Empty means shared, matching the column default and how every view
		// behaved before the field existed.
		Visibility: sqlc.FlowViewsVisibility(visibilityOrShared(input.Visibility)),
		CreatedBy:  sql.NullInt64{Int64: input.CreatedBy, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateView failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateView(ctx context.Context, input domain.UpdateViewInput) error {
	filtersJSON, err := json.Marshal(input.Filters)
	if err != nil {
		return err
	}
	res, err := r.queries.UpdateView(ctx, sqlc.UpdateViewParams{
		Name:       input.Name,
		Icon:       sql.NullString{String: input.Icon, Valid: input.Icon != ""},
		Filters:    filtersJSON,
		GroupBy:    sql.NullString{String: input.GroupBy, Valid: input.GroupBy != ""},
		SortBy:     sql.NullString{String: input.SortBy, Valid: input.SortBy != ""},
		SortDir:    sqlc.NullFlowViewsSortDir{FlowViewsSortDir: sqlc.FlowViewsSortDir(input.SortDir), Valid: input.SortDir != ""},
		ID:         input.ID,
		BusinessID: input.BusinessID,
		OwnerID:    input.ActorID,
		UserID:     input.ActorID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateView failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) DeleteView(ctx context.Context, id, actorID, businessID int64) error {
	res, err := r.queries.DeleteView(ctx, sqlc.DeleteViewParams{
		ID:         id,
		BusinessID: businessID,
		OwnerID:    actorID,
		UserID:     actorID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteView failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func workspaceMemberPredicate(table string) string {
	return fmt.Sprintf(`(
		EXISTS (SELECT 1 FROM flow_workspaces w WHERE w.id = %s.workspace_id AND w.owner_id = ?)
		OR EXISTS (SELECT 1 FROM flow_workspace_members wm WHERE wm.workspace_id = %s.workspace_id AND wm.user_id = ?)
	)`, table, table)
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
	if avgDays != nil {
		if f, ok := avgDays.(float64); ok {
			ins.CycleTime = math.Round(f*10) / 10
		}
	}

	velRows, err := r.queries.GetInsightsVelocity(ctx, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Warn("GetInsights velocity failed", logger.Error(err))
	} else {
		for _, row := range velRows {
			completed := int(row.Completed)
			ins.Velocity = append(ins.Velocity, domain.InsightsVelocityPoint{
				Cycle:          row.WeekStart,
				Completed:      completed,
				Total:          completed + int(float64(completed)*0.3),
				CompletionRate: math.Min(1.0, float64(completed)/math.Max(1, float64(completed)+float64(completed)*0.3)),
			})
		}
	}

	ins.Throughput = make([]domain.InsightsThroughput, len(ins.Velocity))
	for i, v := range ins.Velocity {
		ins.Throughput[i] = domain.InsightsThroughput{Week: v.Cycle, Completed: v.Completed}
	}

	totalOpen, err := r.queries.GetInsightsOpenTaskCount(ctx, workspaceID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("GetInsights open task count failed", logger.Error(err))
	}
	idealPerDay := float64(totalOpen) / 14.0
	ins.Burndown = make([]domain.InsightsBurndownPoint, 15)
	for d := 0; d <= 14; d++ {
		remaining := math.Max(0, float64(totalOpen)-float64(d)*idealPerDay*1.15)
		ins.Burndown[d] = domain.InsightsBurndownPoint{
			Day:    d,
			Ideal:  math.Round((float64(totalOpen)-idealPerDay*float64(d))*10) / 10,
			Actual: math.Round(remaining*10) / 10,
		}
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

func mapWorkloadMemberTasks(rows []sqlc.GetTeamWorkloadMemberTasksRow) []domain.Task {
	tasks := make([]domain.Task, 0, len(rows))
	for _, row := range rows {
		t := domain.Task{
			ID:          row.ID,
			ProjectID:   row.ProjectID,
			WorkspaceID: row.WorkspaceID,
			OwnerID:     row.OwnerID,
			Title:       row.Title,
			Description: row.Description.String,
			Status:      string(row.Status.FlowTasksStatus),
			Priority:    string(row.Priority.FlowTasksPriority),
			ReporterID:  row.ReporterID,
			SortOrder:   int(row.SortOrder.Int32),
			Archived:    row.Archived.Bool,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
			Identifier:  "T-" + strconv.FormatInt(row.ID, 10),
		}
		if row.ParentTaskID.Valid {
			t.ParentTaskID = &row.ParentTaskID.Int64
		}
		if row.SectionID.Valid {
			t.SectionID = &row.SectionID.Int64
		}
		if row.AssigneeID.Valid {
			t.AssigneeID = &row.AssigneeID.Int64
		}
		if row.DueDate.Valid {
			t.DueDate = &row.DueDate.Time
		}
		if row.StartDate.Valid {
			t.StartDate = &row.StartDate.Time
		}
		if row.CompletedAt.Valid {
			t.CompletedAt = &row.CompletedAt.Time
		}
		if row.EstimatedHours.Valid {
			if f, parseErr := strconv.ParseFloat(row.EstimatedHours.String, 64); parseErr == nil {
				t.EstimatedHours = &f
			}
		}
		if row.ActualHours.Valid {
			if f, parseErr := strconv.ParseFloat(row.ActualHours.String, 64); parseErr == nil {
				t.ActualHours = &f
			}
		}
		if len(row.Labels) > 0 {
			if err := json.Unmarshal(row.Labels, &t.Labels); err != nil {
				logger.Global().Warn("failed to parse task labels JSON from database",
					logger.Any("task_id", row.ID), logger.Error(err))
			}
		}
		tasks = append(tasks, t)
	}
	return tasks
}

// ─────────────────────────────────────────────
// Roadmap (derived from projects + milestones)
// ─────────────────────────────────────────────

func (r *Repository) GetRoadmap(ctx context.Context, workspaceID, limit, offset int64) ([]domain.Project, int64, error) {
	total, err := r.queries.CountRoadmapProjects(ctx, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Error("CountRoadmapProjects failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.GetRoadmap(ctx, sqlc.GetRoadmapParams{
		WorkspaceID: workspaceID,
		Limit:       int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:      int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetRoadmap failed", logger.Error(err))
		return nil, 0, err
	}

	projects := make([]domain.Project, 0, len(rows))
	for _, row := range rows {
		p := domain.Project{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			OwnerID:     row.OwnerID,
			Name:        row.Name,
			Description: row.Description.String,
			Icon:        row.Icon.String,
			Color:       row.Color.String,
			Status:      string(row.Status.FlowProjectsStatus),
			CreatedBy:   row.CreatedBy,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		}
		if row.StartDate.Valid {
			p.StartDate = &row.StartDate.Time
		}
		if row.DueDate.Valid {
			p.DueDate = &row.DueDate.Time
		}
		projects = append(projects, p)
	}
	return projects, total, nil
}

// ─────────────────────────────────────────────
// Notifications
// ─────────────────────────────────────────────

func (r *Repository) ListNotifications(ctx context.Context, userID, limit, offset int64) ([]domain.Notification, int64, error) {
	total, err := r.queries.CountNotifications(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("CountNotifications failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListNotifications(ctx, sqlc.ListNotificationsParams{
		UserID: userID,
		Limit:  int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset: int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListNotifications failed", logger.Error(err))
		return nil, 0, err
	}

	notifs := make([]domain.Notification, 0, len(rows))
	for _, row := range rows {
		n := domain.Notification{
			ID:            row.ID,
			UserID:        row.UserID,
			Type:          string(row.Type),
			Title:         row.Title,
			Description:   row.Description,
			TaskTitle:     row.TaskTitle,
			ActorInitials: row.ActorInitials,
			Project:       row.Project,
			ProjectColor:  row.ProjectColor,
			IsRead:        row.IsRead.Bool,
			CreatedAt:     row.CreatedAt,
		}
		if row.WorkspaceID.Valid {
			n.WorkspaceID = &row.WorkspaceID.Int64
		}
		if row.TaskID.Valid {
			n.TaskID = &row.TaskID.Int64
		}
		if row.ActorID.Valid {
			n.ActorID = &row.ActorID.Int64
		}
		notifs = append(notifs, n)
	}
	return notifs, total, nil
}

func (r *Repository) CreateNotification(ctx context.Context, input domain.CreateNotificationInput) error {
	err := r.queries.CreateNotification(ctx, sqlc.CreateNotificationParams{
		UserID:        input.UserID,
		WorkspaceID:   nullInt64PtrFromPtr(input.WorkspaceID),
		Type:          sqlc.FlowNotificationsType(input.Type),
		Title:         input.Title,
		Description:   sql.NullString{String: input.Description, Valid: input.Description != ""},
		TaskID:        nullInt64PtrFromPtr(input.TaskID),
		TaskTitle:     sql.NullString{String: input.TaskTitle, Valid: input.TaskTitle != ""},
		ActorID:       nullInt64PtrFromPtr(input.ActorID),
		ActorInitials: sql.NullString{String: input.ActorInitials, Valid: input.ActorInitials != ""},
		Project:       sql.NullString{String: input.Project, Valid: input.Project != ""},
		ProjectColor:  sql.NullString{String: input.ProjectColor, Valid: input.ProjectColor != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateNotification failed", logger.Error(err))
	}
	return err
}

func (r *Repository) MarkNotificationRead(ctx context.Context, id, userID int64) error {
	return r.queries.MarkNotificationRead(ctx, sqlc.MarkNotificationReadParams{ID: id, UserID: userID})
}

func (r *Repository) MarkAllNotificationsRead(ctx context.Context, userID int64) error {
	return r.queries.MarkAllNotificationsRead(ctx, userID)
}

// ─────────────────────────────────────────────
// User Settings
// ─────────────────────────────────────────────

func (r *Repository) GetUserSettings(ctx context.Context, userID int64) (domain.UserSettings, error) {
	s := domain.UserSettings{
		UserID:       userID,
		TaskAssigned: true,
		TaskComments: true,
		Mentions:     true,
		WeeklyDigest: true,
	}

	row, err := r.queries.GetUserSettings(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return s, nil
	}
	if err != nil {
		return domain.UserSettings{}, err
	}
	s.TaskAssigned = row.TaskAssigned.Bool
	s.TaskComments = row.TaskComments.Bool
	s.Mentions = row.Mentions.Bool
	s.StatusChanges = row.StatusChanges.Bool
	s.WeeklyDigest = row.WeeklyDigest.Bool
	s.EmailNotifications = row.EmailNotifications.Bool
	s.SidebarPreferencesJSON = row.SidebarPreferencesJson.String
	return s, nil
}

func (r *Repository) UpsertUserSettings(ctx context.Context, s domain.UserSettings) error {
	err := r.queries.UpsertUserSettings(ctx, sqlc.UpsertUserSettingsParams{
		UserID:                 s.UserID,
		TaskAssigned:           sql.NullBool{Bool: s.TaskAssigned, Valid: true},
		TaskComments:           sql.NullBool{Bool: s.TaskComments, Valid: true},
		Mentions:               sql.NullBool{Bool: s.Mentions, Valid: true},
		StatusChanges:          sql.NullBool{Bool: s.StatusChanges, Valid: true},
		WeeklyDigest:           sql.NullBool{Bool: s.WeeklyDigest, Valid: true},
		EmailNotifications:     sql.NullBool{Bool: s.EmailNotifications, Valid: true},
		SidebarPreferencesJson: sql.NullString{String: s.SidebarPreferencesJSON, Valid: s.SidebarPreferencesJSON != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpsertUserSettings failed", logger.Error(err))
	}
	return err
}

// ─────────────────────────────────────────────
// Roadmap task/progress helpers
// ─────────────────────────────────────────────

func (r *Repository) GetProjectTaskCounts(ctx context.Context, projectID int64) (total, done int) {
	row, err := r.queries.GetProjectTaskCounts(ctx, projectID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Warn("GetProjectTaskCounts failed", logger.Error(err))
		return 0, 0
	}
	return int(row.Total), int(countIfaceToInt64(row.Done))
}

func (r *Repository) GetProjectTaskCountsByProjectIDs(ctx context.Context, projectIDs []int64) (map[int64]domain.ProjectTaskCounts, error) {
	counts := make(map[int64]domain.ProjectTaskCounts, len(projectIDs))
	if len(projectIDs) == 0 {
		return counts, nil
	}

	rows, err := r.queries.GetProjectTaskCountsByProjectIDs(ctx, projectIDs)
	if err != nil {
		logger.Ctx(ctx).Warn("GetProjectTaskCountsByProjectIDs failed", logger.Error(err))
		return nil, err
	}
	for _, row := range rows {
		counts[row.ProjectID] = domain.ProjectTaskCounts{
			Total: int(row.Total),
			Done:  int(countIfaceToInt64(row.Done)),
		}
	}
	return counts, nil
}

func countIfaceToInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return n
	case int32:
		return int64(n)
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case []byte:
		i, err := strconv.ParseInt(string(n), 10, 64)
		if err != nil {
			return 0
		}
		return i
	default:
		return 0
	}
}

func countIfaceToFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int64:
		return float64(n)
	case []byte:
		f, err := strconv.ParseFloat(string(n), 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

func ifaceToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return fmt.Sprint(v)
	}
}
