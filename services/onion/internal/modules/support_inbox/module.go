package supportinbox

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/support_inbox/internal/api"
	"github.com/voxire/onion/internal/modules/support_inbox/internal/application"
	supportinboxv1 "github.com/voxire/voxie/gen/go/support/inbox/v1"
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
