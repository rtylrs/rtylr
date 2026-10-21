package products

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/erp/products/internal/api"
	"github.com/voxire/onion/internal/modules/erp/products/internal/application"
	"github.com/voxire/onion/internal/modules/erp/products/internal/catalog"
	moduleinfra "github.com/voxire/onion/internal/modules/erp/products/internal/infrastructure"
	erpproductv1 "github.com/voxire/voxie/gen/go/erp/product/v1"
)

type Module struct {
	service    *application.Service
	catalog    *catalog.Module
	repository *moduleinfra.Repository
}

func (m *Module) SetCatalogBusinessContext(owner func(context.Context, int64) (int64, error), onlineOrdering func(context.Context, int64) (bool, error)) {
	m.catalog.SetBusinessContext(owner, onlineOrdering)
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	repo := moduleinfra.NewRepository(db)
	return &Module{
		service:    application.NewService(repo),
		catalog:    catalog.NewModule(ctx, db),
		repository: repo,
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
	m.catalog.RegisterAPI(mux)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{"/pos/menu-item/v1/menu-items/public"}
}

// SalesProduct returns a current active product quote for one business.
// Missing or inactive products return sql.ErrNoRows; cost and supplier fields are excluded.
func (m *Module) SalesProduct(ctx context.Context, businessID, productID int64) (name, sku, price, tax string, err error) {
	return m.repository.SalesProduct(ctx, businessID, productID)
}
