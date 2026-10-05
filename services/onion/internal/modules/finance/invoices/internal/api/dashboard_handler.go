package api

import (
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
)

type DashboardHandler struct {
	FinanceService *application.Service
}

type DashboardRequest struct{}

type DashboardResponse struct {
	TotalReceivable  float64 `json:"total_receivable"`
	OverdueAmount    float64 `json:"overdue_amount"`
	TotalExpensesMTD float64 `json:"total_expenses_mtd"`
	TotalRevenueMTD  float64 `json:"total_revenue_mtd"`
	NetProfitMTD     float64 `json:"net_profit_mtd"`
}
