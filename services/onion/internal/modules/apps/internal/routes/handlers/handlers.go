package handlers

import (
	"context"

	"github.com/voxire/onion/internal/modules/apps/internal/application"
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

func (h *ListAppsHandler) Serve(_ context.Context, _ *ListAppsRequest) (ListAppsResponse, *errorModels.ServerError) {
	apps := h.Service.ListApps()

	payloads := make([]AppPayload, 0, len(apps))
	for _, a := range apps {
		payloads = append(payloads, AppPayload{
			ID:          a.ID,
			Name:        a.Name,
			Description: a.Description,
			Icon:        a.Icon,
			IsEnabled:   a.IsEnabled,
		})
	}

	return ListAppsResponse{Apps: payloads}, nil
}
