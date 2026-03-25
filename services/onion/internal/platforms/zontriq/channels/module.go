package channels

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/platforms/zontriq/channels/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/channels/internal/routes"
)

type Module struct {
	channelService *application.Service
}

func NewModule(ctx context.Context) *Module {
	channelService := application.NewService(ctx)

	return &Module{
		channelService: channelService,
	}
}

func (m *Module) GetService() *application.Service {
	return m.channelService
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.channelService)
}
