package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrperformancev1 "github.com/voxire/voxie/gen/go/hr/performance/v1"
)

func ListGoalsRoutes() []string { return []string{"/hr/performance/goals"} }

type ListGoalsHandler struct{ Service *application.Service }

type ListGoalsRequest struct {
	EmployeeID    int64 `json:"employee_id"`
	ReviewCycleID int64 `json:"review_cycle_id"`
	Limit         int32 `json:"limit"`
	Offset        int32 `json:"offset"`
}

type ListGoalsResponse struct {
	Goals  []*hrperformancev1.ReviewGoalPayload `json:"goals"`
	Total  int64                                `json:"total"`
	Limit  int32                                `json:"limit"`
	Offset int32                                `json:"offset"`
}

func (h *ListGoalsHandler) Serve(ctx context.Context, req *ListGoalsRequest) (ListGoalsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListGoalsResponse{}, onekiterrors.PermissionDenied()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	goals, total, err := h.Service.ListGoals(ctx, device.User.UserID, domain.ListGoalsParams{
		EmployeeID:    req.EmployeeID,
		ReviewCycleID: req.ReviewCycleID,
		Limit:         limit,
		Offset:        req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list goals", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListGoalsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list goals")
	}

	payloads := make([]*hrperformancev1.ReviewGoalPayload, 0, len(goals))
	for _, g := range goals {
		payloads = append(payloads, goalToProto(g))
	}

	return ListGoalsResponse{Goals: payloads, Total: total, Limit: limit, Offset: req.Offset}, nil
}
