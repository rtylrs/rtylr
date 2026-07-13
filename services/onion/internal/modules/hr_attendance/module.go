package hr_attendance

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/api"
	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
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
	if err := hrattendancev1.RegisterAttendanceServiceServer(
		api.NewAttendanceAPI(m.service),
		hrattendancev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register HR attendance service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/api/hr/attendance/v1/attendance/clock-in",
	}
}
