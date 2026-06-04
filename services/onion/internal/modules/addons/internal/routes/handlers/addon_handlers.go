package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/addons/internal/application"
	"github.com/voxire/onion/internal/modules/addons/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	hubappv1 "github.com/voxire/voxie/gen/go/hub/app/v1"
)

// addonsCacheScope namespaces the per-business addons list cache. The
// business_addons enable state is written only within this module
// (Enable/Disable), so a short TTL plus invalidation on those mutations
// keeps reads consistent.
const addonsCacheScope = "addons"

func invalidateAddonsCache(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(addonsCacheScope, businessID))
	}
}

// ─── List addons for the current business ───────────────────────────────────

func ListAddonsRoutes() []string { return []string{"/addons"} }

type (
	ListAddonsHandler struct{ Service *application.Service }
	ListAddonsRequest struct{}
)

type ListAddonsResponse struct {
	Addons []*hubappv1.Addon `json:"addons"`
}

func (h *ListAddonsHandler) Serve(ctx context.Context, _ *ListAddonsRequest) (ListAddonsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListAddonsResponse{}, errorModels.PermissionDeniedServerError()
	}

	res, cacheErr := cacher.Of[ListAddonsResponse]().GetOrComputeSingle(
		ctx,
		cacheKeys.Scoped(addonsCacheScope, authUser.BusinessID, "list"),
		func(ctx context.Context) (ListAddonsResponse, error) {
			return h.computeList(ctx, authUser.BusinessID)
		},
		cacher.WithExpiry(ctx, 3*time.Minute),
	)
	if cacheErr == nil {
		return res, nil
	}

	res, computeErr := h.computeList(ctx, authUser.BusinessID)
	if computeErr != nil {
		return ListAddonsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list addons",
		}
	}
	return res, nil
}

func (h *ListAddonsHandler) computeList(ctx context.Context, businessID int64) (ListAddonsResponse, error) {
	addons, err := h.Service.ListForBusiness(ctx, businessID)
	if err != nil {
		return ListAddonsResponse{}, err
	}

	payloads := make([]*hubappv1.Addon, 0, len(addons))
	for _, a := range addons {
		payloads = append(payloads, toProtoAddon(a))
	}

	return ListAddonsResponse{Addons: payloads}, nil
}

func toProtoAddon(a domain.BusinessAddon) *hubappv1.Addon {
	p := &hubappv1.Addon{
		Id:          a.AddonID,
		Name:        a.Name,
		Description: a.Description,
		Icon:        a.Icon,
		Price:       a.Price,
		IsFree:      a.IsFree,
		IsEnabled:   a.IsEnabled,
	}
	if a.EnabledAt != nil {
		s := a.EnabledAt.Format("2006-01-02T15:04:05Z07:00")
		p.EnabledAt = &s
	}
	if a.DisabledAt != nil {
		s := a.DisabledAt.Format("2006-01-02T15:04:05Z07:00")
		p.DisabledAt = &s
	}
	return p
}

// ─── Enable addon ───────────────────────────────────────────────────────────

func EnableAddonRoutes() []string { return []string{"/addons/enable"} }

type EnableAddonHandler struct{ Service *application.Service }

type EnableAddonRequest struct {
	AddonID string `json:"addon_id"`
}

type EnableAddonResponse struct {
	Message string `json:"message"`
}

func (h *EnableAddonHandler) Validate(_ context.Context, req *EnableAddonRequest) *errorModels.ServerError {
	if req.AddonID == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Addon ID is required"}
	}
	return nil
}

func (h *EnableAddonHandler) Serve(ctx context.Context, req *EnableAddonRequest) (EnableAddonResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return EnableAddonResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.Enable(ctx, authUser.BusinessID, req.AddonID); err != nil {
		return EnableAddonResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to enable addon",
		}
	}

	invalidateAddonsCache(ctx, authUser.BusinessID)

	return EnableAddonResponse{Message: "Addon enabled"}, nil
}

// ─── Disable addon ──────────────────────────────────────────────────────────

func DisableAddonRoutes() []string { return []string{"/addons/disable"} }

type DisableAddonHandler struct{ Service *application.Service }

type DisableAddonRequest struct {
	AddonID string `json:"addon_id"`
}

type DisableAddonResponse struct {
	Message string `json:"message"`
}

func (h *DisableAddonHandler) Validate(_ context.Context, req *DisableAddonRequest) *errorModels.ServerError {
	if req.AddonID == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Addon ID is required"}
	}
	return nil
}

func (h *DisableAddonHandler) Serve(ctx context.Context, req *DisableAddonRequest) (DisableAddonResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DisableAddonResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.Disable(ctx, authUser.BusinessID, req.AddonID); err != nil {
		return DisableAddonResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to disable addon",
		}
	}

	invalidateAddonsCache(ctx, authUser.BusinessID)

	return DisableAddonResponse{Message: "Addon disabled"}, nil
}

// ─── Check addon status ─────────────────────────────────────────────────────

func AddonStatusRoutes() []string { return []string{"/addons/status"} }

type AddonStatusHandler struct{ Service *application.Service }

type AddonStatusRequest struct {
	AddonID string `json:"addon_id"`
}

type AddonStatusResponse struct {
	IsEnabled bool `json:"is_enabled"`
}

func (h *AddonStatusHandler) Serve(ctx context.Context, req *AddonStatusRequest) (AddonStatusResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return AddonStatusResponse{}, errorModels.PermissionDeniedServerError()
	}

	enabled := h.Service.IsEnabled(ctx, authUser.BusinessID, req.AddonID)
	return AddonStatusResponse{IsEnabled: enabled}, nil
}

// ─── Unused import guard ────────────────────────────────────────────────────.
var _ domain.Addon
