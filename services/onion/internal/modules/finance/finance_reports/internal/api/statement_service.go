package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/finance/finance_reports/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	financestatementv1 "github.com/voxire/voxie/gen/go/finance/statement/v1"
)

type StatementAPI struct {
	Service *application.Service
}

func NewStatementAPI(service *application.Service) *StatementAPI {
	return &StatementAPI{Service: service}
}

func (a *StatementAPI) GetBalanceSheet(ctx context.Context, req *financestatementv1.GetBalanceSheetRequest) (*financestatementv1.GetBalanceSheetResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(device.User, authorization.AppFinance, "finance:reports:view") {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	asOf := ""
	if req != nil {
		asOf = req.AsOf
	}
	bs, err := a.Service.GetBalanceSheetAsOf(ctx, businessID, asOf)
	if errors.Is(err, application.ErrInvalidAsOf) {
		return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to generate balance sheet")
	}

	lines := make([]*financestatementv1.BalanceSheetLine, 0, len(bs.Lines))
	for _, line := range bs.Lines {
		lines = append(lines, &financestatementv1.BalanceSheetLine{
			AccountId: line.AccountID,
			Code:      line.Code,
			Name:      line.Name,
			Type:      line.Type,
			Amount:    line.Amount,
		})
	}
	return &financestatementv1.GetBalanceSheetResponse{
		TotalAssets:        bs.TotalAssets,
		TotalLiabilities:   bs.TotalLiabilities,
		TotalEquity:        bs.TotalEquity,
		CashBalance:        bs.CashBalance,
		AccountsReceivable: bs.AccountsReceivable,
		AccountsPayable:    bs.AccountsPayable,
		AsOf:               bs.AsOf,
		Lines:              lines,
		NetIncome:          bs.NetIncome,
	}, nil
}

func (a *StatementAPI) GetCashFlow(ctx context.Context, req *financestatementv1.GetCashFlowRequest) (*financestatementv1.GetCashFlowResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(device.User, authorization.AppFinance, "finance:reports:view") {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	period := strings.TrimSpace(req.GetPeriod())
	if period == "" {
		period = "month"
	}

	cf, err := a.Service.GetCashFlow(ctx, businessID, period)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to generate cash flow report")
	}

	return &financestatementv1.GetCashFlowResponse{
		Period:           cf.Period,
		OperatingInflow:  cf.OperatingInflow,
		OperatingOutflow: cf.OperatingOutflow,
		NetOperating:     cf.NetOperating,
		NetCashFlow:      cf.NetCashFlow,
	}, nil
}
