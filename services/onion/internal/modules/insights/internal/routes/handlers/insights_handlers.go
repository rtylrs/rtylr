package handlers

import (
	"context"
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
)

// --- Insights / Analytics ---

func InsightsRoutes() []string { return []string{"/flow/insights"} }

type GetInsightsHandler struct{ Service *application.Service }
type GetInsightsRequest struct{}
type GetInsightsResponse struct {
	TotalTasks      int64   `json:"total_tasks"`
	CompletedTasks  int64   `json:"completed_tasks"`
	InProgressTasks int64   `json:"in_progress_tasks"`
	OverdueTasks    int64   `json:"overdue_tasks"`
	AvgCycleTimeDays float64 `json:"avg_cycle_time_days"`
	Velocity        int64   `json:"velocity"`
	VelocityChange  int64   `json:"velocity_change"`
}

func (h *GetInsightsHandler) Serve(ctx context.Context, _ *GetInsightsRequest) (GetInsightsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetInsightsResponse{}, errorModels.PermissionDeniedServerError()
	}

	var workspaceID *int64
	if wsStr := chi.URLParamFromCtx(ctx, "workspaceId"); wsStr != "" {
		if ws, err := strconv.ParseInt(wsStr, 10, 64); err == nil {
			workspaceID = &ws
		}
	}

	data, err := h.Service.GetInsights(ctx, device.User.UserID, workspaceID)
	if err != nil {
		return GetInsightsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get insights"}
	}
	return GetInsightsResponse{
		TotalTasks:       data.TotalTasks,
		CompletedTasks:   data.CompletedTasks,
		InProgressTasks:  data.InProgressTasks,
		OverdueTasks:     data.OverdueTasks,
		AvgCycleTimeDays: math.Round(data.AvgCycleTimeDays*100) / 100,
		Velocity:         data.TasksThisWeek,
		VelocityChange:   data.TasksThisWeek - data.TasksLastWeek,
	}, nil
}

// --- Team Workload ---

func WorkloadRoutes() []string { return []string{"/flow/team/workload"} }

type GetWorkloadHandler struct{ Service *application.Service }
type GetWorkloadRequest struct{}

type WorkloadPayload struct {
	UserID          int64 `json:"user_id"`
	OpenTasks       int64 `json:"open_tasks"`
	InProgressTasks int64 `json:"in_progress_tasks"`
	CompletedTasks  int64 `json:"completed_tasks"`
	OverdueTasks    int64 `json:"overdue_tasks"`
	TotalActive     int64 `json:"total_active"`
}

type GetWorkloadResponse struct {
	Workload []WorkloadPayload `json:"workload"`
}

func (h *GetWorkloadHandler) Serve(ctx context.Context, _ *GetWorkloadRequest) (GetWorkloadResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetWorkloadResponse{}, errorModels.PermissionDeniedServerError()
	}

	var workspaceID *int64
	if wsStr := chi.URLParamFromCtx(ctx, "workspaceId"); wsStr != "" {
		if ws, err := strconv.ParseInt(wsStr, 10, 64); err == nil {
			workspaceID = &ws
		}
	}

	entries, err := h.Service.GetTeamWorkload(ctx, device.User.UserID, workspaceID)
	if err != nil {
		return GetWorkloadResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get workload"}
	}
	payloads := make([]WorkloadPayload, len(entries))
	for i, e := range entries {
		payloads[i] = WorkloadPayload{
			UserID:          e.UserID,
			OpenTasks:       e.OpenTasks,
			InProgressTasks: e.InProgressTasks,
			CompletedTasks:  e.CompletedTasks,
			OverdueTasks:    e.OverdueTasks,
			TotalActive:     e.OpenTasks + e.InProgressTasks,
		}
	}
	return GetWorkloadResponse{Workload: payloads}, nil
}

// --- Roadmap ---

func RoadmapRoutes() []string { return []string{"/flow/roadmap"} }

type GetRoadmapHandler struct{ Service *application.Service }
type GetRoadmapRequest struct{}

type RoadmapProjectPayload struct {
	ProjectID       int64   `json:"project_id"`
	ProjectName     string  `json:"project_name"`
	Status          string  `json:"status"`
	TotalTasks      int64   `json:"total_tasks"`
	CompletedTasks  int64   `json:"completed_tasks"`
	ProgressPercent float64 `json:"progress_percent"`
	StartDate       *string `json:"start_date"`
	DueDate         *string `json:"due_date"`
}

type GetRoadmapResponse struct {
	Projects []RoadmapProjectPayload `json:"projects"`
}

func (h *GetRoadmapHandler) Serve(ctx context.Context, _ *GetRoadmapRequest) (GetRoadmapResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetRoadmapResponse{}, errorModels.PermissionDeniedServerError()
	}

	var workspaceID *int64
	if wsStr := chi.URLParamFromCtx(ctx, "workspaceId"); wsStr != "" {
		if ws, err := strconv.ParseInt(wsStr, 10, 64); err == nil {
			workspaceID = &ws
		}
	}

	projects, err := h.Service.GetRoadmap(ctx, device.User.UserID, workspaceID)
	if err != nil {
		return GetRoadmapResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get roadmap"}
	}
	payloads := make([]RoadmapProjectPayload, len(projects))
	for i, p := range projects {
		var progress float64
		if p.TotalTasks > 0 {
			progress = math.Round(float64(p.CompletedTasks) / float64(p.TotalTasks) * 10000) / 100
		}
		payload := RoadmapProjectPayload{
			ProjectID:       p.ProjectID,
			ProjectName:     p.ProjectName,
			Status:          p.Status,
			TotalTasks:      p.TotalTasks,
			CompletedTasks:  p.CompletedTasks,
			ProgressPercent: progress,
		}
		if p.StartDate != nil {
			s := p.StartDate.Format("2006-01-02")
			payload.StartDate = &s
		}
		if p.DueDate != nil {
			s := p.DueDate.Format("2006-01-02")
			payload.DueDate = &s
		}
		payloads[i] = payload
	}
	return GetRoadmapResponse{Projects: payloads}, nil
}
