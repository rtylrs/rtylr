package terminals

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/pos/terminals/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/pos/terminals/internal/api"
	"github.com/voxire/onion/internal/modules/pos/terminals/internal/application"
	posterminalv1 "github.com/voxire/voxie/gen/go/pos/terminal/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posterminalv1.RegisterTerminalServiceServer(
		api.NewTerminalAPI(m.service),
		posterminalv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS terminal service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
