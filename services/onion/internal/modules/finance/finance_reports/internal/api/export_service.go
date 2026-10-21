package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type ExportStatementRequest struct {
	Period string `json:"period"`
}

type ExportStatementResponse struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	CSV         string `json:"csv"`
}

// ExportCashFlow is the finance report export boundary. It deliberately uses
// the same scoped application query as the interactive report while requiring
// the separate export permission.
func (a *StatementAPI) ExportCashFlow(ctx context.Context, req *ExportStatementRequest) (ExportStatementResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppFinance, "finance:reports:export") {
		return ExportStatementResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return ExportStatementResponse{}, onekiterrors.PermissionDenied()
	}
	period := strings.TrimSpace(req.GetPeriod())
	if period == "" {
		period = "month"
	}
	flow, err := a.Service.GetCashFlow(ctx, businessID, period)
	if err != nil {
		return ExportStatementResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to generate cash flow export")
	}
	filename := fmt.Sprintf("cash_flow_%s.csv", period)
	csv := strings.Join([]string{
		"period,operating_inflow,operating_outflow,net_operating,net_cash_flow",
		strings.Join([]string{flow.Period, strconv.FormatFloat(flow.OperatingInflow, 'f', 2, 64), strconv.FormatFloat(flow.OperatingOutflow, 'f', 2, 64), strconv.FormatFloat(flow.NetOperating, 'f', 2, 64), strconv.FormatFloat(flow.NetCashFlow, 'f', 2, 64)}, ","),
	}, "\n")
	return ExportStatementResponse{Filename: filename, ContentType: "text/csv", CSV: csv + "\n"}, nil
}

func (r *ExportStatementRequest) GetPeriod() string {
	if r == nil {
		return ""
	}
	return r.Period
}
