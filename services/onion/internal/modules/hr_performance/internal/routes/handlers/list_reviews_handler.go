package handlers

import (
	"context"
	"net/http"

	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
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
	Reviews []*hrv1.PerformanceReviewPayload `json:"reviews"`
	Total   int64                            `json:"total"`
	Limit   int32                            `json:"limit"`
	Offset  int32                            `json:"offset"`
}

func (h *ListReviewsHandler) Serve(ctx context.Context, req *ListReviewsRequest) (ListReviewsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListReviewsResponse{}, errorModels.PermissionDeniedServerError()
	}

	reviews, total, err := h.Service.ListReviews(ctx, device.User.UserID, domain.ListReviewsParams{
		CycleID:    req.CycleID,
		EmployeeID: req.EmployeeID,
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list reviews", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListReviewsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list reviews"}
	}

	payloads := make([]*hrv1.PerformanceReviewPayload, 0, len(reviews))
	for _, r := range reviews {
		payloads = append(payloads, reviewToProto(r))
	}

	return ListReviewsResponse{Reviews: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
