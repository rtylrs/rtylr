package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	insightsdashboardv1 "github.com/voxire/voxie/gen/go/insights/dashboard/v1"
)

// projectInvoicePermission is the finance permission that lets a viewer see a
// project's money. Without it the dashboard still answers, without invoices.
const projectInvoicePermission = "finance:invoices:view"

// ListProjects lists the Work projects the caller may open a project
// dashboard for. Flow decides which: every public project, and the private
// ones the caller owns, created or is a project member of.
func (a *DashboardAPI) ListProjects(ctx context.Context, _ *insightsdashboardv1.ListProjectsRequest) (*insightsdashboardv1.ListProjectsResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}
	userID := authPkg.DeviceCtx(ctx).User.UserID

	projects, err := a.Service.ListDashboardProjects(ctx, businessID, userID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load projects")
	}

	out := make([]*insightsdashboardv1.DashboardProject, 0, len(projects))
	for _, p := range projects {
		out = append(out, toDashboardProject(p))
	}
	return &insightsdashboardv1.ListProjectsResponse{Projects: out}, nil
}

// GetProjectDashboard reports one Work project. A private project the caller
// is not a member of answers 404, exactly like a missing project or another
// business's, so its existence does not leak. Invoice totals are included only
// when the caller may view finance invoices.
func (a *DashboardAPI) GetProjectDashboard(ctx context.Context, req *insightsdashboardv1.GetProjectDashboardRequest) (*insightsdashboardv1.GetProjectDashboardResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:dashboard:view")
	if err != nil {
		return nil, err
	}
	user := authPkg.DeviceCtx(ctx).User
	withInvoices := authPkg.HasAppPermission(user, authorization.AppFinance, projectInvoicePermission)

	dashboard, err := a.Service.GetProjectDashboard(ctx, businessID, user.UserID, req.GetProjectId(), withInvoices)
	if errors.Is(err, application.ErrProjectNotFound) {
		return nil, onekiterrors.New(http.StatusNotFound, "Project not found")
	}
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load project dashboard")
	}

	return toProjectDashboardResponse(dashboard), nil
}

func toDashboardProject(p domain.DashboardProject) *insightsdashboardv1.DashboardProject {
	return &insightsdashboardv1.DashboardProject{
		ProjectId:   p.ID,
		ProjectName: p.Name,
		Status:      p.Status,
		Visibility:  p.Visibility,
	}
}

func toProjectDashboardResponse(d domain.ProjectDashboard) *insightsdashboardv1.GetProjectDashboardResponse {
	overdue := make([]*insightsdashboardv1.OverdueTask, 0, len(d.Overdue))
	for _, t := range d.Overdue {
		task := &insightsdashboardv1.OverdueTask{TaskId: t.ID, Title: t.Title, DueDate: t.DueDate}
		if t.AssigneeID > 0 {
			assignee := t.AssigneeID
			task.AssigneeId = &assignee
			if t.AssigneeName != "" {
				name := t.AssigneeName
				task.AssigneeName = &name
			}
		}
		overdue = append(overdue, task)
	}
	resp := &insightsdashboardv1.GetProjectDashboardResponse{
		Project: toDashboardProject(d.Project),
		Progress: &insightsdashboardv1.ProjectProgress{
			TotalTasks:     d.TotalTasks,
			CompletedTasks: d.CompletedTasks,
			Percent:        d.ProgressPct,
		},
		OverdueTaskCount: d.OverdueTasks,
		OverdueTasks:     overdue,
		Hours: &insightsdashboardv1.ProjectHoursLogged{
			ThisMonth: &insightsdashboardv1.InsightsMetricCard{
				Value:    d.Logged.Value,
				DeltaPct: d.Logged.DeltaPct,
			},
			PriorMinutes:   d.PriorMinutes,
			AllTimeMinutes: d.AllTimeMinutes,
		},
	}
	if d.Invoices != nil {
		resp.Invoices = &insightsdashboardv1.ProjectInvoices{
			InvoiceCount: d.Invoices.InvoiceCount,
			Invoiced:     d.Invoices.Invoiced,
			Paid:         d.Invoices.Paid,
			Outstanding:  d.Invoices.Outstanding,
			Overdue:      d.Invoices.Overdue,
			OverdueCount: d.Invoices.OverdueCount,
		}
	}
	return resp
}
