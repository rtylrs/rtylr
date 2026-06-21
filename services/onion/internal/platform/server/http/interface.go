package http

import (
	"context"

	"github.com/go-chi/chi/v5"
)

type ChiRunMode interface {
	Setup(ctx context.Context, r *chi.Mux)
}

type DatabaseHealthMode interface {
	UsesDatabaseHealthCheck() bool
}
