package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/voxire/onion/internal/modules/business/internal/application"
	"github.com/voxire/onion/internal/modules/business/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func PublicBusinessBySlugRoutes() []string { return []string{"/pos/public/business/by-slug"} }
func BusinessSlugRoutes() []string         { return []string{"/business/me/slug"} }

type PublicBusinessBySlugHandler struct{ Service *application.Service }

type PublicBusinessBySlugRequest struct {
	Slug string `json:"slug"`
}

type PublicBusinessBySlugResponse struct {
	Business  domain.PublicBusiness   `json:"business"`
	Locations []domain.PublicLocation `json:"locations"`
}

var slugRe = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,78}[a-z0-9])?$`)

func validateSlug(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return slugRe.MatchString(s)
}

func (h *PublicBusinessBySlugHandler) Validate(_ context.Context, req *PublicBusinessBySlugRequest) *errorModels.ServerError {
	if !validateSlug(req.Slug) {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid slug",
		}
	}
	return nil
}

func (h *PublicBusinessBySlugHandler) Serve(ctx context.Context, req *PublicBusinessBySlugRequest) (PublicBusinessBySlugResponse, *errorModels.ServerError) {
	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	res, err := h.Service.LookupBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PublicBusinessBySlugResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Business not found",
			}
		}
		logger.Ctx(ctx).Error("slug lookup failed", logger.String("slug", slug), logger.Error(err))
		return PublicBusinessBySlugResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to lookup business",
		}
	}
	if res.Locations == nil {
		res.Locations = []domain.PublicLocation{}
	}
	return PublicBusinessBySlugResponse{Business: res.Business, Locations: res.Locations}, nil
}

type SetSlugHandler struct{ Service *application.Service }

type SetSlugRequest struct {
	Slug string `json:"slug"`
}

type SetSlugResponse struct {
	Slug string `json:"slug"`
}

func (h *SetSlugHandler) Validate(_ context.Context, req *SetSlugRequest) *errorModels.ServerError {
	if !validateSlug(req.Slug) {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Slug must be 1–80 chars, lowercase a–z, 0–9, and hyphens.",
		}
	}
	return nil
}

func (h *SetSlugHandler) Serve(ctx context.Context, req *SetSlugRequest) (SetSlugResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return SetSlugResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID := device.User.BusinessID
	if businessID == 0 {
		return SetSlugResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "No business attached to this user",
		}
	}
	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	if err := h.Service.UpdateSlug(ctx, businessID, slug); err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return SetSlugResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusConflict,
				Code:     http.StatusConflict,
				Message:  "That slug is taken.",
			}
		}
		logger.Ctx(ctx).Error("set slug failed",
			logger.Any("business_id", businessID),
			logger.String("slug", slug),
			logger.Error(err),
		)
		return SetSlugResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to update slug",
		}
	}
	return SetSlugResponse{Slug: slug}, nil
}
