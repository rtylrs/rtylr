package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/status"
	"github.com/voxire/onion/internal/platform/status/providers"
)

func RegisterStatusRoutes(ctx context.Context, r *chi.Mux) {
	// Initialize status providers
	initializeStatusProviders()

	r.Get("/status", func(w http.ResponseWriter, req *http.Request) {
		authUser, ok := authPkg.GetAuthUser(req.Context())
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		systemStatus := status.GetSystemStatus(req.Context())
		if !authPkg.IsConsoleAdmin(authUser) {
			redactStatusDetails(&systemStatus)
		}
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

func redactStatusDetails(systemStatus *status.SystemStatus) {
	if systemStatus == nil {
		return
	}
	for name, component := range systemStatus.Components {
		component.Message = ""
		component.Details = nil
		systemStatus.Components[name] = component
	}
}

func initializeStatusProviders() {
	status.RegisterProvider(providers.NewCacheStatusProvider())
	status.RegisterProvider(providers.NewDatabaseStatusProvider())
	status.RegisterProvider(providers.NewTrinoStatusProvider())
}
