package suppliers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/suppliers/internal/api"
	"github.com/voxire/onion/internal/modules/suppliers/internal/application"
	erpsupplierv1 "github.com/voxire/voxie/gen/go/erp/supplier/v1"
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
	if err := erpsupplierv1.RegisterSupplierServiceServer(
		api.NewSupplierAPI(m.service),
		erpsupplierv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register ERP supplier service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
