package invalidation

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/platform/models"
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

func (c Handler) Serve(ctx context.Context, req *Request) (Response, *models.ServerError) {
	if ok := InvalidateCacheGroup(ctx, req.InvalidationType, req.IDs); !ok {
		return Response{}, &models.ServerError{
			Code:    http.StatusBadRequest,
			Message: "unknown cache invalidation group",
		}
	}

	return Response{}, nil
}
