package tax

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/tax/internal/api"
	"github.com/voxire/onion/internal/modules/tax/internal/application"
	financetaxv1 "github.com/voxire/voxie/gen/go/finance/tax/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financetaxv1.RegisterTaxServiceServer(
		api.NewTaxAPI(m.service),
		financetaxv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance tax service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
