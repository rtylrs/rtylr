package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/platform/cache/invalidate"
	"github.com/voxire/onion/internal/platform/cache/keys"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// InvalidateRoutes is the public path for cache invalidation.
func InvalidateRoutes() []string { return []string{"/invalidate-cache"} }

// InvalidateHandler exposes a typed, scope-based cache wipe. Body:
//
//	{ "type": "user", "id": "1" }
//
// Returns the matched patterns and the number of keys removed.
type InvalidateHandler struct{}

type InvalidateRequest struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type InvalidateResponse struct {
	Type     string   `json:"type"`
	ID       string   `json:"id"`
	Patterns []string `json:"patterns"`
	Deleted  int      `json:"deleted"`
}

// Validate enforces a non-empty type from the known set + a non-empty id
// (except for the special "all" scope).
func (h *InvalidateHandler) Validate(_ context.Context, req *InvalidateRequest) *errorModels.ServerError {
	if req.Type == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "type is required",
		}
	}
	if !keys.IsKnownInvalidationType(req.Type) {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "unknown invalidation type",
		}
	}
	if req.ID == "" && keys.InvalidationType(req.Type) != keys.InvalidateAll {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "id is required",
		}
	}
	return nil
}

func (h *InvalidateHandler) Serve(
	ctx context.Context,
	req *InvalidateRequest,
) (InvalidateResponse, *errorModels.ServerError) {
	res := invalidate.Invalidate(ctx, keys.InvalidationType(req.Type), req.ID)
	return InvalidateResponse{
		Type:     string(res.Type),
		ID:       res.ID,
		Patterns: res.Patterns,
		Deleted:  res.Deleted,
	}, nil
}
