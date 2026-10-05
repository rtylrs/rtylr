package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	financedashboardv1 "github.com/voxire/voxie/gen/go/finance/dashboard/v1"
)

type DashboardAPI struct {
	Service *application.Service
}

func NewDashboardAPI(service *application.Service) *DashboardAPI {
	return &DashboardAPI{Service: service}
}

func (a *DashboardAPI) GetFinanceDashboard(ctx context.Context, _ *financedashboardv1.GetFinanceDashboardRequest) (*financedashboardv1.GetFinanceDashboardResponse, error) {
	if err := requireFinancePermission(ctx, "finance:reports:view"); err != nil {
		return nil, err
	}
	businessID, errModel := invoiceBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	dashboard, err := a.Service.GetDashboard(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get finance dashboard")
	}

	return &financedashboardv1.GetFinanceDashboardResponse{
		TotalReceivable:  dashboard.TotalReceivable,
		OverdueAmount:    dashboard.OverdueAmount,
		TotalExpensesMtd: dashboard.TotalExpensesMTD,
		TotalRevenueMtd:  dashboard.TotalRevenueMTD,
		NetProfitMtd:     dashboard.NetProfitMTD,
	}, nil
}
