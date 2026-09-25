package bankaccounts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listFinanceBankAccountsToolName = "finance.list_bank_accounts"
	getFinanceBankAccountToolName   = "finance.get_bank_account"
	getFinanceBankSummaryToolName   = "finance.get_bank_summary"
)

type getBankAccountToolInput struct {
	BankAccountID int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listFinanceBankAccountsToolName:
		output, err := m.executeListBankAccountsTool(ctx)
		return output, true, err
	case getFinanceBankAccountToolName:
		output, err := m.executeGetBankAccountTool(ctx, input)
		return output, true, err
	case getFinanceBankSummaryToolName:
		output, err := m.executeGetBankSummaryTool(ctx)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeGetBankAccountTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := bankAccountsToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	parsed, err := parseGetBankAccountToolInput(input)
	if err != nil {
		return nil, err
	}

	account, err := m.service.GetByID(ctx, businessID, parsed.BankAccountID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"account": bankAccountToolPayload(account),
	}, nil
}

func (m *Module) executeListBankAccountsTool(ctx context.Context) (map[string]any, error) {
	businessID, err := bankAccountsToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	accounts, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}
	totalBalance, err := m.service.TotalBalance(ctx, businessID)
	if err != nil {
		return nil, err
	}
	connectedCount, err := m.service.ConnectedCount(ctx, businessID)
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
	businessID, err := bankAccountsToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	totalBalance, err := m.service.TotalBalance(ctx, businessID)
	if err != nil {
		return nil, err
	}
	connectedCount, err := m.service.ConnectedCount(ctx, businessID)
	if err != nil {
		return nil, err
	}
	accounts, err := m.service.List(ctx, businessID)
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

func bankAccountsToolBusinessID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	return tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
}

func parseGetBankAccountToolInput(input map[string]any) (getBankAccountToolInput, error) {
	bankAccountID, err := bankAccountIDToolValue(firstBankAccountToolValue(input, "bank_account_id", "account_id", "account", "id"))
	if err != nil {
		return getBankAccountToolInput{}, fmt.Errorf("invalid bank_account_id: %w", err)
	}
	if bankAccountID <= 0 {
		return getBankAccountToolInput{}, errors.New("bank_account_id is required")
	}
	return getBankAccountToolInput{BankAccountID: bankAccountID}, nil
}

func firstBankAccountToolValue(input map[string]any, keys ...string) any {
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

func bankAccountIDToolValue(value any) (int64, error) {
	switch v := value.(type) {
	case string:
		raw := strings.TrimSpace(v)
		raw = strings.TrimPrefix(raw, "#")
		raw = strings.ToLower(raw)
		for _, prefix := range []string{"bank-account-", "bank_account_", "account-", "account_", "acct-", "acct_"} {
			raw = strings.TrimPrefix(raw, prefix)
		}
		return bankAccountToolInt64(raw)
	default:
		return bankAccountToolInt64(value)
	}
}

func bankAccountToolInt64(value any) (int64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		if math.Trunc(v) != v {
			return 0, errors.New("value must be an integer")
		}
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return 0, nil
		}
		return strconv.ParseInt(s, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported number type %T", value)
	}
}

func bankAccountToolPayload(account domain.BankAccount) map[string]any {
	return map[string]any{
		"id":             account.ID,
		"name":           account.Name,
		"institution":    account.Institution,
		"account_type":   account.AccountType,
		"last_four":      account.LastFour,
		"balance":        account.Balance,
		"currency":       account.Currency,
		"is_primary":     account.IsPrimary,
		"is_connected":   account.IsConnected,
		"last_synced_at": account.LastSyncedAt,
		"created_at":     account.CreatedAt,
	}
}
