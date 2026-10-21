package journal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/finance/journal/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listFinanceJournalEntriesToolName = "finance.list_journal_entries"
	defaultFinanceJournalPage         = int64(1)
	defaultFinanceJournalPageSize     = int64(20)
	maxFinanceJournalPageSize         = int64(50)
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != listFinanceJournalEntriesToolName {
		return nil, false, nil
	}

	businessID, err := financeJournalToolBusinessID(ctx)
	if err != nil {
		return nil, true, err
	}
	page, pageSize, err := financeJournalToolPagination(input)
	if err != nil {
		return nil, true, err
	}

	entries, hasMore, err := m.service.List(ctx, businessID, page, pageSize)
	if err != nil {
		return nil, true, err
	}

	payloads := make([]map[string]any, 0, len(entries))
	posted := 0
	draft := 0
	for _, entry := range entries {
		if entry.IsPosted() {
			posted++
		}
		if entry.IsDraft() {
			draft++
		}
		payloads = append(payloads, financeJournalEntryToolPayload(entry))
	}

	return map[string]any{
		"entries":      payloads,
		"returned":     len(payloads),
		"posted_count": posted,
		"draft_count":  draft,
		"page":         page,
		"page_size":    pageSize,
		"has_more":     hasMore,
	}, true, nil
}

func financeJournalToolBusinessID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return 0, errors.New("authentication is required")
	}
	return businessID, nil
}

func financeJournalToolPagination(input map[string]any) (int64, int64, error) {
	page, err := financeJournalToolNonNegativeInt64(firstFinanceJournalToolValue(input, "page"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = defaultFinanceJournalPage
	}
	pageSize, err := financeJournalToolNonNegativeInt64(firstFinanceJournalToolValue(input, "page_size", "limit"))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid page_size: %w", err)
	}
	if pageSize == 0 {
		pageSize = defaultFinanceJournalPageSize
	}
	if pageSize > maxFinanceJournalPageSize {
		pageSize = maxFinanceJournalPageSize
	}
	return page, pageSize, nil
}

func firstFinanceJournalToolValue(input map[string]any, keys ...string) any {
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

func financeJournalToolNonNegativeInt64(value any) (int64, error) {
	parsed, err := financeJournalToolInt64(value)
	if err != nil {
		return 0, err
	}
	if parsed < 0 {
		return 0, errors.New("value must be non-negative")
	}
	return parsed, nil
}

func financeJournalToolInt64(value any) (int64, error) {
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

func financeJournalEntryToolPayload(entry domain.JournalEntry) map[string]any {
	lines := make([]map[string]any, 0, len(entry.Lines))
	for _, line := range entry.Lines {
		lines = append(lines, map[string]any{
			"id":         line.ID,
			"entry_id":   line.EntryID,
			"account_id": line.AccountID,
			"debit":      line.Debit,
			"credit":     line.Credit,
		})
	}
	return map[string]any{
		"id":            entry.ID,
		"date":          entry.Date,
		"description":   entry.Description,
		"status":        entry.Status,
		"posted_at":     entry.PostedAt,
		"created_at":    entry.CreatedAt,
		"line_count":    len(lines),
		"total_debits":  entry.TotalDebits(),
		"total_credits": entry.TotalCredits(),
		"is_posted":     entry.IsPosted(),
		"is_draft":      entry.IsDraft(),
		"lines":         lines,
	}
}
