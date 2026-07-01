package invalidation

import (
	"context"
	"net/http"

	errorModels "github.com/voxire/onion/internal/platform/errors"
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

func (c Handler) Serve(ctx context.Context, req *Request) (Response, *errorModels.ServerError) {
	if ok := InvalidateCacheGroup(ctx, req.InvalidationType, req.IDs); !ok {
		return Response{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "unknown cache invalidation group",
		}
	}

	return Response{}, nil
}
