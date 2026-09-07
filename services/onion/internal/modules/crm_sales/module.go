package crmsales

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/api"
	"github.com/voxire/onion/internal/modules/crm_sales/internal/application"
	crmsalesv1 "github.com/voxire/voxie/gen/go/crm/sales/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

// RegisterAPI mounts the generated crm.sales.v1 contract service on the
// shared oneKit mux (#809). The legacy chi routes stay until every caller is
// switched to the generated client.
func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := crmsalesv1.RegisterContractServiceServer(
		api.NewContractAPI(m.service),
		crmsalesv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register crm sales service: %w", err))
	}
}
