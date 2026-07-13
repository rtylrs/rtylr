package booking_appointments

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	bookingappointmentv1 "github.com/voxire/voxie/gen/go/booking/appointment/v1"

	"github.com/voxire/onion/internal/modules/booking_appointments/internal/api"
	"github.com/voxire/onion/internal/modules/booking_appointments/internal/application"
)

type Module struct{ service *application.Service }

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
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
