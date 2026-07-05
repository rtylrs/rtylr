package hr_dashboard

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	hrdashboardv1 "github.com/voxire/voxie/gen/go/hr/dashboard/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/api"
	"github.com/voxire/onion/internal/modules/hr_dashboard/internal/application"
	"github.com/voxire/onion/internal/modules/hr_dashboard/stats"
)

type Module struct {
	service   *application.Service
	loadStats stats.Loader
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
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
		hrdashboardv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register HR dashboard service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
