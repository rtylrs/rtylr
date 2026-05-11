package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── Insights ─────────────────────────────────────────────────────────────────

func GetInsightsRoutes() []string { return []string{"/flow/insights"} }

type GetInsightsRequest struct {
	WorkspaceID string `query:"workspace_id"`
}

type GetInsightsResponse struct {
	Velocity     []velocityPoint   `json:"velocity"`
	Burndown     []burndownPoint   `json:"burndown"`
	CycleTime    float64           `json:"cycleTime"`
	TimeByMember []memberTime      `json:"timeByMember"`
	Throughput   []throughputPoint `json:"throughput"`
	TotalTasks   int               `json:"totalTasks"`
	DoneTasks    int               `json:"doneTasks"`
}

type velocityPoint struct {
	Cycle          string  `json:"cycle"`
	Completed      int     `json:"completed"`
	Total          int     `json:"total"`
	CompletionRate float64 `json:"completionRate"`
}

type burndownPoint struct {
	Day    int     `json:"day"`
	Ideal  float64 `json:"ideal"`
	Actual float64 `json:"actual"`
}

type memberTime struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Hours float64 `json:"hours"`
}

type throughputPoint struct {
	Week      string `json:"week"`
	Completed int    `json:"completed"`
}

func (r *GetInsightsRequest) Validate() *errorModels.ServerError { return nil }

type GetInsightsHandler struct{ FlowService *application.Service }

func (h *GetInsightsHandler) Serve(ctx context.Context, r *GetInsightsRequest) (GetInsightsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetInsightsResponse{}, errorModels.PermissionDeniedServerError()
	}

	wsID := authUser.BusinessID
	if r.WorkspaceID != "" {
		if id, err := strconv.ParseInt(r.WorkspaceID, 10, 64); err == nil {
			wsID = id
		}
	}

	ins, err := h.FlowService.GetInsights(ctx, wsID)
	if err != nil {
		return GetInsightsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	resp := GetInsightsResponse{
		CycleTime:  ins.CycleTime,
		TotalTasks: ins.TotalTasks,
		DoneTasks:  ins.DoneTasks,
	}

	for _, v := range ins.Velocity {
		resp.Velocity = append(resp.Velocity, velocityPoint{
			Cycle: v.Cycle, Completed: v.Completed, Total: v.Total, CompletionRate: v.CompletionRate,
		})
	}
	for _, b := range ins.Burndown {
		resp.Burndown = append(resp.Burndown, burndownPoint{Day: b.Day, Ideal: b.Ideal, Actual: b.Actual})
	}
	for _, m := range ins.TimeByMember {
		resp.TimeByMember = append(resp.TimeByMember, memberTime{ID: m.ID, Name: m.Name, Hours: m.Hours})
	}
	for _, t := range ins.Throughput {
		resp.Throughput = append(resp.Throughput, throughputPoint{Week: t.Week, Completed: t.Completed})
	}

	if resp.Velocity == nil {
		resp.Velocity = []velocityPoint{}
	}
	if resp.Burndown == nil {
		resp.Burndown = []burndownPoint{}
	}
	if resp.TimeByMember == nil {
		resp.TimeByMember = []memberTime{}
	}
	if resp.Throughput == nil {
		resp.Throughput = []throughputPoint{}
	}

	return resp, nil
}

// ─── Team Workload ────────────────────────────────────────────────────────────

func GetTeamWorkloadRoutes() []string { return []string{"/flow/team/workload"} }

type GetTeamWorkloadRequest struct {
	WorkspaceID string `query:"workspace_id"`
}

type GetTeamWorkloadResponse struct {
	Members []teamMemberPayload `json:"members"`
}

type teamMemberPayload struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Initials      string        `json:"initials"`
	Capacity      int           `json:"capacity"`
	AssignedHours float64       `json:"assignedHours"`
	Load          float64       `json:"load"`
	Tasks         []taskSummary `json:"tasks"`
}

type taskSummary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
	DueDate  string `json:"dueDate,omitempty"`
}

func (r *GetTeamWorkloadRequest) Validate() *errorModels.ServerError { return nil }

type GetTeamWorkloadHandler struct{ FlowService *application.Service }

func (h *GetTeamWorkloadHandler) Serve(ctx context.Context, r *GetTeamWorkloadRequest) (GetTeamWorkloadResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetTeamWorkloadResponse{}, errorModels.PermissionDeniedServerError()
	}

	wsID := authUser.BusinessID
	if r.WorkspaceID != "" {
		if id, err := strconv.ParseInt(r.WorkspaceID, 10, 64); err == nil {
			wsID = id
		}
	}

	workload, err := h.FlowService.GetTeamWorkload(ctx, wsID)
	if err != nil {
		return GetTeamWorkloadResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	members := make([]teamMemberPayload, 0, len(workload))
	for _, w := range workload {
		tasks := make([]taskSummary, 0, len(w.Tasks))
		for _, t := range w.Tasks {
			s := taskSummary{
				ID:       strconv.FormatInt(t.ID, 10),
				Title:    t.Title,
				Status:   t.Status,
				Priority: t.Priority,
			}
			if t.DueDate != nil {
				s.DueDate = t.DueDate.Format("2006-01-02")
			}
			tasks = append(tasks, s)
		}
		members = append(members, teamMemberPayload{
			ID:            w.ID,
			Name:          w.Name,
			Initials:      w.Initials,
			Capacity:      w.Capacity,
			AssignedHours: w.AssignedHours,
			Load:          w.Load,
			Tasks:         tasks,
		})
	}
	return GetTeamWorkloadResponse{Members: members}, nil
}

// ─── Roadmap ──────────────────────────────────────────────────────────────────

func GetRoadmapRoutes() []string { return []string{"/flow/roadmap"} }

type GetRoadmapRequest struct {
	WorkspaceID string `query:"workspace_id"`
}

type GetRoadmapResponse struct {
	Projects []roadmapProject `json:"projects"`
}

type roadmapProject struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	StartDate   string `json:"startDate"`
	TargetDate  string `json:"targetDate"`
	Progress    int    `json:"progress"`
	DoneCount   int    `json:"doneCount"`
	TaskCount   int    `json:"taskCount"`
	Description string `json:"description,omitempty"`
	WorkspaceID string `json:"workspaceId"`
}

func (r *GetRoadmapRequest) Validate() *errorModels.ServerError { return nil }

type GetRoadmapHandler struct{ FlowService *application.Service }

func (h *GetRoadmapHandler) Serve(ctx context.Context, r *GetRoadmapRequest) (GetRoadmapResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetRoadmapResponse{}, errorModels.PermissionDeniedServerError()
	}

	wsID := authUser.BusinessID
	if r.WorkspaceID != "" {
		if id, err := strconv.ParseInt(r.WorkspaceID, 10, 64); err == nil {
			wsID = id
		}
	}

	projects, err := h.FlowService.GetRoadmap(ctx, wsID)
	if err != nil {
		return GetRoadmapResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	payloads := make([]roadmapProject, 0, len(projects))
	for _, p := range projects {
		total, done := h.FlowService.GetProjectTaskCounts(ctx, p.ID)
		progress := 0
		if total > 0 {
			progress = int(float64(done) / float64(total) * 100)
		}

		rp := roadmapProject{
			ID:          strconv.FormatInt(p.ID, 10),
			Name:        p.Name,
			Color:       p.Color,
			Progress:    progress,
			DoneCount:   done,
			TaskCount:   total,
			Description: p.Description,
			WorkspaceID: strconv.FormatInt(p.WorkspaceID, 10),
		}
		if p.StartDate != nil {
			rp.StartDate = p.StartDate.Format("2006-01-02")
		}
		if p.DueDate != nil {
			rp.TargetDate = p.DueDate.Format("2006-01-02")
		}
		payloads = append(payloads, rp)
	}
	return GetRoadmapResponse{Projects: payloads}, nil
}
