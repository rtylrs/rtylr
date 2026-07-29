package supportcustomers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/support_customers/internal/api"
	"github.com/voxire/onion/internal/modules/support_customers/internal/application"
	supportcustomerv1 "github.com/voxire/voxie/gen/go/support/customer/v1"
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
	if err := supportcustomerv1.RegisterCustomerServiceServer(
		api.NewCustomerAPI(m.service),
		supportcustomerv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register support customer service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
