package hr_attendance

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/api"
	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
	hrcore "github.com/voxire/onion/internal/modules/hr_core"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
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

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	attendanceAPI := api.NewAttendanceAPI(m.service, m.employeeResolver)
	serverHttp.RegisterGET[api.GetAttendanceRequest, api.GetAttendanceResponse](
		ctx, router, api.GetAttendanceRoutes(), &api.GetAttendanceHandler{
			Service: m.service,
			Resolve: attendanceAPI.ResolveIdentity,
		},
	)
	serverHttp.RegisterPUTWithValidation[api.SetClockInPINRequest, api.SetClockInPINResponse](
		ctx, router, api.SetClockInPINRoutes(), &api.SetClockInPINHandler{
			Service:          m.service,
			EmployeeResolver: m.employeeResolver,
		},
	)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrattendancev1.RegisterAttendanceServiceServer(
		api.NewAttendanceAPI(m.service, m.employeeResolver),
		hrattendancev1.WithMux(mux),
		hrattendancev1.WithMiddleware(api.AttendancePaginationMiddleware),
	); err != nil {
		panic(fmt.Errorf("register HR attendance service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
