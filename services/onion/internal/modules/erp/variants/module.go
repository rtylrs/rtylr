package variants

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/erp/variants/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/erp/variants/internal/api"
	"github.com/voxire/onion/internal/modules/erp/variants/internal/application"
	erpvariantv1 "github.com/voxire/voxie/gen/go/erp/variant/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, moduleinfra.NewRepository(db)),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := erpvariantv1.RegisterVariantServiceServer(
		api.NewVariantAPI(m.service),
		erpvariantv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register ERP variant service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
