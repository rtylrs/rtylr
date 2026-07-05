package bills

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	financebillv1 "github.com/voxire/voxie/gen/go/finance/bill/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/bills/internal/api"
	"github.com/voxire/onion/internal/modules/bills/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financebillv1.RegisterBillServiceServer(
		api.NewBillAPI(m.service),
		financebillv1.WithMux(mux),
		financebillv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register finance bill service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
