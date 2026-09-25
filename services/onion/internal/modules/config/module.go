package config

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/config/internal/api"
	"github.com/voxire/onion/internal/modules/config/internal/application"
	custom_domains "github.com/voxire/onion/internal/modules/custom_domains"
	hubconfigv1 "github.com/voxire/voxie/gen/go/hub/config/v1"
)

type Module struct {
	ConfigService *application.Service
	CustomDomains *custom_domains.API
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	configService := application.NewService(ctx, db)
	customDomainService := custom_domains.NewService(ctx, db)

	return &Module{
		ConfigService: configService,
		CustomDomains: custom_domains.NewAPI(customDomainService),
	}
}
func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubconfigv1.RegisterConfigServiceServer(
		api.NewConfigAPI(m.ConfigService, m.CustomDomains),
		hubconfigv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub config service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/hub/config/v1/industries",
	}
}
