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

type cacheErrorResponse struct {
	Error string `json:"error"`
}

type cacheOKResponse struct {
	Status string `json:"status"`
}

func RegisterCacheRoutes(ctx context.Context, r *chi.Mux) {
	cacheInvalidate.RegisterDefaultGroups()

	r.Post("/invalidate-cache", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !hasValidInternalToken(r) {
			w.WriteHeader(http.StatusForbidden)
			if err := json.NewEncoder(w).Encode(cacheErrorResponse{Error: "Forbidden"}); err != nil {
				logger.Ctx(ctx).Warn("Failed to encode forbidden response", logger.Error(err))
			}
			return
		}

		var req invalidation.Request
		if err := DecodeJSONRequestBody(w, r, &req); err != nil {
			if isRequestBodyTooLarge(err) {
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				if encErr := json.NewEncoder(w).Encode(cacheErrorResponse{Error: "Request body too large"}); encErr != nil {
					logger.Ctx(ctx).Warn("Failed to encode error response", logger.Error(encErr))
				}
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(cacheErrorResponse{Error: "Invalid request body"}); err != nil {
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
			if err := json.NewEncoder(w).Encode(cacheErrorResponse{Error: serverErr.Message}); err != nil {
				logger.Ctx(ctx).Warn("Failed to encode error response", logger.Error(err))
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(cacheOKResponse{Status: "ok"}); err != nil {
			logger.Ctx(ctx).Warn("Failed to encode success response", logger.Error(err))
		}
	})
}
