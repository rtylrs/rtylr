package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func PLReportRoutes() []string {
	return []string{"/finance/pl"}
}

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
	GrossProfit       float64 `json:"grossProfit"`
	OperatingExpenses float64 `json:"operatingExpenses"`
	NetProfit         float64 `json:"netProfit"`
	MarginPct         float64 `json:"marginPct"`
}

func (h *PLReportHandler) Serve(ctx context.Context, req *PLReportRequest) (PLReportResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	period := req.Period
	if period == "" {
		period = "month"
	}

	report, err := h.FinanceService.GetPLReport(ctx, device.User.UserID, period, req.Date)
	if err != nil {
		return PLReportResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to generate P&L report",
		}
	}

	return PLReportResponse{
		Period:            report.Period,
		Revenue:           report.Revenue,
		COGS:              report.COGS,
		GrossProfit:       report.GrossProfit,
		OperatingExpenses: report.OperatingExpenses,
		NetProfit:         report.NetProfit,
		MarginPct:         report.MarginPct,
	}, nil
}
