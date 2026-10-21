package supporttickets

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/support/tickets/internal/api"
	"github.com/voxire/onion/internal/modules/support/tickets/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/support/tickets/internal/infrastructure"
	supportticketsv1 "github.com/voxire/voxie/gen/go/support/tickets/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, moduleinfra.NewRepository(db)),
	}
}

// RegisterAPI serves support.tickets.v1.
func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := supportticketsv1.RegisterTicketServiceServer(api.NewTicketAPI(m.service), supportticketsv1.WithMux(mux)); err != nil {
		panic(fmt.Errorf("register support ticket service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
