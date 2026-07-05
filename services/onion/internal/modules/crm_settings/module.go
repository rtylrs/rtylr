package crmsettings

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	crmsettingsv1 "github.com/voxire/voxie/gen/go/crm/settings/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/crm_settings/internal/api"
	"github.com/voxire/onion/internal/modules/crm_settings/internal/application"
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
	if err := crmsettingsv1.RegisterSettingsServiceServer(
		api.NewSettingsAPI(m.service),
		crmsettingsv1.WithMux(mux),
		crmsettingsv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register crm settings service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
