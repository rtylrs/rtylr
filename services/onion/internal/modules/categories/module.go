package categories

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/categories/internal/api"
	"github.com/voxire/onion/internal/modules/categories/internal/application"
	erpcategoryv1 "github.com/voxire/voxie/gen/go/erp/category/v1"
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
	if err := erpcategoryv1.RegisterCategoryServiceServer(
		api.NewCategoryAPI(m.service),
		erpcategoryv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register ERP category service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
