package departments

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/departments/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hr/departments/internal/api"
	"github.com/voxire/onion/internal/modules/hr/departments/internal/application"
	hrdepartmentv1 "github.com/voxire/voxie/gen/go/hr/department/v1"
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
	if err := hrdepartmentv1.RegisterDepartmentServiceServer(
		api.NewDepartmentAPI(m.service),
		hrdepartmentv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hr department service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
