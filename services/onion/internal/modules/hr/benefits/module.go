package benefits

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/benefits/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hr/benefits/internal/api"
	"github.com/voxire/onion/internal/modules/hr/benefits/internal/application"
	"github.com/voxire/onion/internal/modules/hr/benefits/internal/domain"
	hrbenefitv1 "github.com/voxire/voxie/gen/go/hr/benefit/v1"
)

type Module struct {
	service *application.Service
}

// NewModule takes an EmployeeDirectory (hr_core satisfies it) so self-service
// enrollment can resolve the caller's employee record.
func NewModule(ctx context.Context, db *sql.DB, employees domain.EmployeeDirectory) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db), employees)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrbenefitv1.RegisterBenefitServiceServer(
		api.NewBenefitAPI(m.service),
		hrbenefitv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hr benefit service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
