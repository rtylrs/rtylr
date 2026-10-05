package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrperformancev1 "github.com/voxire/voxie/gen/go/hr/performance/v1"
)

func CreateGoalRoutes() []string { return []string{"/hr/performance/goals"} }

type CreateGoalHandler struct{ Service *application.Service }

type CreateGoalRequest struct {
	EmployeeID      int64   `json:"employee_id"`
	ReviewCycleID   int64   `json:"review_cycle_id"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Weight          float64 `json:"weight"`
	SuccessCriteria string  `json:"success_criteria"`
}

type CreateGoalResponse struct {
	Goal *hrperformancev1.ReviewGoalPayload `json:"goal"`
}

func (h *CreateGoalHandler) Validate(_ context.Context, req *CreateGoalRequest) error {
	if req == nil {
		return onekiterrors.New(http.StatusBadRequest, "request body is required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return onekiterrors.New(http.StatusBadRequest, "title is required")
	}
	if req.EmployeeID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id is required")
	}
	if req.ReviewCycleID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "review_cycle_id is required")
	}
	if err := domain.ValidateGoal(domain.CreateGoalParams{
		EmployeeID:    req.EmployeeID,
		ReviewCycleID: req.ReviewCycleID,
		Title:         req.Title,
		Weight:        req.Weight,
	}); err != nil {
		return onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return nil
}

func (h *CreateGoalHandler) Serve(ctx context.Context, req *CreateGoalRequest) (CreateGoalResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateGoalResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.Validate(ctx, req); err != nil {
		return CreateGoalResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateGoalResponse{}, onekiterrors.PermissionDenied()
	}

	goal, err := h.Service.CreateGoal(ctx, businessID, domain.CreateGoalParams{
		EmployeeID:      req.EmployeeID,
		ReviewCycleID:   req.ReviewCycleID,
		Title:           req.Title,
		Description:     req.Description,
		Weight:          req.Weight,
		SuccessCriteria: req.SuccessCriteria,
	})
	if err != nil {
		if errors.Is(err, domain.ErrReviewCycleNotFound) {
			return CreateGoalResponse{}, onekiterrors.New(http.StatusBadRequest, "review_cycle_id is invalid for this business")
		}
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			return CreateGoalResponse{}, onekiterrors.New(http.StatusNotFound, "Employee not found in this business")
		}
		if errors.Is(err, domain.ErrGoalSubjectNotAllowed) {
			return CreateGoalResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrNegativeGoalWeight) || errors.Is(err, domain.ErrInvalidGoalWeight) || errors.Is(err, domain.ErrEmptyGoalTitle) {
			return CreateGoalResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		logger.Ctx(ctx).Error("failed to create goal", logger.Any("business_id", businessID), logger.Error(err))
		return CreateGoalResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create goal")
	}

	return CreateGoalResponse{Goal: goalToProto(*goal)}, nil
}
