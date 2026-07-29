package booking_dashboard

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/booking_dashboard/internal/api"
	"github.com/voxire/onion/internal/modules/booking_dashboard/internal/application"
	bookingdashboardv1 "github.com/voxire/voxie/gen/go/booking/dashboard/v1"
)

type Module struct{ service *application.Service }

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := bookingdashboardv1.RegisterDashboardServiceServer(
		api.NewDashboardAPI(m.service),
		bookingdashboardv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register booking dashboard service: %w", err))
	}
}
func (m *Module) GetAnonymousRoutes() []string { return []string{} }
