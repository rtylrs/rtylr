package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
)

// ErrProjectNotFound is returned for a project the viewer may not open: one
// that does not exist, belongs to another business, or is private and Flow
// does not let the viewer see it. The three are deliberately
// indistinguishable.
var ErrProjectNotFound = errors.New("insights: project not found")

// ListDashboardProjects returns the Work projects userID may open a project
// dashboard for. It is not cached: the answer depends on the viewer.
func (s *Service) ListDashboardProjects(ctx context.Context, businessID, userID int64) ([]domain.DashboardProject, error) {
	flow := s.dashboard.FlowProjects
	if flow == nil {
		return nil, errDashboardReadersUnavailable
	}
	projects, err := flow.FlowDashboardProjects(ctx, businessID, userID)
	if err != nil {
		return nil, fmt.Errorf("read dashboard projects: %w", err)
	}
	if projects == nil {
		projects = []domain.DashboardProject{}
	}
	return projects, nil
}

// GetProjectDashboard returns one project's dashboard for userID, or
// ErrProjectNotFound. withInvoices asks for the project's invoice totals; the
// caller decides whether the viewer may see them. Nothing is cached: the
// visibility check depends on the viewer, and a project's own page should
// reflect a task closed a moment ago.
func (s *Service) GetProjectDashboard(ctx context.Context, businessID, userID, projectID int64, withInvoices bool) (domain.ProjectDashboard, error) {
	return s.readProjectDashboard(ctx, businessID, userID, projectID, withInvoices, time.Now())
}

func (s *Service) readProjectDashboard(ctx context.Context, businessID, userID, projectID int64, withInvoices bool, now time.Time) (domain.ProjectDashboard, error) {
	flow := s.dashboard.FlowProjects
	if flow == nil || (withInvoices && s.dashboard.FinanceProjects == nil) {
		return domain.ProjectDashboard{}, errDashboardReadersUnavailable
	}
	if projectID <= 0 || userID <= 0 {
		return domain.ProjectDashboard{}, ErrProjectNotFound
	}
	project, found, err := flow.FlowDashboardProject(ctx, businessID, userID, projectID)
	if err != nil {
		return domain.ProjectDashboard{}, fmt.Errorf("read dashboard project: %w", err)
	}
	if !found {
		return domain.ProjectDashboard{}, ErrProjectNotFound
	}

	progress, err := flow.FlowProjectProgress(ctx, businessID, projectID)
	if err != nil {
		return domain.ProjectDashboard{}, fmt.Errorf("read project progress: %w", err)
	}
	thisFrom, thisTo, priorFrom, priorTo := dashboardMTDRanges(now)
	logged, err := flow.FlowProjectTime(ctx, businessID, projectID, thisFrom, thisTo, priorFrom, priorTo)
	if err != nil {
		return domain.ProjectDashboard{}, fmt.Errorf("read project time: %w", err)
	}

	overdue := progress.Overdue
	if overdue == nil {
		overdue = []domain.ProjectOverdueTask{}
	}
	dashboard := domain.ProjectDashboard{
		Project:        project,
		TotalTasks:     progress.TotalTasks,
		CompletedTasks: progress.CompletedTasks,
		ProgressPct:    progressPct(progress.CompletedTasks, progress.TotalTasks),
		OverdueTasks:   progress.OverdueTasks,
		Overdue:        overdue,
		Logged:         domain.CountCard{Value: logged.ThisMinutes, DeltaPct: dashboardDeltaPct(logged.ThisMinutes, logged.PriorMinutes)},
		PriorMinutes:   logged.PriorMinutes,
		AllTimeMinutes: logged.AllTimeMinutes,
	}
	if !withInvoices {
		return dashboard, nil
	}
	totals, err := s.dashboard.FinanceProjects.InsightsProjectInvoices(ctx, businessID, projectID)
	if err != nil {
		return domain.ProjectDashboard{}, fmt.Errorf("read project invoices: %w", err)
	}
	dashboard.Invoices = &domain.ProjectInvoices{
		InvoiceCount: totals.InvoiceCount,
		Invoiced:     totals.Invoiced,
		Paid:         totals.Paid,
		Outstanding:  totals.Outstanding,
		Overdue:      totals.Overdue,
		OverdueCount: totals.OverdueCount,
	}
	return dashboard, nil
}

// progressPct is done as a share of total, 0 to 100, and 0 for a project
// with no tasks.
func progressPct(done, total int64) float64 {
	if total <= 0 {
		return 0
	}
	pct := float64(done) / float64(total) * 100
	if pct > 100 {
		return 100
	}
	return pct
}
