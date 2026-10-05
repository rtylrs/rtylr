package booking_calendar

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/booking/calendar/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/booking/calendar/internal/api"
	"github.com/voxire/onion/internal/modules/booking/calendar/internal/application"
	bookingcalendarv1 "github.com/voxire/voxie/gen/go/booking/calendar/v1"
)

type Module struct{ service *application.Service }

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := bookingcalendarv1.RegisterCalendarServiceServer(
		api.NewCalendarAPI(m.service),
		bookingcalendarv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register booking calendar service: %w", err))
	}
}
func (m *Module) GetAnonymousRoutes() []string { return []string{} }
