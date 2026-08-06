package pos_orders

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/routes"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/routes/handlers"
	posdashboardv1 "github.com/voxire/voxie/gen/go/pos/dashboard/v1"
	posordersdiscountsv1 "github.com/voxire/voxie/gen/go/pos/orders/discounts/v1"
	posordersholdsv1 "github.com/voxire/voxie/gen/go/pos/orders/holds/v1"
	posordersitemsv1 "github.com/voxire/voxie/gen/go/pos/orders/items/v1"
	posorderspaymentsv1 "github.com/voxire/voxie/gen/go/pos/orders/payments/v1"
	posordersrefundsv1 "github.com/voxire/voxie/gen/go/pos/orders/refunds/v1"
	posorderssplitv1 "github.com/voxire/voxie/gen/go/pos/orders/splits/v1"
	posordersv1 "github.com/voxire/voxie/gen/go/pos/orders/v1"
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
	register := func(err error, label string) {
		if err != nil {
			panic(fmt.Errorf("register %s: %w", label, err))
		}
	}

	register(posdashboardv1.RegisterDashboardServiceServer(
		handlers.NewDashboardAPI(m.service),
		posdashboardv1.WithMux(mux),
	), "POS dashboard service")

	register(posordersv1.RegisterOrderServiceServer(
		handlers.NewOrderAPI(m.service),
		posordersv1.WithMux(mux),
	), "POS order service")

	register(posorderspaymentsv1.RegisterOrderPaymentServiceServer(
		handlers.NewPaymentAPI(m.service),
		posorderspaymentsv1.WithMux(mux),
	), "POS order payment service")

	register(posordersrefundsv1.RegisterOrderRefundServiceServer(
		handlers.NewRefundAPI(m.service),
		posordersrefundsv1.WithMux(mux),
	), "POS order refund service")

	register(posorderssplitv1.RegisterOrderSplitServiceServer(
		handlers.NewSplitAPI(m.service),
		posorderssplitv1.WithMux(mux),
	), "POS order split service")

	register(posordersdiscountsv1.RegisterOrderDiscountServiceServer(
		handlers.NewDiscountAPI(m.service),
		posordersdiscountsv1.WithMux(mux),
	), "POS order discount service")

	register(posordersholdsv1.RegisterOrderHoldServiceServer(
		handlers.NewHoldAPI(m.service),
		posordersholdsv1.WithMux(mux),
	), "POS order hold service")

	register(posordersitemsv1.RegisterOrderItemServiceServer(
		handlers.NewItemAPI(m.service),
		posordersitemsv1.WithMux(mux),
	), "POS order item service")
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	routes.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{
		"/pos/payments/webhook",
		"/pos/orders/v1/orders/public",
		"/pos/orders/v1/orders/{order_id}/public-status",
		"/pos/ws",
	}
}
