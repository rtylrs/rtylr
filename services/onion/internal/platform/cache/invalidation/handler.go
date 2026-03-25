package invalidation

import (
	"context"

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
	InvalidateCacheGroup(ctx, req.InvalidationType, req.IDs)

	return Response{}, nil
}
