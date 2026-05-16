package handlers

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/apps/internal/application"
	"github.com/voxire/onion/internal/modules/apps/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	hubv1 "github.com/voxire/voxie/gen/go/hub/v1"
)

func AppRoutes() []string { return []string{"/apps"} }

type (
	ListAppsHandler struct{ Service *application.Service }
	ListAppsRequest struct{}
)

// ListAppsResponse uses the proto-generated App type from voxie.
type ListAppsResponse struct {
	Apps []*hubv1.App `json:"apps"`
}

func (h *ListAppsHandler) Serve(ctx context.Context, _ *ListAppsRequest) (ListAppsResponse, *errorModels.ServerError) {
	if authUser, ok := authPkg.GetAuthUser(ctx); ok {
		res, err := cacher.Of[ListAppsResponse]().GetOrComputeSingle(
			ctx,
			cacheKeys.UserApps(authUser.UserID),
			func(ctx context.Context) (ListAppsResponse, error) {
				return h.computeApps(ctx), nil
			},
			cacher.WithExpiry(ctx, 5*time.Minute),
		)
		if err == nil {
			return res, nil
		}
	}
	return h.computeApps(ctx), nil
}

func (h *ListAppsHandler) computeApps(ctx context.Context) ListAppsResponse {
	allApps := h.Service.ListApps()

	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		payloads := make([]*hubv1.App, 0, len(allApps))
		for _, a := range allApps {
			if a.AddonGated == "" {
				payloads = append(payloads, toProtoApp(a, a.IsEnabled))
			}
		}
		return ListAppsResponse{Apps: payloads}
	}

	enabledAddons := h.Service.GetEnabledAddons(ctx, authUser.BusinessID)

	buildPayload := func(a domain.App) *hubv1.App {
		enabled := a.IsEnabled
		if a.AddonGated != "" {
			enabled = enabledAddons[a.AddonGated]
		}
		return toProtoApp(a, enabled)
	}

	if authUser.Role == "owner" || authUser.Role == "admin" {
		payloads := make([]*hubv1.App, 0, len(allApps))
		for _, a := range allApps {
			payloads = append(payloads, buildPayload(a))
		}
		return ListAppsResponse{Apps: payloads}
	}

	allowedApps := h.Service.GetUserAllowedApps(ctx, authUser.BusinessID, authUser.UserID)
	allowed := make(map[string]bool)
	for _, a := range allowedApps {
		allowed[a] = true
	}

	payloads := make([]*hubv1.App, 0)
	for _, a := range allApps {
		if allowed[a.ID] {
			payloads = append(payloads, buildPayload(a))
		}
	}

	return ListAppsResponse{Apps: payloads}
}

func toProtoApp(a domain.App, enabled bool) *hubv1.App {
	return &hubv1.App{
		Id:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		Icon:        a.Icon,
		IsEnabled:   enabled,
	}
}
