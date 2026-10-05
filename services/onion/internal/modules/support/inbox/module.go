package supportinbox

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/support/inbox/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/support/inbox/internal/api"
	"github.com/voxire/onion/internal/modules/support/inbox/internal/application"
	supportinboxv1 "github.com/voxire/voxie/gen/go/support/inbox/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, moduleinfra.NewRepository(db)),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := supportinboxv1.RegisterInboxServiceServer(
		api.NewInboxAPI(m.service),
		supportinboxv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register support inbox service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
