package billing

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	hubbillingv1 "github.com/voxire/voxie/gen/go/hub/billing/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/billing/internal/api"
	"github.com/voxire/onion/internal/modules/billing/internal/application"
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
	if err := hubbillingv1.RegisterBillingServiceServer(
		api.NewBillingAPI(m.service),
		hubbillingv1.WithMux(mux),
		hubbillingv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register hub billing service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
