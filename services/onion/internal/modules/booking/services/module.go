package booking_services

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/booking/services/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/booking/services/internal/api"
	"github.com/voxire/onion/internal/modules/booking/services/internal/application"
	bookingservicev1 "github.com/voxire/voxie/gen/go/booking/service/v1"
)

type Module struct{ service *application.Service }

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := bookingservicev1.RegisterBookingServiceServiceServer(
		api.NewBookingServiceAPI(m.service),
		bookingservicev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register booking service service: %w", err))
	}
}
func (m *Module) GetAnonymousRoutes() []string { return []string{} }
