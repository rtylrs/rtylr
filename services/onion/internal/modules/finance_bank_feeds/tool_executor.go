package financebankfeeds

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listFinanceBankFeedTransactionsToolName = "finance.list_bank_feed_transactions"
	defaultFinanceBankFeedToolLimit         = int64(20)
	maxFinanceBankFeedToolLimit             = int64(50)
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != listFinanceBankFeedTransactionsToolName {
		return nil, false, nil
	}

	businessID, err := financeBankFeedToolBusinessID(ctx)
	if err != nil {
		return nil, true, err
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || !authPkg.HasAppPermission(authUser, authorization.AppFinance, "finance:bank_feeds:view") {
		return nil, true, errors.New("permission denied")
	}
	includeSensitive := authPkg.HasAppPermission(authUser, authorization.AppFinance, "finance:bank_feeds:sensitive_view")
	limit, offset, err := financeBankFeedToolPagination(input)
	if err != nil {
		return nil, true, err
	}

	transactions, total, err := m.service.List(ctx, businessID, limit, offset)
	if err != nil {
		return nil, true, err
	}

	payloads := make([]map[string]any, 0, len(transactions))
	matchedCount := 0
	for _, transaction := range transactions {
		if transaction.IsMatched() {
			matchedCount++
		}
		payloads = append(payloads, financeBankFeedTransactionToolPayload(transaction, includeSensitive))
	}

	return map[string]any{
		"transactions":  payloads,
		"total":         total,
		"returned":      len(payloads),
		"matched_count": matchedCount,
		"limit":         limit,
		"offset":        offset,
	}, true, nil
}

func financeBankFeedToolBusinessID(ctx context.Context) (int64, error) {
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

func financeBankFeedToolPagination(input map[string]any) (int32, int32, error) {
	limit, err := financeBankFeedToolNonNegativeInt64(firstFinanceBankFeedToolValue(input, "limit", "page_size"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultFinanceBankFeedToolLimit
	}
	if limit > maxFinanceBankFeedToolLimit {
		limit = maxFinanceBankFeedToolLimit
	}
	offset, err := financeBankFeedToolNonNegativeInt64(firstFinanceBankFeedToolValue(input, "offset"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid offset: %w", err)
	}
	if offset > math.MaxInt32 {
		return 0, 0, errors.New("offset is too large")
	}
	return int32(limit), int32(offset), nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func firstFinanceBankFeedToolValue(input map[string]any, keys ...string) any {
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

func financeBankFeedToolNonNegativeInt64(value any) (int64, error) {
	parsed, err := financeBankFeedToolInt64(value)
	if err != nil {
		return 0, err
	}
	if parsed < 0 {
		return 0, errors.New("value must be non-negative")
	}
	return parsed, nil
}

func financeBankFeedToolInt64(value any) (int64, error) {
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

func financeBankFeedTransactionToolPayload(transaction domain.BankFeedTransaction, includeSensitive bool) map[string]any {
	payload := map[string]any{
		"id":                 transaction.ID,
		"amount_minor_units": transaction.AmountMinorUnits,
		"amount":             transaction.AmountDecimal(100),
		"currency":           transaction.Currency,
		"occurred_at":        transaction.OccurredAt,
		"description":        transaction.Description,
		"is_matched":         transaction.IsMatched(),
		"is_credit":          transaction.IsCredit(),
		"is_debit":           transaction.IsDebit(),
	}
	if includeSensitive {
		payload["bank_account_id"] = transaction.BankAccountID
		payload["external_id"] = transaction.ExternalID
		payload["matched_expense_id"] = transaction.MatchedExpenseID
		payload["matched_invoice_id"] = transaction.MatchedInvoiceID
	}
	return payload
}
