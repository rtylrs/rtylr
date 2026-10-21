package booking_appointments

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/booking/appointments/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/booking/appointments/internal/api"
	"github.com/voxire/onion/internal/modules/booking/appointments/internal/application"
	bookingappointmentv1 "github.com/voxire/voxie/gen/go/booking/appointment/v1"
)

type Module struct{ service *application.Service }

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := bookingappointmentv1.RegisterAppointmentServiceServer(
		api.NewAppointmentAPI(m.service),
		bookingappointmentv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register booking appointment service: %w", err))
	}
}
func (m *Module) GetAnonymousRoutes() []string { return []string{} }
