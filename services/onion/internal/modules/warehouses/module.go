package warehouses

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/warehouses/internal/api"
	"github.com/voxire/onion/internal/modules/warehouses/internal/application"
	erpwarehousev1 "github.com/voxire/voxie/gen/go/erp/warehouse/v1"
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
	if err := erpwarehousev1.RegisterWarehouseServiceServer(
		api.NewWarehouseAPI(m.service),
		erpwarehousev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register ERP warehouse service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
