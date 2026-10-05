package warehouses

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/erp/warehouses/internal/api"
	"github.com/voxire/onion/internal/modules/erp/warehouses/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/erp/warehouses/internal/infrastructure"
	erpwarehousev1 "github.com/voxire/voxie/gen/go/erp/warehouse/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(moduleinfra.NewRepository(db)),
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
