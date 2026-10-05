package timesheets

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/timesheets/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hr/timesheets/internal/api"
	"github.com/voxire/onion/internal/modules/hr/timesheets/internal/application"
	hrtimesheetv1 "github.com/voxire/voxie/gen/go/hr/timesheet/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrtimesheetv1.RegisterTimesheetServiceServer(
		api.NewTimesheetAPI(m.service),
		hrtimesheetv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register timesheet service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
