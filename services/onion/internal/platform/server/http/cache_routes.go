package http

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/cache/invalidation"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

func RegisterCacheRoutes(ctx context.Context, r *chi.Mux) {
	r.Post("/invalidate-cache", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !hasValidInternalToken(r) {
			w.WriteHeader(http.StatusForbidden)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"error": "Forbidden"}); err != nil {
				logger.Ctx(ctx).Warn("Failed to encode forbidden response", logger.Error(err))
			}
			return
		}

		var req invalidation.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			response := map[string]interface{}{
				"error": "Invalid request body",
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				logger.Ctx(ctx).Warn("Failed to encode error response", logger.Error(err))
			}

			return
		}

		handler := invalidation.Handler{}

		_, serverErr := handler.Serve(ctx, &req)
		if serverErr != nil {
			// models.ServerError only carries Code (no HTTPCode), so
			// trust it directly when it's a valid 4xx/5xx and only
			// fall back to 500 when Code is unset / out of range.
			status := serverErr.Code
			if status < 400 || status >= 600 {
				status = http.StatusInternalServerError
			}
			w.WriteHeader(status)

			response := map[string]interface{}{
				"error": serverErr.Message,
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				logger.Ctx(ctx).Warn("Failed to encode error response", logger.Error(err))
			}

			return
		}

		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"status": "ok",
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Ctx(ctx).Warn("Failed to encode success response", logger.Error(err))
		}
	})
}

func hasValidInternalToken(r *http.Request) bool {
	var expected string
	if config.GlobalConfig != nil {
		expected = config.GlobalConfig.InternalToken
	} else {
		expected = os.Getenv("INTERNAL_TOKEN")
	}
	if expected == "" {
		return false
	}

	actual := r.Header.Get("Internal-Token")
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}
