package hr_attendance

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/api"
	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
	hrcore "github.com/voxire/onion/internal/modules/hr_core"
	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"
)

type Module struct {
	service          *application.Service
	employeeResolver api.EmployeeResolver
}

func NewModule(ctx context.Context, db *sql.DB, hrCoreModule *hrcore.Module) *Module {
	return &Module{
		service:          application.NewService(ctx, db),
		employeeResolver: hrCoreModule,
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrattendancev1.RegisterAttendanceServiceServer(
		api.NewAttendanceAPI(m.service, m.employeeResolver),
		hrattendancev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register HR attendance service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
