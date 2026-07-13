package products

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	erpproductv1 "github.com/voxire/voxie/gen/go/erp/product/v1"

	"github.com/voxire/onion/internal/modules/products/internal/api"
	"github.com/voxire/onion/internal/modules/products/internal/application"
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
	if err := erpproductv1.RegisterProductServiceServer(
		api.NewProductAPI(m.service),
		erpproductv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register ERP product service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
