package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func GetDashboardRoutes() []string { return []string{"/flow/dashboard"} }

// ─────────────────────────────────────────────
// GET /flow/dashboard
// ─────────────────────────────────────────────

type GetDashboardHandler struct {
	FlowService *application.Service
}

type GetDashboardRequest struct {
	WorkspaceID int64 `json:"workspace_id"`
}

type GetDashboardResponse struct {
	OpenTasks         int64         `json:"open_tasks"`
	InProgressTasks   int64         `json:"in_progress_tasks"`
	CompletedThisWeek int64         `json:"completed_this_week"`
	OverdueTasks      int64         `json:"overdue_tasks"`
	MyTasks           []TaskPayload `json:"my_tasks"`
}

func (h *GetDashboardHandler) Validate(_ context.Context, req *GetDashboardRequest) *errorModels.ServerError {
	if req.WorkspaceID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "workspace_id is required"}
	}

	return nil
}

func (h *GetDashboardHandler) Serve(ctx context.Context, req *GetDashboardRequest) (GetDashboardResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetDashboardResponse{}, errorModels.PermissionDeniedServerError()
	}

	stats, err := h.FlowService.GetDashboard(ctx, req.WorkspaceID, authUser.UserID, authUser.UserID)
	if err != nil {
		return GetDashboardResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get dashboard",
		}
	}

	myTaskPayloads := make([]TaskPayload, 0, len(stats.MyTasks))
	for _, t := range stats.MyTasks {
		myTaskPayloads = append(myTaskPayloads, toTaskPayload(t))
	}

	return GetDashboardResponse{
		OpenTasks:         stats.OpenTasks,
		InProgressTasks:   stats.InProgressTasks,
		CompletedThisWeek: stats.CompletedThisWeek,
		OverdueTasks:      stats.OverdueTasks,
		MyTasks:           myTaskPayloads,
	}, nil
}
