package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func UpdateGoalRoutes() []string { return []string{"/hr/performance/goals/{id}"} }

type UpdateGoalHandler struct{ Service *application.Service }

type UpdateGoalRequest struct {
	ID              int64   `json:"id"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Weight          float64 `json:"weight"`
	SuccessCriteria string  `json:"success_criteria"`
	Status          string  `json:"status"`
}

type UpdateGoalResponse struct {
	Success bool `json:"success"`
}

func (h *UpdateGoalHandler) Validate(_ context.Context, req *UpdateGoalRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return onekiterrors.New(http.StatusBadRequest, "title is required")
	}
	return nil
}

func (h *UpdateGoalHandler) Serve(ctx context.Context, req *UpdateGoalRequest) (UpdateGoalResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateGoalResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateGoalResponse{}, onekiterrors.PermissionDenied()
	}

	err := h.Service.UpdateGoal(ctx, businessID, req.ID, domain.UpdateGoalParams{
		Title:           req.Title,
		Description:     req.Description,
		Weight:          req.Weight,
		SuccessCriteria: req.SuccessCriteria,
		Status:          req.Status,
	})
	if err != nil {
		if errors.Is(err, domain.ErrGoalNotFound) {
			return UpdateGoalResponse{}, onekiterrors.New(http.StatusNotFound, "Goal not found")
		}
		logger.Ctx(ctx).Error("failed to update goal", logger.Any("owner_id", device.User.UserID), logger.Any("goal_id", req.ID), logger.Error(err))
		return UpdateGoalResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update goal")
	}

	return UpdateGoalResponse{Success: true}, nil
}
