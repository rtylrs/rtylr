package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/flow_insights/internal/application"
	"github.com/voxire/onion/internal/modules/flow_insights/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func GetInsightsRoutes() []string     { return []string{"/flow/insights"} }
func GetTeamWorkloadRoutes() []string { return []string{"/flow/team/workload"} }
func GetRoadmapRoutes() []string      { return []string{"/flow/roadmap"} }

// ─────────────────────────────────────────────
// GET /flow/insights
// ─────────────────────────────────────────────

type GetInsightsHandler struct {
	Service *application.Service
}

type GetInsightsRequest struct{}

type GetInsightsResponse struct {
	TotalTasks        int64            `json:"totalTasks"`
	CompletedTasks    int64            `json:"completedTasks"`
	CompletionRate    float64          `json:"completionRate"`
	AvgCompletionDays float64          `json:"avgCompletionDays"`
	VelocityThisWeek  int64            `json:"velocityThisWeek"`
	VelocityLastWeek  int64            `json:"velocityLastWeek"`
	OverdueCount      int64            `json:"overdueCount"`
	TasksByStatus     map[string]int64 `json:"tasksByStatus"`
	TasksByPriority   map[string]int64 `json:"tasksByPriority"`
}

func (h *GetInsightsHandler) Serve(ctx context.Context, _ *GetInsightsRequest) (GetInsightsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetInsightsResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return GetInsightsResponse{}, errorModels.PermissionDeniedServerError()
	}

	stats, err := h.Service.GetInsights(ctx, businessID)
	if err != nil {
		return GetInsightsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get insights",
		}
	}

	return toInsightsResponse(stats), nil
}

func toInsightsResponse(s domain.InsightsStats) GetInsightsResponse {
	return GetInsightsResponse{
		TotalTasks:        s.TotalTasks,
		CompletedTasks:    s.CompletedTasks,
		CompletionRate:    s.CompletionRate,
		AvgCompletionDays: s.AvgCompletionDays,
		VelocityThisWeek:  s.VelocityThisWeek,
		VelocityLastWeek:  s.VelocityLastWeek,
		OverdueCount:      s.OverdueCount,
		TasksByStatus:     s.TasksByStatus,
		TasksByPriority:   s.TasksByPriority,
	}
}

// ─────────────────────────────────────────────
// GET /flow/team/workload
// ─────────────────────────────────────────────

type GetTeamWorkloadHandler struct {
	Service *application.Service
}

type GetTeamWorkloadRequest struct{}

type TeamMemberPayload struct {
	UserID          int64  `json:"userId"`
	Name            string `json:"name"`
	TotalTasks      int64  `json:"totalTasks"`
	CompletedTasks  int64  `json:"completedTasks"`
	InProgressTasks int64  `json:"inProgressTasks"`
	OverdueTasks    int64  `json:"overdueTasks"`
}

type GetTeamWorkloadResponse struct {
	Members []TeamMemberPayload `json:"members"`
}

func (h *GetTeamWorkloadHandler) Serve(ctx context.Context, _ *GetTeamWorkloadRequest) (GetTeamWorkloadResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetTeamWorkloadResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return GetTeamWorkloadResponse{}, errorModels.PermissionDeniedServerError()
	}

	result, err := h.Service.GetTeamWorkload(ctx, businessID)
	if err != nil {
		return GetTeamWorkloadResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get team workload",
		}
	}

	members := make([]TeamMemberPayload, 0, len(result.Members))
	for _, m := range result.Members {
		members = append(members, TeamMemberPayload{
			UserID:          m.UserID,
			Name:            m.Name,
			TotalTasks:      m.TotalTasks,
			CompletedTasks:  m.CompletedTasks,
			InProgressTasks: m.InProgressTasks,
			OverdueTasks:    m.OverdueTasks,
		})
	}

	return GetTeamWorkloadResponse{Members: members}, nil
}

// ─────────────────────────────────────────────
// GET /flow/roadmap
// ─────────────────────────────────────────────

type GetRoadmapHandler struct {
	Service *application.Service
}

type GetRoadmapRequest struct{}

type RoadmapMilestonePayload struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	DueDate   string `json:"dueDate"`
	TaskCount int64  `json:"taskCount"`
	DoneCount int64  `json:"doneCount"`
}

type RoadmapProjectPayload struct {
	ID         int64                     `json:"id"`
	Name       string                    `json:"name"`
	StartDate  string                    `json:"startDate"`
	DueDate    string                    `json:"dueDate"`
	Progress   float64                   `json:"progress"`
	Milestones []RoadmapMilestonePayload `json:"milestones"`
}

type GetRoadmapResponse struct {
	Projects []RoadmapProjectPayload `json:"projects"`
}

func (h *GetRoadmapHandler) Serve(ctx context.Context, _ *GetRoadmapRequest) (GetRoadmapResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetRoadmapResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return GetRoadmapResponse{}, errorModels.PermissionDeniedServerError()
	}

	result, err := h.Service.GetRoadmap(ctx, businessID)
	if err != nil {
		return GetRoadmapResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get roadmap",
		}
	}

	projects := make([]RoadmapProjectPayload, 0, len(result.Projects))
	for _, p := range result.Projects {
		ms := make([]RoadmapMilestonePayload, 0, len(p.Milestones))
		for _, m := range p.Milestones {
			ms = append(ms, RoadmapMilestonePayload{
				ID:        m.ID,
				Name:      m.Name,
				DueDate:   m.DueDate,
				TaskCount: m.TaskCount,
				DoneCount: m.DoneCount,
			})
		}
		projects = append(projects, RoadmapProjectPayload{
			ID:         p.ID,
			Name:       p.Name,
			StartDate:  p.StartDate,
			DueDate:    p.DueDate,
			Progress:   p.Progress,
			Milestones: ms,
		})
	}

	return GetRoadmapResponse{Projects: projects}, nil
}
