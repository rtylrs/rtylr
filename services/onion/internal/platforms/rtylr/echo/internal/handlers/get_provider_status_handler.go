package handlers

import (
	"context"

	errorModels "github.com/voxire/onion/internal/platform/errors"
	echov1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/rtylr/echo/v1"
	"github.com/voxire/onion/internal/platforms/rtylr/echo/internal/application"
)

func ProviderStatusRoutes() []string {
	return []string{
		"/ai/status",
	}
}

type ProviderStatusHandler struct {
	SnipsService *application.Service
}

func (h *ProviderStatusHandler) Serve(ctx context.Context, _ *echov1.GetProviderStatusRequest) (echov1.GetProviderStatusResponse, *errorModels.ServerError) {
	status := h.SnipsService.GetProviderDetails(ctx)

	return echov1.GetProviderStatusResponse{
		Available:       status.Available,
		PrimaryProvider: status.PrimaryProvider,
		Providers:       status.Providers,
	}, nil
}
