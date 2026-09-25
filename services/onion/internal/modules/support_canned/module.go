package supportcanned

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/support_canned/internal/api"
	"github.com/voxire/onion/internal/modules/support_canned/internal/application"
	supportcannedv1 "github.com/voxire/voxie/gen/go/support/canned/v1"
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
	if err := supportcannedv1.RegisterCannedResponseServiceServer(
		api.NewCannedResponseAPI(m.service),
		supportcannedv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register support canned service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
