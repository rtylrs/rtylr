package pos_manager

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	posmanagerv1 "github.com/voxire/voxie/gen/go/pos/manager/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/api"
	"github.com/voxire/onion/internal/modules/pos_manager/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := posmanagerv1.RegisterManagerServiceServer(
		api.NewManagerAPI(m.service),
		posmanagerv1.WithMux(mux),
		posmanagerv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register POS manager service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
