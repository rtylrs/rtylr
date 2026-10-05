package insights

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/insights/internal/api"
	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure"
	insightscustomerv1 "github.com/voxire/voxie/gen/go/insights/customer/v1"
	insightsdashboardv1 "github.com/voxire/voxie/gen/go/insights/dashboard/v1"
	insightsfinancialv1 "github.com/voxire/voxie/gen/go/insights/financial/v1"
	insightsoperationv1 "github.com/voxire/voxie/gen/go/insights/operation/v1"
	insightsproductv1 "github.com/voxire/voxie/gen/go/insights/product/v1"
	insightsreportv1 "github.com/voxire/voxie/gen/go/insights/report/v1"
	insightsrevenuev1 "github.com/voxire/voxie/gen/go/insights/revenue/v1"
	insightssettingv1 "github.com/voxire/voxie/gen/go/insights/setting/v1"
)

// Module is the cross-domain aggregation layer that powers the /insights/*
// namespace. Most endpoints read across Finance, CRM, POS, ERP and HR schemas;
// settings endpoints also persist insights preferences such as scheduled reports.
type Module struct {
	service *application.Service
	flow    api.FlowReader
}

// FlowReader is re-exported so a caller can name the dependency this module
// needs without importing an internal package.
type FlowReader = api.FlowReader

type FlowDelivery = api.FlowDelivery
type FlowWorkloadMember = api.FlowWorkloadMember
type FlowRoadmapProject = api.FlowRoadmapProject

type FinanceDashboardReader = application.FinanceDashboardReader
type POSDashboardReader = application.POSDashboardReader
type CRMDashboardReader = application.CRMDashboardReader
type DashboardReaders = application.DashboardReaders

// NewModule takes Flow's insights module because insights.dashboard.v1 declares
// three projections of Flow's delivery data. Passing the module rather than a
// database handle keeps the flow_* schema owned by Flow: this module asks Flow
// for numbers and never queries its tables.
func NewModule(ctx context.Context, db *sql.DB, flow api.FlowReader, dashboard DashboardReaders) *Module {
	return &Module{service: application.NewService(infrastructure.NewRepository(db), dashboard), flow: flow}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	register := func(name string, fn func() error) {
		if err := fn(); err != nil {
			panic(fmt.Errorf("register insights %s service: %w", name, err))
		}
	}

	register("revenue", func() error {
		return insightsrevenuev1.RegisterRevenueServiceServer(
			api.NewRevenueAPI(m.service),
			insightsrevenuev1.WithMux(mux),
		)
	})
	register("customer", func() error {
		return insightscustomerv1.RegisterCustomerServiceServer(
			api.NewCustomerAPI(m.service),
			insightscustomerv1.WithMux(mux),
		)
	})
	register("product", func() error {
		return insightsproductv1.RegisterProductServiceServer(
			api.NewProductAPI(m.service),
			insightsproductv1.WithMux(mux),
		)
	})
	register("operation", func() error {
		return insightsoperationv1.RegisterOperationServiceServer(
			api.NewOperationAPI(m.service),
			insightsoperationv1.WithMux(mux),
		)
	})
	register("financial", func() error {
		return insightsfinancialv1.RegisterFinancialServiceServer(
			api.NewFinancialAPI(m.service),
			insightsfinancialv1.WithMux(mux),
		)
	})
	register("dashboard", func() error {
		return insightsdashboardv1.RegisterDashboardServiceServer(
			api.NewDashboardAPI(m.service, m.flow),
			insightsdashboardv1.WithMux(mux),
		)
	})
	register("setting", func() error {
		return insightssettingv1.RegisterSettingServiceServer(
			api.NewSettingAPI(m.service),
			insightssettingv1.WithMux(mux),
		)
	})
	register("report", func() error {
		return insightsreportv1.RegisterReportServiceServer(
			api.NewReportAPI(m.service),
			insightsreportv1.WithMux(mux),
			insightsreportv1.WithMiddleware(api.ScheduledReportPresenceMiddleware),
		)
	})
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	_ = ctx
	_ = router
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
