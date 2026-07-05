package pos_combos

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	poscombov1 "github.com/voxire/voxie/gen/go/pos/combo/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/pos_combos/internal/api"
	"github.com/voxire/onion/internal/modules/pos_combos/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := poscombov1.RegisterComboServiceServer(
		api.NewComboAPI(m.service),
		poscombov1.WithMux(mux),
		poscombov1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register POS combo service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
