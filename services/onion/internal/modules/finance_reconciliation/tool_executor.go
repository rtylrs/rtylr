package financereconciliation

import (
	"context"
	"errors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const getFinanceReconciliationOverviewToolName = "finance.get_reconciliation_overview"

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, _ map[string]any) (map[string]any, bool, error) {
	if name != getFinanceReconciliationOverviewToolName {
		return nil, false, nil
	}

	businessID, err := financeReconciliationToolBusinessID(ctx)
	if err != nil {
		return nil, true, err
	}

	overview, err := m.service.Overview(ctx, businessID)
	if err != nil {
		return nil, true, err
	}

	highConfidence := 0
	for _, suggestion := range overview.Suggestions {
		if suggestion.IsHighConfidence() {
			highConfidence++
		}
	}

	return map[string]any{
		"overview":                    overview,
		"unmatched_bank_txn_count":    len(overview.UnmatchedBankTxns),
		"unmatched_invoice_count":     len(overview.UnmatchedInvoices),
		"suggestion_count":            len(overview.Suggestions),
		"high_confidence_suggestions": highConfidence,
		"has_unmatched_items":         overview.HasUnmatchedItems(),
		"has_suggestions":             overview.HasSuggestions(),
	}, true, nil
}

func financeReconciliationToolBusinessID(ctx context.Context) (int64, error) {
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
