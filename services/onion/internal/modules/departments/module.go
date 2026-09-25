package departments

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/departments/internal/api"
	"github.com/voxire/onion/internal/modules/departments/internal/application"
	hrdepartmentv1 "github.com/voxire/voxie/gen/go/hr/department/v1"
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
