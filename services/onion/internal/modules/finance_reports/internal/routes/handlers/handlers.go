package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/finance_reports/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ---- Balance Sheet ----

func BalanceSheetRoutes() []string { return []string{"/finance/balance-sheet"} }

type BalanceSheetHandler struct{ Service *application.Service }
type BalanceSheetRequest struct{}
type BalanceSheetResponse struct {
	TotalAssets         float64 `json:"totalAssets"`
	TotalLiabilities    float64 `json:"totalLiabilities"`
	TotalEquity         float64 `json:"totalEquity"`
	CashBalance         float64 `json:"cashBalance"`
	AccountsReceivable  float64 `json:"accountsReceivable"`
	AccountsPayable     float64 `json:"accountsPayable"`
}

func (h *BalanceSheetHandler) Serve(ctx context.Context, _ *BalanceSheetRequest) (BalanceSheetResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return BalanceSheetResponse{}, errorModels.PermissionDeniedServerError()
	}

	bs, err := h.Service.GetBalanceSheet(ctx, device.User.UserID)
	if err != nil {
		return BalanceSheetResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to generate balance sheet"}
	}

	return BalanceSheetResponse{
		TotalAssets: bs.TotalAssets, TotalLiabilities: bs.TotalLiabilities,
		TotalEquity: bs.TotalEquity, CashBalance: bs.CashBalance,
		AccountsReceivable: bs.AccountsReceivable, AccountsPayable: bs.AccountsPayable,
	}, nil
}

// ---- Cash Flow ----

func CashFlowRoutes() []string { return []string{"/finance/cash-flow"} }

type CashFlowHandler struct{ Service *application.Service }
type CashFlowRequest struct {
	Period string `json:"period"`
}
type CashFlowResponse struct {
	Period           string  `json:"period"`
	OperatingInflow  float64 `json:"operatingInflow"`
	OperatingOutflow float64 `json:"operatingOutflow"`
	NetOperating     float64 `json:"netOperating"`
	NetCashFlow      float64 `json:"netCashFlow"`
}

func (h *CashFlowHandler) Serve(ctx context.Context, req *CashFlowRequest) (CashFlowResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CashFlowResponse{}, errorModels.PermissionDeniedServerError()
	}

	period := req.Period
	if period == "" {
		period = "month"
	}

	cf, err := h.Service.GetCashFlow(ctx, device.User.UserID, period)
	if err != nil {
		return CashFlowResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to generate cash flow report"}
	}

	return CashFlowResponse{
		Period: cf.Period, OperatingInflow: cf.OperatingInflow,
		OperatingOutflow: cf.OperatingOutflow, NetOperating: cf.NetOperating,
		NetCashFlow: cf.NetCashFlow,
	}, nil
}
