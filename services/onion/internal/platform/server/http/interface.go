package http

import (
	"context"
	"net/http"
)

// RunMode registers a server's routes on the root ServeMux.
type RunMode interface {
	Setup(ctx context.Context, mux *http.ServeMux)
}

// MiddlewareProvider is implemented by runmodes that contribute HTTP
// middlewares. The server applies the platform middlewares and these together
// as a wrapper around the bare router, which lets a keyed internal bypass send
// requests straight to the router without running any middleware.
type MiddlewareProvider interface {
	Middlewares() []MiddlewareFunc
}

type DatabaseHealthMode interface {
	UsesDatabaseHealthCheck() bool
}
