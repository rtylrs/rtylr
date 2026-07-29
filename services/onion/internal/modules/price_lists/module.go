package price_lists

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/price_lists/internal/api"
	"github.com/voxire/onion/internal/modules/price_lists/internal/application"
	erppricelistv1 "github.com/voxire/voxie/gen/go/erp/price_list/v1"
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
	if err := erppricelistv1.RegisterPriceListServiceServer(
		api.NewPriceListAPI(m.service),
		erppricelistv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register ERP price list service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
