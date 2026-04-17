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

func SubmitReviewRoutes() []string { return []string{"/hr/performance/reviews/{id}/submit"} }

type SubmitReviewHandler struct{ Service *application.Service }

type SubmitReviewRequest struct {
	ID        int64  `json:"id"`
	Rating    string `json:"rating"`
	Narrative string `json:"narrative"`
}

type SubmitReviewResponse struct {
	Success bool `json:"success"`
}

func (h *SubmitReviewHandler) Validate(_ context.Context, req *SubmitReviewRequest) *errorModels.ServerError {
	if req.Rating == "" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "rating is required"}
	}
	return nil
}

func (h *SubmitReviewHandler) Serve(ctx context.Context, req *SubmitReviewRequest) (SubmitReviewResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return SubmitReviewResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.Service.SubmitReview(ctx, device.User.UserID, req.ID, req.Rating, req.Narrative)
	if err != nil {
		if errors.Is(err, domain.ErrReviewNotFound) {
			return SubmitReviewResponse{}, &errorModels.ServerError{HTTPCode: http.StatusNotFound, Code: http.StatusNotFound, Message: "Review not found"}
		}
		if errors.Is(err, domain.ErrReviewAlreadySubmitted) {
			return SubmitReviewResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Review has already been submitted"}
		}
		if errors.Is(err, domain.ErrInvalidRating) {
			return SubmitReviewResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Rating is invalid for the configured scale"}
		}
		logger.Ctx(ctx).Error("failed to submit review", logger.Any("owner_id", device.User.UserID), logger.Any("review_id", req.ID), logger.Error(err))
		return SubmitReviewResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to submit review"}
	}

	return SubmitReviewResponse{Success: true}, nil
}
