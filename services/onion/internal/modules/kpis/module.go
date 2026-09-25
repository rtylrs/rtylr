package kpis

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/kpis/internal/api"
	"github.com/voxire/onion/internal/modules/kpis/internal/application"
	hubdashboardv1 "github.com/voxire/voxie/gen/go/hub/dashboard/v1"
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
	if err := hubdashboardv1.RegisterDashboardServiceServer(
		api.NewDashboardAPI(m.service),
		hubdashboardv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub dashboard service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
