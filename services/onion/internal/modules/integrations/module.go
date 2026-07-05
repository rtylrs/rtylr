package integrations

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	hubintegrationv1 "github.com/voxire/voxie/gen/go/hub/integration/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/integrations/internal/api"
	"github.com/voxire/onion/internal/modules/integrations/internal/application"
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
	if err := hubintegrationv1.RegisterIntegrationServiceServer(
		api.NewIntegrationAPI(m.service),
		hubintegrationv1.WithMux(mux),
		hubintegrationv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register hub integration service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
