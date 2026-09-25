package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrperformancev1 "github.com/voxire/voxie/gen/go/hr/performance/v1"
)

func ListReviewsRoutes() []string { return []string{"/hr/performance/reviews"} }

type ListReviewsHandler struct{ Service *application.Service }

type ListReviewsRequest struct {
	CycleID    int64 `json:"cycle_id"`
	EmployeeID int64 `json:"employee_id"`
	Limit      int32 `json:"limit"`
	Offset     int32 `json:"offset"`
}

type ListReviewsResponse struct {
	Reviews []*hrperformancev1.PerformanceReviewPayload `json:"reviews"`
	Total   int64                                       `json:"total"`
	Limit   int32                                       `json:"limit"`
	Offset  int32                                       `json:"offset"`
}

func (h *ListReviewsHandler) Serve(ctx context.Context, req *ListReviewsRequest) (ListReviewsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListReviewsResponse{}, onekiterrors.PermissionDenied()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListReviewsResponse{}, onekiterrors.PermissionDenied()
	}

	reviews, total, err := h.Service.ListReviews(ctx, businessID, domain.ListReviewsParams{
		CycleID:    req.CycleID,
		EmployeeID: req.EmployeeID,
		Limit:      limit,
		Offset:     req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list reviews", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListReviewsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list reviews")
	}

	payloads := make([]*hrperformancev1.PerformanceReviewPayload, 0, len(reviews))
	for _, r := range reviews {
		payloads = append(payloads, reviewToProto(r))
	}

	return ListReviewsResponse{Reviews: payloads, Total: total, Limit: limit, Offset: req.Offset}, nil
}
