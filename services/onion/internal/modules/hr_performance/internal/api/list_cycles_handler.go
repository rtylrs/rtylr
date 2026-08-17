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

func ListCyclesRoutes() []string { return []string{"/hr/performance/cycles"} }

type ListCyclesHandler struct{ Service *application.Service }

type ListCyclesRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListCyclesResponse struct {
	Cycles []*hrperformancev1.ReviewCyclePayload `json:"cycles"`
	Total  int64                                 `json:"total"`
	Limit  int32                                 `json:"limit"`
	Offset int32                                 `json:"offset"`
}

func (h *ListCyclesHandler) Serve(ctx context.Context, req *ListCyclesRequest) (ListCyclesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListCyclesResponse{}, onekiterrors.PermissionDenied()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	cycles, total, err := h.Service.ListReviewCycles(ctx, device.User.UserID, domain.ListCyclesParams{
		Limit:  limit,
		Offset: req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list review cycles", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListCyclesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list review cycles")
	}

	payloads := make([]*hrperformancev1.ReviewCyclePayload, 0, len(cycles))
	for _, c := range cycles {
		payloads = append(payloads, cycleToProto(c))
	}

	return ListCyclesResponse{Cycles: payloads, Total: total, Limit: limit, Offset: req.Offset}, nil
}
