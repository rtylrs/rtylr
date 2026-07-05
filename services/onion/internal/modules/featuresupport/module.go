package featuresupport

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	hubfeaturev1 "github.com/voxire/voxie/gen/go/hub/feature/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/featuresupport/internal/api"
	"github.com/voxire/onion/internal/modules/featuresupport/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubfeaturev1.RegisterFeatureServiceServer(
		api.NewFeatureAPI(m.service),
		hubfeaturev1.WithMux(mux),
		hubfeaturev1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register hub feature service: %w", err))
	}
}

// GetAnonymousRoutes exposes /features to the landing page without auth. Note
// the per-user reads use /features/me (authenticated, not listed here), and the
// /features/admin/* surface stays behind auth + an owner/admin role gate.
func (m *Module) GetAnonymousRoutes() []string {
	return []string{"/features"}
}
