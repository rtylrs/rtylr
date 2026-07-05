package pos_sessions

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_sessions/internal/api"
	"github.com/voxire/onion/internal/modules/pos_sessions/internal/application"
	possessionv1 "github.com/voxire/voxie/gen/go/pos/session/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := possessionv1.RegisterSessionServiceServer(
		api.NewSessionAPI(m.service),
		possessionv1.WithMux(mux),
		possessionv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register POS session service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
