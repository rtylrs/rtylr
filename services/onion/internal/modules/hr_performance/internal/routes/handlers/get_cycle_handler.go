package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	hrperformancev1 "github.com/voxire/voxie/gen/go/hr/performance/v1"
)

func GetCycleRoutes() []string { return []string{"/hr/performance/cycles/{id}"} }

type GetCycleHandler struct{ Service *application.Service }

type GetCycleRequest struct{}

type GetCycleResponse struct {
	Cycle *hrperformancev1.ReviewCyclePayload `json:"cycle"`
}

func (h *GetCycleHandler) Serve(ctx context.Context, _ *GetCycleRequest) (GetCycleResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetCycleResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetCycleResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Invalid review cycle id"}
	}

	cycle, err := h.Service.GetReviewCycle(ctx, device.User.UserID, id)
	if err != nil {
		if errors.Is(err, domain.ErrReviewCycleNotFound) {
			return GetCycleResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Review cycle not found"}
		}
		logger.Ctx(ctx).Error("failed to get review cycle", logger.Any("owner_id", device.User.UserID), logger.Any("cycle_id", id), logger.Error(err))
		return GetCycleResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get review cycle"}
	}

	return GetCycleResponse{Cycle: cycleToProto(*cycle)}, nil
}
