package users

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/platforms/zontriq/users/adapter"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/routes"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/routes/handlers"
)

type Module struct {
	usersService *application.Service
	Adapter      *adapter.UserAdapter
}

func NewModule(ctx context.Context) *Module {
	usersService := application.NewService(ctx)

	return &Module{
		usersService: usersService,
		Adapter:      adapter.NewUserAdapter(usersService),
	}
}

func (m *Module) GetService() *application.Service {
	return m.usersService
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...any) {
	routes.RegisterRoutes(ctx, router, m.usersService)
}

func (m *Module) GetAnonymousRoutes() []string {
	anonRoutes := handlers.ListPlansRoutes()
	anonRoutes = append(anonRoutes, handlers.WaitlistRoutes()...)

	return anonRoutes
}
