package bankaccounts

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const (
	listFinanceBankAccountsToolName = "finance.list_bank_accounts"
	getFinanceBankSummaryToolName   = "finance.get_bank_summary"
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, _ map[string]any) (map[string]any, bool, error) {
	switch name {
	case listFinanceBankAccountsToolName:
		output, err := m.executeListBankAccountsTool(ctx)
		return output, true, err
	case getFinanceBankSummaryToolName:
		output, err := m.executeGetBankSummaryTool(ctx)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListBankAccountsTool(ctx context.Context) (map[string]any, error) {
	ownerID, err := bankAccountsToolOwnerID(ctx)
	if err != nil {
		return nil, err
	}

	accounts, err := m.service.List(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	totalBalance, err := m.service.TotalBalance(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	connectedCount, err := m.service.ConnectedCount(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(accounts))
	for _, account := range accounts {
		payloads = append(payloads, bankAccountToolPayload(account))
	}

	return map[string]any{
		"accounts":        payloads,
		"returned":        len(payloads),
		"total_accounts":  len(accounts),
		"connected_count": connectedCount,
		"total_balance":   totalBalance,
	}, nil
}

func (m *Module) executeGetBankSummaryTool(ctx context.Context) (map[string]any, error) {
	ownerID, err := bankAccountsToolOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	totalBalance, err := m.service.TotalBalance(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	connectedCount, err := m.service.ConnectedCount(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	accounts, err := m.service.List(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"total_balance":   totalBalance,
		"connected_count": connectedCount,
		"account_count":   len(accounts),
		"has_connections": connectedCount > 0,
	}, nil
}

func bankAccountsToolOwnerID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	return authUser.UserID, nil
}

func bankAccountToolPayload(account domain.BankAccount) map[string]any {
	return map[string]any{
		"id":             account.ID,
		"name":           account.Name,
		"institution":    account.Institution,
		"account_type":   account.AccountType,
		"balance":        account.Balance,
		"currency":       account.Currency,
		"is_connected":   account.IsConnected,
		"last_synced_at": account.LastSyncedAt,
		"created_at":     account.CreatedAt,
	}
}
