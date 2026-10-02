package customers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/customers/internal/api"
	"github.com/voxire/onion/internal/modules/customers/internal/application"
	crmcustomerv1 "github.com/voxire/voxie/gen/go/crm/customer/v1"
	crmdashboardv1 "github.com/voxire/voxie/gen/go/crm/dashboard/v1"
	crmloyaltyv1 "github.com/voxire/voxie/gen/go/crm/loyalty/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := crmcustomerv1.RegisterCustomerServiceServer(api.NewCustomerAPI(m.service), crmcustomerv1.WithMux(mux)); err != nil {
		panic(fmt.Errorf("register crm customer service: %w", err))
	}
	if err := crmdashboardv1.RegisterDashboardServiceServer(api.NewDashboardAPI(m.service), crmdashboardv1.WithMux(mux)); err != nil {
		panic(fmt.Errorf("register crm dashboard service: %w", err))
	}

	if err := crmloyaltyv1.RegisterLoyaltyServiceServer(
		api.NewLoyaltyAPI(m.service),
		crmloyaltyv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register crm loyalty service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
