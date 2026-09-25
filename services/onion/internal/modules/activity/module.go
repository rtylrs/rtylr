package activity

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/activity/internal/api"
	"github.com/voxire/onion/internal/modules/activity/internal/application"
	hubactivityv1 "github.com/voxire/voxie/gen/go/hub/activity/v1"
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
	if err := hubactivityv1.RegisterActivityServiceServer(
		api.NewActivityAPI(m.service),
		hubactivityv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub activity service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
