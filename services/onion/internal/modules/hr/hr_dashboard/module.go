package hr_dashboard

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_dashboard/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hr/hr_dashboard/internal/api"
	"github.com/voxire/onion/internal/modules/hr/hr_dashboard/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_dashboard/stats"
	hrdashboardv1 "github.com/voxire/voxie/gen/go/hr/dashboard/v1"
)

type Module struct {
	service   *application.Service
	loadStats stats.Loader
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) WithStatsLoader(loadStats stats.Loader) *Module {
	m.loadStats = loadStats
	return m
}

func (m *Module) GetService() *application.Service {
	return m.service
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrdashboardv1.RegisterDashboardServiceServer(
		api.NewDashboardAPI(m.loadStats, m.service),
		hrdashboardv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register HR dashboard service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
