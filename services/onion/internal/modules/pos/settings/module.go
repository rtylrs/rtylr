package settings

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/pos/settings/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/pos/settings/internal/api"
	"github.com/voxire/onion/internal/modules/pos/settings/internal/application"
	"github.com/voxire/onion/internal/modules/pos/settings/internal/domain"
	possettingsv1 "github.com/voxire/voxie/gen/go/pos/settings/v1"
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
	return &Service{inner: application.NewService(ctx, moduleinfra.NewRepository(db))}
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

func (s *Service) GetProfile(ctx context.Context, businessID int64) (string, error) {
	settings, err := s.GetSettings(ctx, businessID)
	if err != nil {
		return "", err
	}
	return settings.Profile, nil
}

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := possettingsv1.RegisterSettingsServiceServer(
		api.NewSettingsAPI(m.service),
		possettingsv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS settings service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
