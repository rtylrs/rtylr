package pos_customers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	poscustomerv1 "github.com/voxire/voxie/gen/go/pos/customer/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/pos_customers/internal/api"
	"github.com/voxire/onion/internal/modules/pos_customers/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := poscustomerv1.RegisterCustomerServiceServer(
		api.NewCustomerAPI(m.service),
		poscustomerv1.WithMux(mux),
		poscustomerv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register POS customer service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
