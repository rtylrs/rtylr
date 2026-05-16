package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/cache/invalidation"
	"github.com/voxire/onion/internal/platform/logger"
)

func RegisterCacheRoutes(ctx context.Context, r *chi.Mux) {
	r.Post("/invalidate-cache", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

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
			// Mirror writeServerError: fall back to Code (the app
			// status) when HTTPCode is unset, so a 400 "unknown
			// invalidation type" stops being reported as 500.
			status := serverErr.HTTPCode
			if status == 0 {
				if serverErr.Code >= 400 && serverErr.Code < 600 {
					status = serverErr.Code
				} else {
					status = http.StatusInternalServerError
				}
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
