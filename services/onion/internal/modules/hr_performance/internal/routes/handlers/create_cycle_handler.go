package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
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

func (h *CreateCycleHandler) Validate(_ context.Context, req *CreateCycleRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "name is required"}
	}
	if req.PeriodStart == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "period_start is required"}
	}
	if req.PeriodEnd == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "period_end is required"}
	}
	if _, err := time.Parse("2006-01-02", req.PeriodStart); err != nil {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "period_start must be in YYYY-MM-DD format"}
	}
	if _, err := time.Parse("2006-01-02", req.PeriodEnd); err != nil {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "period_end must be in YYYY-MM-DD format"}
	}
	return nil
}

func (h *CreateCycleHandler) Serve(ctx context.Context, req *CreateCycleRequest) (CreateCycleResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateCycleResponse{}, errorModels.PermissionDeniedServerError()
	}

	cycle, err := h.Service.CreateReviewCycle(ctx, device.User.UserID, domain.CreateReviewCycleParams{
		Name:        req.Name,
		CycleType:   req.CycleType,
		PeriodStart: req.PeriodStart,
		PeriodEnd:   req.PeriodEnd,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create review cycle", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return CreateCycleResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create review cycle"}
	}

	return CreateCycleResponse{Cycle: cycleToProto(*cycle)}, nil
}
