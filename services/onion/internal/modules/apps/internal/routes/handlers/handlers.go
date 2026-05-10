package handlers

import (
	"context"

	"github.com/voxire/onion/internal/modules/apps/internal/application"
	"github.com/voxire/onion/internal/modules/apps/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
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
	allApps := h.Service.ListApps()

	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		payloads := make([]*hubv1.App, 0, len(allApps))
		for _, a := range allApps {
			if a.AddonGated == "" {
				payloads = append(payloads, toProtoApp(a, a.IsEnabled))
			}
		}
		return ListAppsResponse{Apps: payloads}, nil
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
		return ListAppsResponse{Apps: payloads}, nil
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

	return ListAppsResponse{Apps: payloads}, nil
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
