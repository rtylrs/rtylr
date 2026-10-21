package cashmovements

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/pos/cashmovements/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/pos/cashmovements/internal/api"
	"github.com/voxire/onion/internal/modules/pos/cashmovements/internal/application"
	poscashv1 "github.com/voxire/voxie/gen/go/pos/cash/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := poscashv1.RegisterCashServiceServer(
		api.NewCashAPI(m.service),
		poscashv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS cash service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
