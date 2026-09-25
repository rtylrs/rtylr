package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrperformancev1 "github.com/voxire/voxie/gen/go/hr/performance/v1"
)

func CreateCycleRoutes() []string { return []string{"/hr/performance/cycles"} }

type CreateCycleHandler struct{ Service *application.Service }

type CreateCycleRequest struct {
	Name        string   `json:"name"`
	CycleType   string   `json:"cycle_type"`
	PeriodStart string   `json:"period_start"`
	PeriodEnd   string   `json:"period_end"`
	ReviewTypes []string `json:"review_types"`
}

type CreateCycleResponse struct {
	Cycle *hrperformancev1.ReviewCyclePayload `json:"cycle"`
}

func (h *CreateCycleHandler) Validate(_ context.Context, req *CreateCycleRequest) error {
	if req == nil {
		return onekiterrors.New(http.StatusBadRequest, "request body is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	if req.PeriodStart == "" {
		return onekiterrors.New(http.StatusBadRequest, "period_start is required")
	}
	if req.PeriodEnd == "" {
		return onekiterrors.New(http.StatusBadRequest, "period_end is required")
	}
	if _, err := time.Parse("2006-01-02", req.PeriodStart); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "period_start must be in YYYY-MM-DD format")
	}
	if _, err := time.Parse("2006-01-02", req.PeriodEnd); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "period_end must be in YYYY-MM-DD format")
	}
	start, _ := time.Parse("2006-01-02", req.PeriodStart)
	end, _ := time.Parse("2006-01-02", req.PeriodEnd)
	if end.Before(start) {
		return onekiterrors.New(http.StatusBadRequest, "period_end must not be before period_start")
	}
	return nil
}

func (h *CreateCycleHandler) Serve(ctx context.Context, req *CreateCycleRequest) (CreateCycleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateCycleResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.Validate(ctx, req); err != nil {
		return CreateCycleResponse{}, err
	}
	if err := requirePerformanceManagePermission(ctx); err != nil {
		return CreateCycleResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateCycleResponse{}, onekiterrors.PermissionDenied()
	}

	cycle, err := h.Service.CreateReviewCycle(ctx, businessID, domain.CreateReviewCycleParams{
		Name:        req.Name,
		CycleType:   req.CycleType,
		PeriodStart: req.PeriodStart,
		PeriodEnd:   req.PeriodEnd,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPerformanceAccessDenied) {
			return CreateCycleResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, domain.ErrEmptyCycleName) || errors.Is(err, domain.ErrInvalidCycleType) {
			return CreateCycleResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		logger.Ctx(ctx).Error("failed to create review cycle", logger.Any("business_id", businessID), logger.Error(err))
		return CreateCycleResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create review cycle")
	}

	return CreateCycleResponse{Cycle: cycleToProto(*cycle)}, nil
}
