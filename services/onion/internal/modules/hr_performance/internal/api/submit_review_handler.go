package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
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

func (h *SubmitReviewHandler) Validate(_ context.Context, req *SubmitReviewRequest) error {
	if req.Rating == "" {
		return onekiterrors.New(http.StatusBadRequest, "rating is required")
	}
	return nil
}

func (h *SubmitReviewHandler) Serve(ctx context.Context, req *SubmitReviewRequest) (SubmitReviewResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return SubmitReviewResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requirePerformanceManagePermission(ctx); err != nil {
		return SubmitReviewResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return SubmitReviewResponse{}, onekiterrors.PermissionDenied()
	}

	err := h.Service.SubmitReview(ctx, businessID, req.ID, req.Rating, req.Narrative)
	if err != nil {
		if errors.Is(err, domain.ErrReviewNotFound) {
			return SubmitReviewResponse{}, onekiterrors.New(http.StatusNotFound, "Review not found")
		}
		if errors.Is(err, domain.ErrReviewAlreadySubmitted) {
			return SubmitReviewResponse{}, onekiterrors.New(http.StatusBadRequest, "Review has already been submitted")
		}
		if errors.Is(err, domain.ErrInvalidRating) {
			return SubmitReviewResponse{}, onekiterrors.New(http.StatusBadRequest, "Rating is invalid for the configured scale")
		}
		logger.Ctx(ctx).Error("failed to submit review", logger.Any("owner_id", device.User.UserID), logger.Any("review_id", req.ID), logger.Error(err))
		return SubmitReviewResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to submit review")
	}

	return SubmitReviewResponse{Success: true}, nil
}
