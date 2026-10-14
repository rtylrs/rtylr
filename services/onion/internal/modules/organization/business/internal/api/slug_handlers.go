package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/organization/business/internal/application"
	"github.com/voxire/onion/internal/modules/organization/business/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type SetSlugHandler struct{ Service *application.Service }

type SetSlugRequest struct {
	Slug string `json:"slug"`
}

type SetSlugResponse struct {
	Slug string `json:"slug"`
}

func (h *SetSlugHandler) Serve(ctx context.Context, req *SetSlugRequest) (SetSlugResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return SetSlugResponse{}, onekiterrors.PermissionDenied()
	}
	businessID := device.User.BusinessID
	if businessID == 0 {
		return SetSlugResponse{}, onekiterrors.New(http.StatusBadRequest, "No business attached to this user")
	}
	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	if domain.ReservedBusinessSlug(slug) {
		return SetSlugResponse{}, onekiterrors.New(http.StatusConflict, "This slug is reserved.")
	}
	if err := h.Service.UpdateSlug(ctx, businessID, device.User.UserID, slug); err != nil {
		if errors.Is(err, application.ErrBusinessNotOwned) {
			return SetSlugResponse{}, onekiterrors.PermissionDenied()
		}
		if errors.Is(err, application.ErrSlugReserved) {
			return SetSlugResponse{}, onekiterrors.New(http.StatusConflict, "This slug is reserved.")
		}
		if strings.Contains(err.Error(), "Duplicate") {
			return SetSlugResponse{}, onekiterrors.New(http.StatusConflict, "That slug is taken.")
		}
		logger.Ctx(ctx).Error("set slug failed",
			logger.Any("business_id", businessID),
			logger.String("slug", slug),
			logger.Error(err),
		)
		return SetSlugResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update slug")
	}
	return SetSlugResponse{Slug: slug}, nil
}
