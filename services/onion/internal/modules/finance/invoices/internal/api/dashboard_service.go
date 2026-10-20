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

	monthly := make([]*financedashboardv1.DashboardMonth, 0, len(dashboard.Monthly))
	for _, m := range dashboard.Monthly {
		monthly = append(monthly, &financedashboardv1.DashboardMonth{Month: m.Month, Revenue: m.Revenue, Expenses: m.Expenses})
	}
	activity := make([]*financedashboardv1.DashboardActivity, 0, len(dashboard.RecentActivity))
	for _, a := range dashboard.RecentActivity {
		activity = append(activity, &financedashboardv1.DashboardActivity{
			Kind:        a.Kind,
			Id:          a.ID,
			Description: a.Description,
			Amount:      a.Amount,
			Date:        a.Date,
			Status:      a.Status,
		})
	}
	overdue := make([]*financedashboardv1.DashboardOverdueInvoice, 0, len(dashboard.OverdueInvoices))
	for _, inv := range dashboard.OverdueInvoices {
		overdue = append(overdue, &financedashboardv1.DashboardOverdueInvoice{
			Id:            inv.ID,
			InvoiceNumber: inv.InvoiceNumber,
			ClientName:    inv.ClientName,
			AmountDue:     inv.AmountDue,
			DaysOverdue:   inv.DaysOverdue,
		})
	}

	return &financedashboardv1.GetFinanceDashboardResponse{
		TotalReceivable:     dashboard.TotalReceivable,
		OverdueAmount:       dashboard.OverdueAmount,
		TotalExpensesMtd:    dashboard.TotalExpensesMTD,
		TotalRevenueMtd:     dashboard.TotalRevenueMTD,
		NetProfitMtd:        dashboard.NetProfitMTD,
		PendingExpenses:     dashboard.PendingExpenses,
		PendingExpenseCount: dashboard.PendingExpenseCount,
		Monthly:             monthly,
		RecentActivity:      activity,
		OverdueInvoices:     overdue,
		OverdueInvoiceCount: dashboard.OverdueInvoiceCount,
	}, nil
}
