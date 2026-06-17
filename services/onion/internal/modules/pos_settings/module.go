package pos_settings

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_settings/internal/application"
	"github.com/voxire/onion/internal/modules/pos_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_settings/internal/routes"
)

const (
	ProfileRetail       = domain.ProfileRetail
	ProfileRestaurant   = domain.ProfileRestaurant
	ProfileQuickService = domain.ProfileQuickService
)

type Settings struct {
	BusinessID int64
	Profile    string
}

type Service struct {
	inner *application.Service
}

func NewService(ctx context.Context, db *sql.DB) *Service {
	return &Service{inner: application.NewService(ctx, db)}
}

func (s *Service) GetSettings(ctx context.Context, businessID int64) (Settings, error) {
	settings, err := s.inner.GetSettings(ctx, businessID)
	if err != nil {
		return Settings{}, err
	}
	return Settings{
		BusinessID: settings.BusinessID,
		Profile:    settings.Profile,
	}, nil
}

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
