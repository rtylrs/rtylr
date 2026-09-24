package invalidation

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func GetInvalidationRoutes() []string {
	return []string{
		"/invalidate-cache",
	}
}

type Request struct {
	InvalidationType string   `json:"type"`
	IDs              []string `json:"ids"`
}

type Response struct{}

type Handler struct{}

func (c Handler) Serve(ctx context.Context, req *Request) (Response, error) {
	if err := InvalidateCacheGroupWithError(ctx, req.InvalidationType, req.IDs); err != nil {
		if errors.Is(err, ErrUnknownCacheGroup) {
			return Response{}, onekiterrors.New(http.StatusBadRequest, "unknown cache invalidation group")
		}
		return Response{}, onekiterrors.New(http.StatusBadRequest, "invalid cache invalidation request")
	}

	return Response{}, nil
}
