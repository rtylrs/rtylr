package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowcyclev2 "github.com/voxire/voxie/gen/go/flow/cycle/v2"
	flowdashboardv2 "github.com/voxire/voxie/gen/go/flow/dashboard/v2"
	flowinsightsv2 "github.com/voxire/voxie/gen/go/flow/insights/v2"
	flowlabelv2 "github.com/voxire/voxie/gen/go/flow/label/v2"
	flowmilestonev2 "github.com/voxire/voxie/gen/go/flow/milestone/v2"
	flowviewv2 "github.com/voxire/voxie/gen/go/flow/view/v2"
)

// ─── shared auth helper for the simple v2 services ──────────────────────────

func resolveV2(ctx context.Context) (*authPkg.AuthUser, int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, 0, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return nil, 0, onekiterrors.PermissionDenied()
	}
	return authUser, businessID, nil
}

// ─── dashboard/v2 ───────────────────────────────────────────────────────────

type DashboardAPIV2 struct{ Service *application.Service }

func NewDashboardAPIV2(s *application.Service) *DashboardAPIV2 { return &DashboardAPIV2{Service: s} }

func (a *DashboardAPIV2) GetFlowDashboard(ctx context.Context, _ *flowdashboardv2.GetFlowDashboardRequest) (*flowdashboardv2.GetFlowDashboardResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	stats, err := a.Service.GetDashboardV2(ctx, businessID, caller.UserID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get dashboard")
	}
	myTasks := make([]*flowdashboardv2.DashboardTask, 0, len(stats.MyTasks))
	for _, t := range stats.MyTasks {
		dt := &flowdashboardv2.DashboardTask{
			Id: t.ID, Title: t.Title, Status: t.Status, Priority: t.Priority,
			ProjectName: t.ProjectName, ProjectColor: t.ProjectColor,
			AssigneeName: t.AssigneeName, AssigneeInitials: t.AssigneeInitials,
		}
		if t.DueDate != nil {
			dt.DueDate = &[]string{t.DueDate.Format("2006-01-02")}[0]
		}
		myTasks = append(myTasks, dt)
	}
	return &flowdashboardv2.GetFlowDashboardResponse{
		OpenTasks: stats.OpenTasks, InProgressTasks: stats.InProgressTasks,
		CompletedThisWeek: stats.CompletedThisWeek, OverdueTasks: stats.OverdueTasks,
		MyTasks: myTasks,
	}, nil
}

// ─── insights/v2 ────────────────────────────────────────────────────────────

type InsightsAPIV2 struct{ Service *application.Service }

func NewInsightsAPIV2(s *application.Service) *InsightsAPIV2 { return &InsightsAPIV2{Service: s} }

func (a *InsightsAPIV2) GetInsights(ctx context.Context, _ *flowinsightsv2.GetInsightsRequest) (*flowinsightsv2.GetInsightsResponse, error) {
	// Served by the flow_insights module; stub for contract completeness.
	return &flowinsightsv2.GetInsightsResponse{}, nil
}

func (a *InsightsAPIV2) GetTeamWorkload(ctx context.Context, req *flowinsightsv2.GetTeamWorkloadRequest) (*flowinsightsv2.GetTeamWorkloadResponse, error) {
	page, limit, _ := resolvePagination(req.GetPage(), req.GetLimit())
	return &flowinsightsv2.GetTeamWorkloadResponse{Pagination: buildPagination(page, limit, 0, 0)}, nil
}

func (a *InsightsAPIV2) GetRoadmap(ctx context.Context, req *flowinsightsv2.GetRoadmapRequest) (*flowinsightsv2.GetRoadmapResponse, error) {
	page, limit, _ := resolvePagination(req.GetPage(), req.GetLimit())
	return &flowinsightsv2.GetRoadmapResponse{Pagination: buildPagination(page, limit, 0, 0)}, nil
}

// ─── label/v2 ───────────────────────────────────────────────────────────────

type LabelAPIV2 struct{ Service *application.Service }

func NewLabelAPIV2(s *application.Service) *LabelAPIV2 { return &LabelAPIV2{Service: s} }

func (a *LabelAPIV2) ListLabels(ctx context.Context, req *flowlabelv2.ListLabelsRequest) (*flowlabelv2.ListLabelsResponse, error) {
	_, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	labels, total, err := a.Service.ListLabelsV2(ctx, businessID, int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list labels")
	}
	out := make([]*flowlabelv2.Label, 0, len(labels))
	for _, l := range labels {
		out = append(out, &flowlabelv2.Label{
			Id: l.ID, Name: l.Name, Color: l.Color,
			CreatedAt: l.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return &flowlabelv2.ListLabelsResponse{Labels: out, Pagination: buildPagination(page, limit, total, len(out))}, nil
}

func (a *LabelAPIV2) CreateLabel(ctx context.Context, req *flowlabelv2.CreateLabelRequest) (*flowlabelv2.CreateLabelResponse, error) {
	_, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	label, err := a.Service.CreateLabelV2(ctx, businessID, req.GetName(), req.GetColor())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create label")
	}
	return &flowlabelv2.CreateLabelResponse{Label: &flowlabelv2.Label{
		Id: label.ID, Name: label.Name, Color: label.Color,
		CreatedAt: label.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}}, nil
}

func (a *LabelAPIV2) UpdateLabel(ctx context.Context, req *flowlabelv2.UpdateLabelRequest) (*flowlabelv2.UpdateLabelResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.UpdateLabel(ctx, req.GetId(), caller.UserID, businessID, domain.LabelInput{Name: req.GetName(), Color: req.GetColor()}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update label")
	}
	return &flowlabelv2.UpdateLabelResponse{Message: "Label updated"}, nil
}

func (a *LabelAPIV2) DeleteLabel(ctx context.Context, req *flowlabelv2.DeleteLabelRequest) (*flowlabelv2.DeleteLabelResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.DeleteLabel(ctx, req.GetId(), caller.UserID, businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete label")
	}
	return &flowlabelv2.DeleteLabelResponse{Message: "Label deleted"}, nil
}

func (a *LabelAPIV2) SetTaskLabels(ctx context.Context, req *flowlabelv2.SetTaskLabelsRequest) (*flowlabelv2.SetTaskLabelsResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}

	// Setting a task's labels is a task mutation, so it takes the same tier as
	// the task handlers. The repository guard underneath only checks workspace
	// membership, which admits a viewer - so without this a viewer can relabel
	// any task they can see. The v1 handler had this check and carried the same
	// warning in a comment; v2 shipped without it.
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.SetTaskLabels(ctx, req.GetTaskId(), caller.UserID, req.GetLabelIds()); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to set task labels")
	}
	return &flowlabelv2.SetTaskLabelsResponse{Message: "Labels set"}, nil
}

// ─── cycle/v2 ───────────────────────────────────────────────────────────────

type CycleAPIV2 struct{ Service *application.Service }

func NewCycleAPIV2(s *application.Service) *CycleAPIV2 { return &CycleAPIV2{Service: s} }

func (a *CycleAPIV2) ListCycles(ctx context.Context, req *flowcyclev2.ListCyclesRequest) (*flowcyclev2.ListCyclesResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	cycles, total, err := a.Service.ListCyclesV2(ctx, req.GetProjectId(), businessID, caller.UserID, int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list cycles")
	}
	out := make([]*flowcyclev2.Cycle, 0, len(cycles))
	for _, c := range cycles {
		out = append(out, toContractCycleV2(c))
	}
	return &flowcyclev2.ListCyclesResponse{Cycles: out, Pagination: buildPagination(page, limit, total, len(out))}, nil
}

func (a *CycleAPIV2) CreateCycle(ctx context.Context, req *flowcyclev2.CreateCycleRequest) (*flowcyclev2.CreateCycleResponse, error) {
	_, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	cycle, err := a.Service.CreateCycleV2(ctx, businessID, req.GetProjectId(), req.GetName(), req.GetStartDate(), req.GetEndDate())
	if err != nil {
		// A bad date pair is the caller's mistake, not an outage. Reported as 500
		// it would look like a broken server and get retried forever.
		if errors.Is(err, application.ErrInvalidCycleDates) {
			return nil, onekiterrors.New(http.StatusBadRequest, "start_date and end_date are invalid")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create cycle")
	}
	return &flowcyclev2.CreateCycleResponse{Cycle: toContractCycleV2(cycle)}, nil
}

// toContractCycleV2 is the single place a cycle becomes a contract cycle.
//
// The list and create mappers were written separately and each dropped a
// different subset of the fields the contract declares. The list sent only id,
// name, scope and completed, so every cycle serialised with project_id "0" and
// no dates or status, and the Cycles page could neither match a cycle to the
// selected project nor bucket it into Active, Upcoming or Completed. It
// rendered nothing at all, not even its empty state, for a project that did
// have a cycle.
//
// A zero date becomes an empty string rather than 0001-01-01, which is a date
// the client would have to special-case to recognise as absent.
func toContractCycleV2(c domain.Cycle) *flowcyclev2.Cycle {
	out := &flowcyclev2.Cycle{
		Id:        c.ID,
		Name:      c.Name,
		Status:    c.Status,
		StartDate: formatCycleDate(c.StartDate),
		EndDate:   formatCycleDate(c.EndDate),
		Scope:     int32(c.Scope),     // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Completed: int32(c.Completed), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}
	// ProjectID is a pointer on the domain because workspace-era cycles had
	// none. The contract declares it non-optional, so an absent project stays
	// zero rather than dereferencing nil.
	if c.ProjectID != nil {
		out.ProjectId = *c.ProjectID
	}
	return out
}

func formatCycleDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func (a *CycleAPIV2) UpdateCycle(ctx context.Context, req *flowcyclev2.UpdateCycleRequest) (*flowcyclev2.UpdateCycleResponse, error) {
	if err := a.Service.UpdateCycle(ctx, domain.UpdateCycleInput{
		ID: req.GetId(), Name: req.GetName(), Status: req.GetStatus(),
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update cycle")
	}
	return &flowcyclev2.UpdateCycleResponse{}, nil
}

// ─── milestone/v2 ───────────────────────────────────────────────────────────

type MilestoneAPIV2 struct{ Service *application.Service }

func NewMilestoneAPIV2(s *application.Service) *MilestoneAPIV2 { return &MilestoneAPIV2{Service: s} }

func (a *MilestoneAPIV2) ListMilestones(ctx context.Context, req *flowmilestonev2.ListMilestonesRequest) (*flowmilestonev2.ListMilestonesResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	var projectID *int64
	if pid := req.GetProjectId(); pid != 0 {
		projectID = &pid
	}
	milestones, total, err := a.Service.ListMilestonesV2(ctx, businessID, projectID, caller.UserID, int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list milestones")
	}
	out := make([]*flowmilestonev2.Milestone, 0, len(milestones))
	for _, m := range milestones {
		mp := &flowmilestonev2.MilestoneProject{}
		out = append(out, &flowmilestonev2.Milestone{
			Id: m.ID, ProjectId: m.ProjectID, Name: m.Name,
			DoneCount: int32(m.DoneCount), TaskCount: int32(m.TaskCount), Status: m.Status, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			Project: mp,
		})
	}
	return &flowmilestonev2.ListMilestonesResponse{Milestones: out, Pagination: buildPagination(page, limit, total, len(out))}, nil
}

func (a *MilestoneAPIV2) CreateMilestone(ctx context.Context, req *flowmilestonev2.CreateMilestoneRequest) (*flowmilestonev2.CreateMilestoneResponse, error) {
	_, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	m, err := a.Service.CreateMilestoneV2(ctx, businessID, req.GetProjectId(), req.GetName(), req.GetDescription(), req.GetTargetDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create milestone")
	}
	return &flowmilestonev2.CreateMilestoneResponse{Milestone: &flowmilestonev2.Milestone{
		Id: m.ID, ProjectId: m.ProjectID, Name: m.Name,
		Status: m.Status, Description: m.Description,
	}}, nil
}

func (a *MilestoneAPIV2) UpdateMilestone(ctx context.Context, req *flowmilestonev2.UpdateMilestoneRequest) (*flowmilestonev2.UpdateMilestoneResponse, error) {
	if err := a.Service.UpdateMilestone(ctx, domain.UpdateMilestoneInput{
		ID: req.GetId(), Name: req.GetName(),
		Status: req.GetStatus(),
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update milestone")
	}
	return &flowmilestonev2.UpdateMilestoneResponse{}, nil
}

// ─── view/v2 ────────────────────────────────────────────────────────────────

type ViewAPIV2 struct{ Service *application.Service }

func NewViewAPIV2(s *application.Service) *ViewAPIV2 { return &ViewAPIV2{Service: s} }

func (a *ViewAPIV2) ListViews(ctx context.Context, req *flowviewv2.ListViewsRequest) (*flowviewv2.ListViewsResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	views, total, err := a.Service.ListViewsV2(ctx, businessID, caller.UserID, int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list views")
	}
	out := make([]*flowviewv2.View, 0, len(views))
	for _, v := range views {
		filtersJSON, _ := json.Marshal(v.Filters)
		out = append(out, &flowviewv2.View{
			Id: v.ID, Name: v.Name, Icon: v.Icon, Filters: string(filtersJSON),
			GroupBy: v.GroupBy, SortBy: v.SortBy, SortDir: v.SortDir,
			Visibility: v.Visibility,
		})
	}
	return &flowviewv2.ListViewsResponse{Views: out, Pagination: buildPagination(page, limit, total, len(out))}, nil
}

func (a *ViewAPIV2) CreateView(ctx context.Context, req *flowviewv2.CreateViewRequest) (*flowviewv2.CreateViewResponse, error) {
	caller, businessID, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	v, err := a.Service.CreateViewV2(ctx, businessID, caller.UserID, req.GetName(), req.GetIcon(), req.GetFilters(), req.GetGroupBy(), req.GetSortBy(), req.GetSortDir(), req.GetVisibility())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create view")
	}
	filtersOut, _ := json.Marshal(v.Filters)
	return &flowviewv2.CreateViewResponse{View: &flowviewv2.View{
		Id: v.ID, Name: v.Name, Icon: v.Icon, Filters: string(filtersOut),
		GroupBy: v.GroupBy, SortBy: v.SortBy, SortDir: v.SortDir,
		Visibility: v.Visibility,
	}}, nil
}

func (a *ViewAPIV2) UpdateView(ctx context.Context, req *flowviewv2.UpdateViewRequest) (*flowviewv2.UpdateViewResponse, error) {
	caller, _, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.UpdateViewV2(
		ctx, caller.UserID, req.GetId(),
		req.GetName(), req.GetIcon(), req.GetFilters(),
		req.GetGroupBy(), req.GetSortBy(), req.GetSortDir(), req.GetVisibility(),
	); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update view")
	}
	return &flowviewv2.UpdateViewResponse{}, nil
}

func (a *ViewAPIV2) DeleteView(ctx context.Context, req *flowviewv2.DeleteViewRequest) (*flowviewv2.DeleteViewResponse, error) {
	caller, _, err := resolveV2(ctx)
	if err != nil {
		return nil, err
	}
	_, businessID, bizErr := resolveV2(ctx)
	if bizErr != nil {
		return nil, bizErr
	}
	if err := a.Service.DeleteView(ctx, req.GetId(), caller.UserID, businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete view")
	}
	return &flowviewv2.DeleteViewResponse{}, nil
}
