package modifiers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/pos/modifiers/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/pos/modifiers/internal/api"
	"github.com/voxire/onion/internal/modules/pos/modifiers/internal/application"
	posmodifierv1 "github.com/voxire/voxie/gen/go/pos/modifier/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posmodifierv1.RegisterModifierServiceServer(
		api.NewModifierAPI(m.service),
		posmodifierv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS modifier service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
