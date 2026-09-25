package pos_day_summary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const getPOSDaySummaryToolName = "pos.get_day_summary"

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != getPOSDaySummaryToolName {
		return nil, false, nil
	}
	authUser, businessID, err := posDaySummaryToolAuth(ctx)
	if err != nil {
		return nil, true, err
	}
	date, err := posDaySummaryToolDate(input)
	if err != nil {
		return nil, true, err
	}
	summary, err := m.service.GetDaySummary(ctx, businessID, date)
	if err != nil {
		return nil, true, err
	}
	return map[string]any{
		"date":             date,
		"business_id":      authUser.BusinessID,
		"summary":          summary,
		"net_revenue":      summary.NetRevenue(),
		"net_sales_ex_tax": summary.NetSalesAfterTax(),
		"average_order":    summary.AverageOrderValue(),
		"cash_sales":       summary.CashSales(),
		"card_sales":       summary.CardSales(),
		"refunds_recorded": summary.HasRefunds(),
	}, true, nil
}

func posDaySummaryToolAuth(ctx context.Context) (*authPkg.AuthUser, int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, 0, errors.New("authentication is required")
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return nil, 0, errors.New("failed to resolve business scope")
	}
	return authUser, businessID, nil
}

func posDaySummaryToolDate(input map[string]any) (string, error) {
	date := strings.TrimSpace(posDaySummaryToolString(firstPOSDaySummaryToolValue(input, "date", "day")))
	if date == "" || strings.EqualFold(date, "today") {
		date = time.Now().UTC().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return "", fmt.Errorf("date must be in YYYY-MM-DD format")
	}
	return date, nil
}

func firstPOSDaySummaryToolValue(input map[string]any, keys ...string) any {
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

func posDaySummaryToolString(value any) string {
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
