package roles

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/roles/internal/api"
	"github.com/voxire/onion/internal/modules/roles/internal/application"
	hubpermissionv1 "github.com/voxire/voxie/gen/go/hub/permission/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubpermissionv1.RegisterPermissionServiceServer(
		api.NewPermissionAPI(m.service),
		hubpermissionv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub permission service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
