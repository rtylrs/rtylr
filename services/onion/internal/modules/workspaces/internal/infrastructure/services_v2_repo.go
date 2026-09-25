package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	sqlc "github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/times"
)

func (r *Repository) GetDashboardStatsV2(ctx context.Context, businessID int64) (domain.DashboardStats, error) {
	weekStart, _ := times.GetWeekRangeStartAndEnd()
	weekEnd := weekStart.AddDate(0, 0, 7)
	row, err := r.queries.GetDashboardStatsV2(ctx, sqlc.GetDashboardStatsV2Params{
		CompletedAt:   sql.NullTime{Time: weekStart, Valid: true},
		CompletedAt_2: sql.NullTime{Time: weekEnd, Valid: true},
		BusinessID:    businessID,
	})
	if err != nil {
		return domain.DashboardStats{}, err
	}
	return domain.DashboardStats{
		OpenTasks:         row.OpenTasks,
		InProgressTasks:   row.InProgressTasks,
		CompletedThisWeek: row.CompletedThisWeek,
		OverdueTasks:      row.OverdueTasks,
	}, nil
}

func (r *Repository) ListLabelsV2(ctx context.Context, businessID, limit, offset int64) ([]domain.Label, int64, error) {
	total, err := r.queries.CountLabelsV2(ctx, businessID)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queries.ListLabelsV2(ctx, sqlc.ListLabelsV2Params{BusinessID: businessID, Limit: int32(limit), Offset: int32(offset)}) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	if err != nil {
		return nil, 0, err
	}
	labels := make([]domain.Label, 0, len(rows))
	for _, l := range rows {
		labels = append(labels, domain.Label{ID: l.ID, Name: l.Name, Color: l.Color, CreatedAt: l.CreatedAt.Time})
	}
	return labels, total, nil
}

func (r *Repository) CreateLabelV2(ctx context.Context, businessID int64, name, color string) (int64, error) {
	res, err := r.queries.CreateLabelV2(ctx, sqlc.CreateLabelV2Params{BusinessID: businessID, Name: name, Color: color})
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListCyclesV2 applies the same project-visibility rule as task listing: a
// caller sees a project's cycles only if they created it, own it, or are a
// member. Without it any member of the business could list the cycles of a
// private project by passing its id, which leaked names and date ranges.
// includeAll is the hub-role bypass.
func (r *Repository) ListCyclesV2(ctx context.Context, projectID, businessID, callerID int64, includeAll bool, limit, offset int64) ([]domain.Cycle, int64, error) {
	bypass := 0
	if includeAll {
		bypass = 1
	}
	total, err := r.queries.CountCyclesV2(ctx, sqlc.CountCyclesV2Params{
		ProjectID: nullInt64(projectID), BusinessID: businessID,
		Column3: bypass, CreatedBy: callerID, OwnerID: callerID, UserID: callerID,
	})
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.queries.ListCyclesV2(ctx, sqlc.ListCyclesV2Params{
		ProjectID: nullInt64(projectID), BusinessID: businessID,
		Column3: bypass, CreatedBy: callerID, OwnerID: callerID, UserID: callerID,
		Limit: int32(limit), Offset: int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		return nil, 0, err
	}
	cycles := make([]domain.Cycle, 0, len(rows))
	for _, c := range rows {
		cycles = append(cycles, domain.Cycle{
			ID: c.ID, Name: c.Name, Status: string(c.Status.FlowCyclesStatus),
			Scope: int(c.Scope), Completed: int(c.Completed),
		})
	}
	return cycles, total, nil
}

func (r *Repository) CreateCycleV2(ctx context.Context, input domain.CycleInput) (int64, error) {
	res, err := r.queries.CreateCycleV2(ctx, sqlc.CreateCycleV2Params{
		BusinessID: input.BusinessID,
		ProjectID:  nullInt64Ptr(input.ProjectID),
		Name:       input.Name,
		StartDate:  sql.NullTime{Time: input.StartDate, Valid: !input.StartDate.IsZero()},
		EndDate:    sql.NullTime{Time: input.EndDate, Valid: !input.EndDate.IsZero()},
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateCycleV2 failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

// ListMilestonesV2 carries the same visibility rule as ListCyclesV2, for the
// same reason: a milestone name and target date describe a project, so a caller
// who cannot see the project must not see them.
func (r *Repository) ListMilestonesV2(ctx context.Context, businessID int64, projectID *int64, callerID int64, includeAll bool, limit, offset int64) ([]domain.Milestone, int64, error) {
	pid := int64(0)
	if projectID != nil {
		pid = *projectID
	}
	bypass := 0
	if includeAll {
		bypass = 1
	}
	rows, err := r.queries.ListMilestonesByBusiness(ctx, sqlc.ListMilestonesByBusinessParams{
		BusinessID: businessID, Column2: pid, ProjectID: pid,
		Column4: bypass, CreatedBy: callerID, OwnerID: callerID, UserID: callerID,
		Limit: int32(limit), Offset: int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		return nil, 0, err
	}
	milestones := make([]domain.Milestone, 0, len(rows))
	for _, m := range rows {
		milestone := domain.Milestone{
			ID: m.ID, ProjectID: m.ProjectID, Name: m.Name,
			Status: string(m.Status), DoneCount: int(m.DoneCount), TaskCount: int(m.TaskCount),
		}
		if m.TargetDate.Valid {
			milestone.DueDate = m.TargetDate.Time
		}
		// The list joins the project for its name and color; without them every
		// milestone card rendered a blank project line.
		if m.ProjectName.Valid {
			milestone.Project = &domain.MilestoneProject{Name: m.ProjectName.String, Color: m.ProjectColor.String}
		}
		milestones = append(milestones, milestone)
	}
	return milestones, int64(len(rows)), nil
}

func (r *Repository) CreateMilestoneV2(ctx context.Context, input domain.MilestoneInput) (int64, error) {
	res, err := r.queries.CreateMilestoneV2(ctx, sqlc.CreateMilestoneV2Params{
		BusinessID: input.BusinessID, ProjectID: input.ProjectID,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		TargetDate:  sql.NullTime{Time: input.DueDate, Valid: !input.DueDate.IsZero()},
	})
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// callerID drives the private-view filter: a private view resolves only for
// whoever created it. Valid: true unconditionally, because a NULL would compare
// against created_by as NULL and hide every private view from its own creator.
func (r *Repository) ListViewsV2(ctx context.Context, businessID, callerID, limit, offset int64) ([]domain.View, int64, error) {
	rows, err := r.queries.ListViewsV2(ctx, sqlc.ListViewsV2Params{
		BusinessID: businessID,
		CreatedBy:  sql.NullInt64{Int64: callerID, Valid: true},
		Limit:      int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		return nil, 0, err
	}
	views := make([]domain.View, 0, len(rows))
	for _, v := range rows {
		var filters map[string]interface{}
		if err := json.Unmarshal(v.Filters, &filters); err != nil {
			logger.Ctx(ctx).Warn("ListViewsV2: failed to unmarshal view filters", logger.Error(err))
		}
		sortDir := string(v.SortDir)
		views = append(views, domain.View{
			ID: v.ID, Name: v.Name, Icon: v.Icon,
			Filters: filters, GroupBy: v.GroupBy, SortBy: v.SortBy, SortDir: sortDir,
			Visibility: string(v.Visibility),
		})
	}
	return views, int64(len(rows)), nil
}

func (r *Repository) CreateViewV2(ctx context.Context, input domain.ViewInput) (int64, error) {
	filters, err := filtersJSON(input.Filters)
	if err != nil {
		return 0, fmt.Errorf("encode view filters: %w", err)
	}
	res, err := r.queries.CreateViewV2(ctx, sqlc.CreateViewV2Params{
		BusinessID: input.BusinessID, Name: input.Name,
		Icon:    sql.NullString{String: input.Icon, Valid: input.Icon != ""},
		Filters: filters,
		GroupBy: sql.NullString{String: input.GroupBy, Valid: input.GroupBy != ""},
		SortBy:  sql.NullString{String: input.SortBy, Valid: input.SortBy != ""},
		SortDir: sqlc.NullFlowViewsSortDir{FlowViewsSortDir: sqlc.FlowViewsSortDir(input.SortDir), Valid: input.SortDir != ""},
		// Empty normalises to shared, matching the column default. An empty string
		// on the ENUM would be rejected, and defaulting to private would hide a
		// view its creator expected to share.
		Visibility: sqlc.FlowViewsVisibility(visibilityOrShared(input.Visibility)),
		CreatedBy:  nullInt64Ptr(&input.CreatedBy),
	})
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func filtersJSON(m map[string]interface{}) (json.RawMessage, error) {
	return json.Marshal(m)
}

// BusinessEmployeeIDs returns every HR employee id for a business.
func (r *Repository) BusinessEmployeeIDs(ctx context.Context, businessID int64) ([]int64, error) {
	rows, err := r.pool.QueryContext(ctx,
		"SELECT e.id FROM hr_employees e WHERE e.business_id = ? AND e.status = 'active'", businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

// EmployeeUserIDs resolves HR employee ids to the auth user ids they belong to,
// for the work-day team view. Work days are keyed on hr_employees.id while every
// roster the client holds is keyed on users.id, and the two are different id
// spaces, so without this join a day cannot be attached to a person.
//
// hr_employees.user_id is nullable. An employee with no user is simply absent
// from the result: there is no account to attribute their day to, which is a
// real state rather than an error.
func (r *Repository) EmployeeUserIDs(ctx context.Context, businessID int64, employeeIDs []int64) (map[int64]int64, error) {
	if len(employeeIDs) == 0 {
		return map[int64]int64{}, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(employeeIDs)), ",")
	args := make([]any, 0, len(employeeIDs)+1)
	args = append(args, businessID)
	for _, id := range employeeIDs {
		args = append(args, id)
	}

	rows, err := r.pool.QueryContext(ctx,
		"SELECT e.id, e.user_id FROM hr_employees e WHERE e.business_id = ? AND e.id IN ("+placeholders+")",
		args...,
	)
	if err != nil {
		logger.Ctx(ctx).Error("EmployeeUserIDs failed", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	byEmployeeID := make(map[int64]int64, len(employeeIDs))
	for rows.Next() {
		var employeeID int64
		var userID sql.NullInt64
		if scanErr := rows.Scan(&employeeID, &userID); scanErr != nil {
			return nil, scanErr
		}
		if userID.Valid {
			byEmployeeID[employeeID] = userID.Int64
		}
	}
	return byEmployeeID, rows.Err()
}

// GetProjectsByCreatorOrMembership lists projects where the caller is the
// creator or a member (the curation rule from the v2 project listing).
func (r *Repository) GetProjectsByCreatorOrMembership(ctx context.Context, businessID, callerID, limit, offset int64) ([]domain.Project, int64, error) {
	// nil managed set on purpose: this helper means creator-or-member exactly,
	// as its name says, so the dept-manager proxy must not widen it.
	return r.ListProjectsV2(ctx, businessID, callerID, false, nil, "", "", "", limit, offset)
}
