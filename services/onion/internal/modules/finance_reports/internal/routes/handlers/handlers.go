package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/finance_reports/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ---- Balance Sheet ----

func BalanceSheetRoutes() []string { return []string{"/finance/balance-sheet"} }

type (
	BalanceSheetHandler  struct{ Service *application.Service }
	BalanceSheetRequest  struct{}
	BalanceSheetResponse struct {
		TotalAssets        float64 `json:"total_assets"`
		TotalLiabilities   float64 `json:"total_liabilities"`
		TotalEquity        float64 `json:"total_equity"`
		CashBalance        float64 `json:"cash_balance"`
		AccountsReceivable float64 `json:"accounts_receivable"`
		AccountsPayable    float64 `json:"accounts_payable"`
	}
)

func (h *BalanceSheetHandler) Serve(ctx context.Context, _ *BalanceSheetRequest) (BalanceSheetResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return BalanceSheetResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return BalanceSheetResponse{}, errorModels.PermissionDeniedServerError()
	}

	bs, err := h.Service.GetBalanceSheet(ctx, businessID)
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

type (
	CashFlowHandler struct{ Service *application.Service }
	CashFlowRequest struct {
		Period string `json:"period"`
	}
)

type CashFlowResponse struct {
	Period           string  `json:"period"`
	OperatingInflow  float64 `json:"operating_inflow"`
	OperatingOutflow float64 `json:"operating_outflow"`
	NetOperating     float64 `json:"net_operating"`
	NetCashFlow      float64 `json:"net_cash_flow"`
}

func (h *CashFlowHandler) Validate(_ context.Context, req *CashFlowRequest) *errorModels.ServerError {
	req.Period = strings.TrimSpace(req.Period)
	return nil
}

func (h *CashFlowHandler) Serve(ctx context.Context, req *CashFlowRequest) (CashFlowResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CashFlowResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CashFlowResponse{}, errorModels.PermissionDeniedServerError()
	}

	period := req.Period
	if period == "" {
		period = "month"
	}

	cf, err := h.Service.GetCashFlow(ctx, businessID, period)
	if err != nil {
		return CashFlowResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to generate cash flow report"}
	}

	return CashFlowResponse{
		Period: cf.Period, OperatingInflow: cf.OperatingInflow,
		OperatingOutflow: cf.OperatingOutflow, NetOperating: cf.NetOperating,
		NetCashFlow: cf.NetCashFlow,
	}, nil
}
