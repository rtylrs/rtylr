package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/flow/insights/internal/domain"
	"github.com/voxire/onion/internal/modules/flow/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// ListDashboardProjects returns the projects userID may open a project
// dashboard for in businessID: public ones and the private ones they belong to.
func (r *Repository) ListDashboardProjects(ctx context.Context, businessID, userID int64) ([]domain.DashboardProject, error) {
	rows, err := r.queries.ListDashboardProjectsForUser(ctx, sqlc.ListDashboardProjectsForUserParams{BusinessID: businessID, UserID: userID})
	if err != nil {
		logger.Ctx(ctx).Error("ListDashboardProjects failed", logger.Error(err))
		return nil, err
	}
	out := make([]domain.DashboardProject, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.DashboardProject{ID: row.ID, Name: row.Name, Status: string(row.Status), Visibility: string(row.Visibility)})
	}
	return out, nil
}

// GetDashboardProject returns one project under the ListDashboardProjects
// rule. found is false when the project does not exist, belongs to another
// business, is a personal My Tasks project, or is private and userID is not
// a member: the caller cannot tell these apart.
func (r *Repository) GetDashboardProject(ctx context.Context, businessID, userID, projectID int64) (project domain.DashboardProject, found bool, err error) {
	row, err := r.queries.GetDashboardProjectForUser(ctx, sqlc.GetDashboardProjectForUserParams{ProjectID: projectID, BusinessID: businessID, UserID: userID})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DashboardProject{}, false, nil
	}
	if err != nil {
		logger.Ctx(ctx).Error("GetDashboardProject failed", logger.Error(err))
		return domain.DashboardProject{}, false, err
	}
	return domain.DashboardProject{ID: row.ID, Name: row.Name, Status: string(row.Status), Visibility: string(row.Visibility)}, true, nil
}

// GetDashboardProjectByID returns one project with no viewer rule, for a
// project Flow's access check has already granted. found is false for a
// missing project, another business's project and a personal project alike.
func (r *Repository) GetDashboardProjectByID(ctx context.Context, businessID, projectID int64) (project domain.DashboardProject, found bool, err error) {
	row, err := r.queries.GetDashboardProjectByID(ctx, sqlc.GetDashboardProjectByIDParams{ProjectID: projectID, BusinessID: businessID})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DashboardProject{}, false, nil
	}
	if err != nil {
		logger.Ctx(ctx).Error("GetDashboardProjectByID failed", logger.Error(err))
		return domain.DashboardProject{}, false, err
	}
	return domain.DashboardProject{ID: row.ID, Name: row.Name, Status: string(row.Status), Visibility: string(row.Visibility)}, true, nil
}

// ListDashboardProjectsByIDs returns the active, non-personal projects among
// projectIDs in businessID. An empty projectIDs reads nothing.
func (r *Repository) ListDashboardProjectsByIDs(ctx context.Context, businessID int64, projectIDs []int64) ([]domain.DashboardProject, error) {
	out := []domain.DashboardProject{}
	if len(projectIDs) == 0 {
		return out, nil
	}
	rows, err := r.queries.ListDashboardProjectsByIDs(ctx, sqlc.ListDashboardProjectsByIDsParams{BusinessID: businessID, ProjectIds: projectIDs})
	if err != nil {
		logger.Ctx(ctx).Error("ListDashboardProjectsByIDs failed", logger.Error(err))
		return nil, err
	}
	for _, row := range rows {
		out = append(out, domain.DashboardProject{ID: row.ID, Name: row.Name, Status: string(row.Status), Visibility: string(row.Visibility)})
	}
	return out, nil
}

// GetProjectTaskProgress returns one project's task counts and its oldest
// overdue tasks.
func (r *Repository) GetProjectTaskProgress(ctx context.Context, businessID, projectID int64) (domain.ProjectTaskProgress, error) {
	counts, err := r.queries.GetProjectTaskProgress(ctx, sqlc.GetProjectTaskProgressParams{BusinessID: businessID, ProjectID: projectID})
	if err != nil {
		logger.Ctx(ctx).Error("GetProjectTaskProgress failed", logger.Error(err))
		return domain.ProjectTaskProgress{}, err
	}
	rows, err := r.queries.ListProjectOverdueTasks(ctx, sqlc.ListProjectOverdueTasksParams{BusinessID: businessID, ProjectID: projectID})
	if err != nil {
		logger.Ctx(ctx).Error("ListProjectOverdueTasks failed", logger.Error(err))
		return domain.ProjectTaskProgress{}, err
	}
	overdue := make([]domain.OverdueTask, 0, len(rows))
	for _, row := range rows {
		overdue = append(overdue, domain.OverdueTask{ID: row.ID, Title: row.Title, DueDate: row.DueDate, AssigneeID: row.AssigneeID.Int64})
	}
	return domain.ProjectTaskProgress{
		TotalTasks:     counts.TotalTasks,
		CompletedTasks: counts.CompletedTasks,
		OverdueTasks:   counts.OverdueTasks,
		Overdue:        overdue,
	}, nil
}

// GetProjectTimeTotals sums Flow's own time logs on one project: all time,
// and with logged_at in [thisFrom, thisTo] and [priorFrom, priorTo].
func (r *Repository) GetProjectTimeTotals(ctx context.Context, businessID, projectID int64, thisFrom, thisTo, priorFrom, priorTo time.Time) (domain.ProjectTimeTotals, error) {
	row, err := r.queries.GetProjectLoggedMinutes(ctx, sqlc.GetProjectLoggedMinutesParams{
		ThisFrom:   sql.NullTime{Time: thisFrom, Valid: true},
		ThisTo:     sql.NullTime{Time: thisTo, Valid: true},
		PriorFrom:  sql.NullTime{Time: priorFrom, Valid: true},
		PriorTo:    sql.NullTime{Time: priorTo, Valid: true},
		BusinessID: businessID,
		ProjectID:  projectID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetProjectTimeTotals failed", logger.Error(err))
		return domain.ProjectTimeTotals{}, err
	}
	return domain.ProjectTimeTotals{AllTimeMinutes: row.AllTimeMinutes, ThisMinutes: row.ThisMinutes, PriorMinutes: row.PriorMinutes}, nil
}
