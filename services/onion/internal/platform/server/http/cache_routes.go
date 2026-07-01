package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	cacheInvalidate "github.com/voxire/onion/internal/platform/cache/invalidate"
	"github.com/voxire/onion/internal/platform/cache/invalidation"
	"github.com/voxire/onion/internal/platform/logger"
)

func RegisterCacheRoutes(ctx context.Context, r *chi.Mux) {
	cacheInvalidate.RegisterDefaultGroups()

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
		if err := DecodeJSONRequestBody(w, r, &req); err != nil {
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
			status := serverErr.HTTPCode
			if status < 400 || status >= 600 {
				status = serverErr.Code
			}
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
