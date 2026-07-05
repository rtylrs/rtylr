package supportsettings

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	supportsettingsv1 "github.com/voxire/voxie/gen/go/support/settings/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/support_settings/internal/api"
	"github.com/voxire/onion/internal/modules/support_settings/internal/application"
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
	if err := supportsettingsv1.RegisterSettingsServiceServer(
		api.NewSettingsAPI(m.service),
		supportsettingsv1.WithMux(mux),
		supportsettingsv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register support settings service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
