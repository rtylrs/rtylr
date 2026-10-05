package http

import (
	"context"
	"fmt"
	stdhttp "net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/logger"
)

type APIRegistrar interface {
	RegisterAPI(*stdhttp.ServeMux)
}

type RouteRegistrar interface {
	RegisterRoutes(context.Context, *chi.Mux, ...interface{})
}

type AnonymousRouteProvider interface {
	GetAnonymousRoutes() []string
}

func AutoRegisterAPIs(ctx context.Context, mux *stdhttp.ServeMux, modules ...APIRegistrar) {
	for _, module := range modules {
		logger.Ctx(ctx).Debug("Registering oneKit services", logger.String("module", fmt.Sprintf("%T", module)))
		module.RegisterAPI(mux)
	}
}

func AutoRegisterRoutes(ctx context.Context, router *chi.Mux, modules ...RouteRegistrar) {
	for _, module := range modules {
		logger.Ctx(ctx).Debug("Registering routes", logger.String("module", fmt.Sprintf("%T", module)))
		module.RegisterRoutes(ctx, router)
	}
}

func AutoRegisterAnonymousRoutes(ctx context.Context, accessPolicy *RouteAccessPolicy, modules ...AnonymousRouteProvider) {
	for _, module := range modules {
		logger.Ctx(ctx).Debug("Registering anonymous routes", logger.String("module", fmt.Sprintf("%T", module)))
		accessPolicy.RegisterAnonymousRoutes(module.GetAnonymousRoutes()...)
	}
}
