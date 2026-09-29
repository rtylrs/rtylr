package api

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/voxire/onion/internal/modules/business/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func BusinessSlugRoutes() []string { return []string{"/business/me/slug"} }

func validateSlug(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return slugRe.MatchString(s)
}

var slugRe = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,78}[a-z0-9])?$`)

type SetSlugHandler struct{ Service *application.Service }

type SetSlugRequest struct {
	Slug string `json:"slug"`
}

type SetSlugResponse struct {
	Slug string `json:"slug"`
}

func (h *SetSlugHandler) Validate(_ context.Context, req *SetSlugRequest) error {
	if !validateSlug(req.Slug) {
		return onekiterrors.New(http.StatusBadRequest, "Slug must be 1–80 chars, lowercase a–z, 0–9, and hyphens.")
	}
	return nil
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
	if err := h.Service.UpdateSlug(ctx, businessID, device.User.UserID, slug); err != nil {
		if errors.Is(err, application.ErrBusinessNotOwned) {
			return SetSlugResponse{}, onekiterrors.PermissionDenied()
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
