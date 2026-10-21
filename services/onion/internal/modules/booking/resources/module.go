package booking_resources

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/booking/resources/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/booking/resources/internal/api"
	"github.com/voxire/onion/internal/modules/booking/resources/internal/application"
	bookingresourcev1 "github.com/voxire/voxie/gen/go/booking/resource/v1"
)

type Module struct{ service *application.Service }

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := bookingresourcev1.RegisterResourceServiceServer(
		api.NewResourceAPI(m.service),
		bookingresourcev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register booking resource service: %w", err))
	}
}
func (m *Module) GetAnonymousRoutes() []string { return []string{} }
