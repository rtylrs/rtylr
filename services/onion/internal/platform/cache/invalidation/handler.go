package invalidation

import (
	"context"
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
	if ok := InvalidateCacheGroup(ctx, req.InvalidationType, req.IDs); !ok {
		return Response{}, onekiterrors.New(http.StatusBadRequest, "unknown cache invalidation group")
	}

	return Response{}, nil
}
