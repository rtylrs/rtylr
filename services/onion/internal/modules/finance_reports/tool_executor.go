package financereports

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const (
	getFinanceBalanceSheetToolName   = "finance.get_balance_sheet"
	getFinanceCashFlowReportToolName = "finance.get_cash_flow_report"
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case getFinanceBalanceSheetToolName:
		ownerID, err := financeReportsToolOwnerID(ctx)
		if err != nil {
			return nil, true, err
		}
		balanceSheet, err := m.service.GetBalanceSheet(ctx, ownerID)
		if err != nil {
			return nil, true, err
		}
		return map[string]any{"balance_sheet": balanceSheet}, true, nil
	case getFinanceCashFlowReportToolName:
		ownerID, err := financeReportsToolOwnerID(ctx)
		if err != nil {
			return nil, true, err
		}
		period := strings.TrimSpace(financeReportsToolString(firstFinanceReportsToolValue(input, "period", "range")))
		if period == "" {
			period = "month"
		}
		cashFlow, err := m.service.GetCashFlow(ctx, ownerID, period)
		if err != nil {
			return nil, true, err
		}
		return map[string]any{"cash_flow": cashFlow}, true, nil
	default:
		return nil, false, nil
	}
}

func financeReportsToolOwnerID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	return authUser.UserID, nil
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
