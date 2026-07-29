package brands

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/brands/internal/api"
	"github.com/voxire/onion/internal/modules/brands/internal/application"
	erpbrandv1 "github.com/voxire/voxie/gen/go/erp/brand/v1"
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
	if err := erpbrandv1.RegisterBrandServiceServer(
		api.NewBrandAPI(m.service),
		erpbrandv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register ERP brand service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
