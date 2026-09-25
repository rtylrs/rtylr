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

func ShareReviewRoutes() []string { return []string{"/hr/performance/reviews/{id}/share"} }

type ShareReviewHandler struct{ Service *application.Service }

type ShareReviewRequest struct {
	ID int64 `json:"id"`
}

type ShareReviewResponse struct {
	Success bool `json:"success"`
}

func (h *ShareReviewHandler) Serve(ctx context.Context, req *ShareReviewRequest) (ShareReviewResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ShareReviewResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ShareReviewResponse{}, onekiterrors.PermissionDenied()
	}

	err := h.Service.ShareReviewWithEmployee(ctx, businessID, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrReviewNotFound) {
			return ShareReviewResponse{}, onekiterrors.New(http.StatusNotFound, "Review not found")
		}
		if errors.Is(err, domain.ErrReviewNotSubmitted) {
			return ShareReviewResponse{}, onekiterrors.New(http.StatusBadRequest, "Review must be submitted or completed before sharing")
		}
		logger.Ctx(ctx).Error("failed to share review", logger.Any("owner_id", device.User.UserID), logger.Any("review_id", req.ID), logger.Error(err))
		return ShareReviewResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to share review")
	}

	return ShareReviewResponse{Success: true}, nil
}
