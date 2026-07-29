package pos_customers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_customers/internal/api"
	"github.com/voxire/onion/internal/modules/pos_customers/internal/application"
	poscustomerv1 "github.com/voxire/voxie/gen/go/pos/customer/v1"
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
	); err != nil {
		panic(fmt.Errorf("register POS customer service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
