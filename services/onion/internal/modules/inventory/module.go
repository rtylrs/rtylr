package inventory

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/inventory/internal/api"
	"github.com/voxire/onion/internal/modules/inventory/internal/application"
	erpinventoryv1 "github.com/voxire/voxie/gen/go/erp/inventory/v1"
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
	if err := erpinventoryv1.RegisterInventoryServiceServer(
		api.NewInventoryAPI(m.service),
		erpinventoryv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register ERP inventory service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
