package handlers

import (
	"context"

	"github.com/voxire/onion/internal/modules/apps/internal/application"
	"github.com/voxire/onion/internal/modules/apps/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func AppRoutes() []string { return []string{"/apps"} }

type ListAppsHandler struct{ Service *application.Service }
type ListAppsRequest struct{}

type AppPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	IsEnabled   bool   `json:"isEnabled"`
}

type ListAppsResponse struct {
	Apps []AppPayload `json:"apps"`
}

func (h *ListAppsHandler) Serve(ctx context.Context, _ *ListAppsRequest) (ListAppsResponse, *errorModels.ServerError) {
	allApps := h.Service.ListApps()

	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		// No auth context — return core apps only
		payloads := make([]AppPayload, 0, len(allApps))
		for _, a := range allApps {
			if a.AddonGated == "" {
				payloads = append(payloads, AppPayload{ID: a.ID, Name: a.Name, Description: a.Description, Icon: a.Icon, IsEnabled: a.IsEnabled})
			}
		}
		return ListAppsResponse{Apps: payloads}, nil
	}

	// Get addon enablement via service -> repo (no DB in handler)
	enabledAddons := h.Service.GetEnabledAddons(ctx, authUser.BusinessID)

	buildPayload := func(a domain.App) AppPayload {
		enabled := a.IsEnabled
		if a.AddonGated != "" {
			enabled = enabledAddons[a.AddonGated]
		}
		return AppPayload{ID: a.ID, Name: a.Name, Description: a.Description, Icon: a.Icon, IsEnabled: enabled}
	}

	if authUser.Role == "owner" || authUser.Role == "admin" {
		payloads := make([]AppPayload, 0, len(allApps))
		for _, a := range allApps {
			payloads = append(payloads, buildPayload(a))
		}
		return ListAppsResponse{Apps: payloads}, nil
	}

	// Staff: only return apps they have access to (via service -> repo)
	allowedApps := h.Service.GetUserAllowedApps(ctx, authUser.BusinessID, authUser.UserID)
	allowed := make(map[string]bool)
	for _, a := range allowedApps {
		allowed[a] = true
	}

	payloads := make([]AppPayload, 0)
	for _, a := range allApps {
		if allowed[a.ID] {
			payloads = append(payloads, buildPayload(a))
		}
	}

	return ListAppsResponse{Apps: payloads}, nil
}
