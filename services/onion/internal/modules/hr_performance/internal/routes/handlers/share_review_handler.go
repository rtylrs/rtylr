package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func ShareReviewRoutes() []string { return []string{"/hr/performance/reviews/{id}/share"} }

type ShareReviewHandler struct{ Service *application.Service }

type ShareReviewRequest struct {
	ID int64 `json:"id"`
}

type ShareReviewResponse struct {
	Success bool `json:"success"`
}

func (h *ShareReviewHandler) Serve(ctx context.Context, req *ShareReviewRequest) (ShareReviewResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ShareReviewResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.Service.ShareReviewWithEmployee(ctx, device.User.UserID, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrReviewNotFound) {
			return ShareReviewResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Review not found"}
		}
		if errors.Is(err, domain.ErrReviewNotSubmitted) {
			return ShareReviewResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Review must be submitted or completed before sharing"}
		}
		logger.Ctx(ctx).Error("failed to share review", logger.Any("owner_id", device.User.UserID), logger.Any("review_id", req.ID), logger.Error(err))
		return ShareReviewResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to share review"}
	}

	return ShareReviewResponse{Success: true}, nil
}
