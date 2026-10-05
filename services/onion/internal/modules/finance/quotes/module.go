package quotes

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/finance/quotes/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/finance/quotes/internal/api"
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/application"
	financequotev1 "github.com/voxire/voxie/gen/go/finance/quote/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financequotev1.RegisterQuoteServiceServer(
		api.NewQuoteAPI(m.service),
		financequotev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register finance quote service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
