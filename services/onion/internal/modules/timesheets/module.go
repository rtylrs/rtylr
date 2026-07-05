package timesheets

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	hrtimesheetv1 "github.com/voxire/voxie/gen/go/hr/timesheet/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/timesheets/internal/api"
	"github.com/voxire/onion/internal/modules/timesheets/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrtimesheetv1.RegisterTimesheetServiceServer(
		api.NewTimesheetAPI(m.service),
		hrtimesheetv1.WithMux(mux),
		hrtimesheetv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register timesheet service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
