package console

import (
	"context"
	"database/sql"
	"fmt"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"net/http"

	"github.com/voxire/onion/internal/modules/organization/console/internal/api"
	"github.com/voxire/onion/internal/modules/organization/console/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/organization/console/internal/infrastructure"
	consoleadminv1 "github.com/voxire/voxie/gen/go/console/admin/v1"
)

type Module struct {
	service *application.Service
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := consoleadminv1.RegisterConsoleAdminServiceServer(
		api.NewConsoleAdminAPI(m.service),
		consoleadminv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register console admin service: %w", err))
	}
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	svc := application.NewService(ctx, moduleinfra.NewRepository(db))
	return &Module{service: svc}
}

func (m *Module) SetEmployeeProvisioner(provisioner application.EmployeeProvisioner) {
	if provisioner != nil {
		m.service.WithEmployeeProvisioner(provisioner)
	}
}

func (m *Module) SetSeatChangeRecorder(recorder application.SeatChangeRecorder) {
	if recorder != nil {
		m.service.WithSeatChangeRecorder(recorder)
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router serverHttp.LegacyRouter, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return nil
}
