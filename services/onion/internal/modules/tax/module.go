package tax

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	financetaxv1 "github.com/voxire/voxie/gen/go/finance/tax/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/tax/internal/api"
	"github.com/voxire/onion/internal/modules/tax/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financetaxv1.RegisterTaxServiceServer(
		api.NewTaxAPI(m.service),
		financetaxv1.WithMux(mux),
		financetaxv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register finance tax service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
