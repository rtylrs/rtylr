package api

import (
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
)

type PLReportHandler struct {
	FinanceService *application.Service
}

type PLReportRequest struct {
	Period string `json:"period"`
	Date   string `json:"date"`
}

type PLReportResponse struct {
	Period            string  `json:"period"`
	Revenue           float64 `json:"revenue"`
	COGS              float64 `json:"cogs"`
	GrossProfit       float64 `json:"gross_profit"`
	OperatingExpenses float64 `json:"operating_expenses"`
	NetProfit         float64 `json:"net_profit"`
	MarginPct         float64 `json:"margin_pct"`
}
