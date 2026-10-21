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
	if req == nil {
		req = &ListCyclesRequest{}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListCyclesResponse{}, onekiterrors.PermissionDenied()
	}

	page, err := normalizePerformancePagination(req.Limit, req.Offset)
	if err != nil {
		return ListCyclesResponse{}, err
	}
	limit, offset := page.limit, page.offset

	cycles, total, err := h.Service.ListReviewCycles(ctx, businessID, domain.ListCyclesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list review cycles", logger.Any("business_id", businessID), logger.Error(err))
		return ListCyclesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list review cycles")
	}

	payloads := make([]*hrperformancev1.ReviewCyclePayload, 0, len(cycles))
	for _, c := range cycles {
		payloads = append(payloads, cycleToProto(c))
	}

	return ListCyclesResponse{Cycles: payloads, Total: total, Limit: limit, Offset: offset}, nil
}
