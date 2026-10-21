package security

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/organization/security/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/organization/security/internal/api"
	"github.com/voxire/onion/internal/modules/organization/security/internal/application"
	hubsecurityv1 "github.com/voxire/voxie/gen/go/hub/security/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, moduleinfra.NewRepository(db)),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubsecurityv1.RegisterSecurityServiceServer(
		api.NewSecurityAPI(m.service),
		hubsecurityv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub security service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
