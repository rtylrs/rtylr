package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/status"
	"github.com/voxire/onion/internal/platform/status/providers"
)

func RegisterStatusRoutes(ctx context.Context, r *chi.Mux) {
	// Initialize status providers
	initializeStatusProviders()

	r.Get("/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		systemStatus := status.GetSystemStatus(ctx)
		statusCode := http.StatusOK

		// Return 503 if overall status is unhealthy
		if systemStatus.OverallStatus != "healthy" {
			statusCode = http.StatusServiceUnavailable
		}

		w.WriteHeader(statusCode)

		if err := json.NewEncoder(w).Encode(systemStatus); err != nil {
			logger.Ctx(ctx).Warn("Failed to encode status response", logger.Error(err))
		}
	})
}

func initializeStatusProviders() {
	status.RegisterProvider(providers.NewCacheStatusProvider())
	status.RegisterProvider(providers.NewDatabaseStatusProvider())
	status.RegisterProvider(providers.NewTrinoStatusProvider())
}
