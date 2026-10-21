package combos

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/pos/combos/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/pos/combos/internal/api"
	"github.com/voxire/onion/internal/modules/pos/combos/internal/application"
	poscombov1 "github.com/voxire/voxie/gen/go/pos/combo/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := poscombov1.RegisterComboServiceServer(
		api.NewComboAPI(m.service),
		poscombov1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS combo service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
