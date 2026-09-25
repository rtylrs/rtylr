package pos_loyalty

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/api"
	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/application"
	posloyaltyv1 "github.com/voxire/voxie/gen/go/pos/loyalty/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posloyaltyv1.RegisterLoyaltyServiceServer(
		api.NewLoyaltyAPI(m.service),
		posloyaltyv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register POS loyalty service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
