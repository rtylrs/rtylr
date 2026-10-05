package featuresupport

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/organization/featuresupport/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/organization/featuresupport/internal/api"
	"github.com/voxire/onion/internal/modules/organization/featuresupport/internal/application"
	hubfeaturev1 "github.com/voxire/voxie/gen/go/hub/feature/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, moduleinfra.NewRepository(db))}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubfeaturev1.RegisterFeatureServiceServer(
		api.NewFeatureAPI(m.service),
		hubfeaturev1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub feature service: %w", err))
	}
}

// GetAnonymousRoutes exposes anonymous feature resolution without auth. The
// per-user and admin surfaces remain protected.
func (m *Module) GetAnonymousRoutes() []string {
	return []string{"/hub/feature/v1/features"}
}
