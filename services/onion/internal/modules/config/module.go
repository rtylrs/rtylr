package config

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/config/internal/api"
	"github.com/voxire/onion/internal/modules/config/internal/application"
	hubconfigv1 "github.com/voxire/voxie/gen/go/hub/config/v1"
)

type Module struct {
	ConfigService *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	configService := application.NewService(ctx, db)

	return &Module{
		ConfigService: configService,
	}
}

func (m *Module) GetService() *application.Service {
	return m.ConfigService
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubconfigv1.RegisterConfigServiceServer(
		api.NewConfigAPI(m.ConfigService),
		hubconfigv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub config service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/api/hub/config/v1/industries",
	}
}
