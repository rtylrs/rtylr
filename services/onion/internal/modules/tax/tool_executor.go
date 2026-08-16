package tax

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/tax/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	getFinanceTaxConfigToolName = "finance.get_tax_config"
	getFinancePLReportToolName  = "finance.get_pl_report"
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, _ map[string]any) (map[string]any, bool, error) {
	switch name {
	case getFinanceTaxConfigToolName:
		output, err := m.executeGetTaxConfigTool(ctx)
		return output, true, err
	case getFinancePLReportToolName:
		output, err := m.executeGetPLReportTool(ctx)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeGetTaxConfigTool(ctx context.Context) (map[string]any, error) {
	businessID, err := taxToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	config, err := m.service.GetConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrTaxConfigNotFound) {
			return map[string]any{
				"configured": false,
				"config":     nil,
			}, nil
		}
		return nil, err
	}

	return map[string]any{
		"configured": true,
		"config":     taxConfigToolPayload(config),
	}, nil
}

func (m *Module) executeGetPLReportTool(ctx context.Context) (map[string]any, error) {
	businessID, err := taxToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	report, err := m.service.GetPLReport(ctx, businessID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"report":                     taxReportToolPayload(report),
		"total_revenue_minor_units":  report.TotalRevenueMinorUnits,
		"total_expenses_minor_units": report.TotalExpensesMinorUnits,
		"gross_profit_minor_units":   report.GrossProfitMinorUnits,
		"net_profit_minor_units":     report.NetProfitMinorUnits,
		"is_profitable":              report.NetProfitMinorUnits > 0,
	}, nil
}

func taxToolBusinessID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return 0, err
	}
	return businessID, nil
}

func taxConfigToolPayload(config domain.TaxConfig) map[string]any {
	return map[string]any{
		"id":                    config.ID,
		"tax_name":              config.TaxName,
		"tax_rate":              float64(config.TaxRateBasisPoints) / 100,
		"tax_rate_basis_points": config.TaxRateBasisPoints,
		"is_inclusive":          config.IsInclusive,
		"is_active":             config.IsActive,
		"created_at":            config.CreatedAt,
		"updated_at":            config.UpdatedAt,
	}
}

func taxReportToolPayload(report domain.PLReport) map[string]any {
	return map[string]any{
		"total_revenue_minor_units":  report.TotalRevenueMinorUnits,
		"total_expenses_minor_units": report.TotalExpensesMinorUnits,
		"gross_profit_minor_units":   report.GrossProfitMinorUnits,
		"net_profit_minor_units":     report.NetProfitMinorUnits,
	}
}
