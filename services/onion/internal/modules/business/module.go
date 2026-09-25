package business

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/business/internal/api"
	"github.com/voxire/onion/internal/modules/business/internal/application"
	usermodule "github.com/voxire/onion/internal/modules/user"
	hubbusinessv1 "github.com/voxire/voxie/gen/go/hub/business/v1"
)

type Module struct {
	service     *application.Service
	userService usermodule.OnboardingPort
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) WithUserService(userService usermodule.OnboardingPort) *Module {
	m.userService = userService
	return m
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	businessAPI := api.NewBusinessAPI(m.service)
	if m.userService != nil {
		businessAPI.WithUserService(m.userService)
	}
	if err := hubbusinessv1.RegisterBusinessServiceServer(
		businessAPI,
		hubbusinessv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub business service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/hub/business/v1/businesses/by-slug",
		"/hub/business/v1/businesses/public",
		"/hub/business/v1/businesses/name-exists",
	}
}
