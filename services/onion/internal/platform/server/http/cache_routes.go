package http

import (
	"context"
	"encoding/json"
	"net/http"

	cacheInvalidate "github.com/voxire/onion/internal/platform/cache/invalidate"
	"github.com/voxire/onion/internal/platform/cache/invalidation"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type cacheOKResponse struct {
	Status string `json:"status"`
}

func RegisterCacheRoutes(ctx context.Context, mux *http.ServeMux) {
	cacheInvalidate.RegisterDefaultGroups()

	mux.HandleFunc("POST /invalidate-cache", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !hasValidInternalToken(r) {
			WriteOneKitError(ctx, w, onekiterrors.New(http.StatusForbidden, "Forbidden"))
			return
		}

		var req invalidation.Request
		if err := DecodeJSONRequestBody(w, r, &req); err != nil {
			if isRequestBodyTooLarge(err) {
				WriteOneKitError(ctx, w, onekiterrors.New(http.StatusRequestEntityTooLarge, "Request body too large"))
				return
			}
			WriteOneKitError(ctx, w, onekiterrors.New(http.StatusBadRequest, "Invalid request body"))
			return
		}

		handler := invalidation.Handler{}

		_, serverErr := handler.Serve(ctx, &req)
		if serverErr != nil {
			WriteOneKitError(ctx, w, serverErr)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(cacheOKResponse{Status: "ok"}); err != nil {
			logger.Ctx(ctx).Warn("Failed to encode success response", logger.Error(err))
		}
	})
}
