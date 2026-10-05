package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowdashboardv1 "github.com/voxire/voxie/gen/go/flow/dashboard/v1"
)

type DashboardAPI struct {
	Service *application.Service
}

func NewDashboardAPI(service *application.Service) *DashboardAPI {
	return &DashboardAPI{Service: service}
}

func formatWorkspaceID(id int64) string {
	if id == 0 {
		return "personal"
	}
	return strconv.FormatInt(id, 10)
}

func (a *DashboardAPI) GetFlowDashboard(ctx context.Context, req *flowdashboardv1.GetFlowDashboardRequest) (*flowdashboardv1.GetFlowDashboardResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	workspaceID, err := a.Service.ResolveWorkspaceID(ctx, authUser.UserID, businessID, formatWorkspaceID(req.GetWorkspaceId()))
	if err != nil || workspaceID == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid workspace_id")
	}

	stats, err := a.Service.GetDashboard(ctx, workspaceID, businessID, authUser.UserID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get dashboard")
	}

	return &flowdashboardv1.GetFlowDashboardResponse{
		OpenTasks:         stats.OpenTasks,
		InProgressTasks:   stats.InProgressTasks,
		CompletedThisWeek: stats.CompletedThisWeek,
		OverdueTasks:      stats.OverdueTasks,
		MyTasks:           tasksToProto(stats.MyTasks),
	}, nil
}
