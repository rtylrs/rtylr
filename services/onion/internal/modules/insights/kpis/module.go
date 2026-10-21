package kpis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/kpis/internal/api"
	"github.com/voxire/onion/internal/modules/insights/kpis/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/insights/kpis/internal/infrastructure"
	hubdashboardv1 "github.com/voxire/voxie/gen/go/hub/dashboard/v1"
)

type Module struct {
	service *application.Service
}

// Sources are the owning modules the KPI summary is read from.
type Sources = moduleinfra.Sources

func NewModule(ctx context.Context, sources Sources) *Module {
	return &Module{
		service: application.NewService(ctx, moduleinfra.NewRepository(sources)),
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
