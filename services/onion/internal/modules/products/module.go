package products

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/products/internal/api"
	"github.com/voxire/onion/internal/modules/products/internal/application"
	erpproductv1 "github.com/voxire/voxie/gen/go/erp/product/v1"
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
	// Keep the generated server for the contract routes, but use a raw-body
	// adapter for PUT updates so optional JSON fields retain wire presence.
	generatedMux := http.NewServeMux()
	productAPI := api.NewProductAPI(m.service)
	if err := erpproductv1.RegisterProductServiceServer(
		productAPI,
		erpproductv1.WithMux(generatedMux),
	); err != nil {
		panic(fmt.Errorf("register ERP product service: %w", err))
	}
	mux.Handle("PUT /erp/product/v1/products/{id}", http.HandlerFunc(productAPI.UpdateProductHTTP))
	mux.Handle("/erp/product/v1/", generatedMux)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
