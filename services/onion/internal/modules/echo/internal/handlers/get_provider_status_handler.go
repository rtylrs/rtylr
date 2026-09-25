package handlers

import (
	"context"

	"github.com/voxire/onion/internal/modules/echo/internal/application"
)

func ProviderStatusRoutes() []string {
	return []string{
		"/ai/status",
	}
}

type ProviderStatusHandler struct {
	SnipsService *application.Service
}

type GetProviderStatusRequest struct{}

type GetProviderStatusResponse struct {
	Available       bool            `json:"available"`
	PrimaryProvider string          `json:"primary_provider"`
	Providers       map[string]bool `json:"providers"`
}

func (h *ProviderStatusHandler) Serve(ctx context.Context, _ *GetProviderStatusRequest) (GetProviderStatusResponse, error) {
	status := h.SnipsService.GetProviderDetails(ctx)

	return GetProviderStatusResponse{
		Available:       status.Available,
		PrimaryProvider: status.PrimaryProvider,
		Providers:       status.Providers,
	}, nil
}
