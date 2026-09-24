package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrperformancev1 "github.com/voxire/voxie/gen/go/hr/performance/v1"
)

func GetCycleRoutes() []string { return []string{"/hr/performance/cycles/{id}"} }

type GetCycleHandler struct{ Service *application.Service }

type GetCycleRequest struct{}

type GetCycleResponse struct {
	Cycle *hrperformancev1.ReviewCyclePayload `json:"cycle"`
}

func (h *GetCycleHandler) Serve(ctx context.Context, _ *GetCycleRequest) (GetCycleResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetCycleResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetCycleResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetCycleResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid review cycle id")
	}

	cycle, err := h.Service.GetReviewCycle(ctx, businessID, id)
	if err != nil {
		if errors.Is(err, domain.ErrReviewCycleNotFound) {
			return GetCycleResponse{}, onekiterrors.New(http.StatusNotFound, "Review cycle not found")
		}
		logger.Ctx(ctx).Error("failed to get review cycle", logger.Any("business_id", businessID), logger.Any("cycle_id", id), logger.Error(err))
		return GetCycleResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get review cycle")
	}

	return GetCycleResponse{Cycle: cycleToProto(*cycle)}, nil
}
