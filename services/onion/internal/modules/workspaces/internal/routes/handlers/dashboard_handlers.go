package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
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
	WorkspaceID WorkspaceIDField `json:"workspace_id"`
}

type GetDashboardResponse struct {
	OpenTasks         int64              `json:"open_tasks"`
	InProgressTasks   int64              `json:"in_progress_tasks"`
	CompletedThisWeek int64              `json:"completed_this_week"`
	OverdueTasks      int64              `json:"overdue_tasks"`
	MyTasks           []*flowtaskv1.Task `json:"my_tasks"`
}

func (h *GetDashboardHandler) Validate(_ context.Context, req *GetDashboardRequest) *errorModels.ServerError {
	if req.WorkspaceID.IsEmpty() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "workspace_id is required"}
	}

	return nil
}

func (h *GetDashboardHandler) Serve(ctx context.Context, req *GetDashboardRequest) (GetDashboardResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetDashboardResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return GetDashboardResponse{}, errorModels.PermissionDeniedServerError()
	}

	workspaceID, err := h.FlowService.ResolveWorkspaceID(ctx, authUser.UserID, req.WorkspaceID.String())
	if err != nil || workspaceID == 0 {
		return GetDashboardResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid workspace_id",
		}
	}

	stats, err := h.FlowService.GetDashboard(ctx, workspaceID, businessID, authUser.UserID)
	if err != nil {
		return GetDashboardResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get dashboard",
		}
	}

	myTaskPayloads := make([]*flowtaskv1.Task, 0, len(stats.MyTasks))
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
