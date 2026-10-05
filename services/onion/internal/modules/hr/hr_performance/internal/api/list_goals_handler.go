package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrperformancev1 "github.com/voxire/voxie/gen/go/hr/performance/v1"
)

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
	if req == nil {
		req = &ListGoalsRequest{}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListGoalsResponse{}, onekiterrors.PermissionDenied()
	}

	page, err := normalizePerformancePagination(req.Limit, req.Offset)
	if err != nil {
		return ListGoalsResponse{}, err
	}
	limit, offset := page.limit, page.offset
	if req.EmployeeID < 0 || req.ReviewCycleID < 0 {
		return ListGoalsResponse{}, onekiterrors.New(http.StatusBadRequest, "filter ids must be positive when supplied")
	}

	goals, total, err := h.Service.ListGoals(ctx, businessID, domain.ListGoalsParams{
		EmployeeID:    req.EmployeeID,
		ReviewCycleID: req.ReviewCycleID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list goals", logger.Any("business_id", businessID), logger.Error(err))
		return ListGoalsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list goals")
	}

	payloads := make([]*hrperformancev1.ReviewGoalPayload, 0, len(goals))
	for _, g := range goals {
		payloads = append(payloads, goalToProto(g))
	}

	return ListGoalsResponse{Goals: payloads, Total: total, Limit: limit, Offset: offset}, nil
}
