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
	TotalTasks      int64   `json:"totalTasks"`
	CompletedTasks  int64   `json:"completedTasks"`
	InProgressTasks int64   `json:"inProgressTasks"`
	OverdueTasks    int64   `json:"overdueTasks"`
	AvgCycleTimeDays float64 `json:"avgCycleTimeDays"`
	Velocity        int64   `json:"velocity"`
	VelocityChange  int64   `json:"velocityChange"`
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
	UserID          int64 `json:"userId"`
	OpenTasks       int64 `json:"openTasks"`
	InProgressTasks int64 `json:"inProgressTasks"`
	CompletedTasks  int64 `json:"completedTasks"`
	OverdueTasks    int64 `json:"overdueTasks"`
	TotalActive     int64 `json:"totalActive"`
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
	ProjectID       int64   `json:"projectId"`
	ProjectName     string  `json:"projectName"`
	Status          string  `json:"status"`
	TotalTasks      int64   `json:"totalTasks"`
	CompletedTasks  int64   `json:"completedTasks"`
	ProgressPercent float64 `json:"progressPercent"`
	StartDate       *string `json:"startDate"`
	DueDate         *string `json:"dueDate"`
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
