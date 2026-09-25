package financereports

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	getFinanceBalanceSheetToolName   = "finance.get_balance_sheet"
	getFinanceCashFlowReportToolName = "finance.get_cash_flow_report"
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case getFinanceBalanceSheetToolName:
		businessID, err := financeReportsToolBusinessID(ctx)
		if err != nil {
			return nil, true, err
		}
		balanceSheet, err := m.service.GetBalanceSheet(ctx, businessID)
		if err != nil {
			return nil, true, err
		}
		return map[string]any{"balance_sheet": balanceSheet}, true, nil
	case getFinanceCashFlowReportToolName:
		businessID, err := financeReportsToolBusinessID(ctx)
		if err != nil {
			return nil, true, err
		}
		period := strings.TrimSpace(financeReportsToolString(firstFinanceReportsToolValue(input, "period", "range")))
		if period == "" {
			period = "month"
		}
		cashFlow, err := m.service.GetCashFlow(ctx, businessID, period)
		if err != nil {
			return nil, true, err
		}
		return map[string]any{"cash_flow": cashFlow}, true, nil
	default:
		return nil, false, nil
	}
}

func financeReportsToolCanRead(authUser *authPkg.AuthUser) bool {
	if authUser == nil || !authPkg.HasAppAccess(authUser, authorization.AppFinance) {
		return false
	}
	if authPkg.LegacyAuthorizationMode() {
		// The legacy compatibility helper treats any finance permission as a
		// match. Keep the Caret boundary role-scoped so an ordinary legacy app
		// member cannot use the tool as a report reader.
		return authPkg.IsBusinessAdmin(authUser) ||
			authPkg.HasAnyRole(authUser, "finance.viewer", "viewer", "finance.manager", "manager", "finance.admin", "admin", "owner", "console_admin")
	}
	return authPkg.HasAppPermission(authUser, authorization.AppFinance, "finance:reports:view")
}

func financeReportsToolBusinessID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	// Caret executes outside the generated finance statement HTTP route, so it
	// must repeat the same feature permission check before resolving tenant
	// context or touching report data. App membership alone is not report-view
	// authorization.
	if !financeReportsToolCanRead(authUser) {
		return 0, errors.New("permission denied")
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return 0, errors.New("authentication is required")
	}
	return businessID, nil
}

func firstFinanceReportsToolValue(input map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := input[key]
		if !ok || value == nil {
			continue
		}
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		return value
	}
	return nil
}

func financeReportsToolString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}
