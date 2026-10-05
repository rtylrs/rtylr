package sessions

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/pos/sessions/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/pos/sessions/internal/api"
	"github.com/voxire/onion/internal/modules/pos/sessions/internal/application"
	possessionv1 "github.com/voxire/voxie/gen/go/pos/session/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := possessionv1.RegisterSessionServiceServer(
		api.NewSessionAPI(m.service),
		possessionv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS session service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
