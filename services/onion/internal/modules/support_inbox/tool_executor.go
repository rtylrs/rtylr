package supportinbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/support_inbox/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const (
	listSupportInboxToolName         = "support.list_inbox"
	getSupportUnreadCountToolName    = "support.get_unread_count"
	defaultListSupportInboxToolLimit = int64(20)
	maxListSupportInboxToolLimit     = int64(50)
	supportInboxToolTimestampDate    = "2006-01-02T15:04:05Z"
)

type listSupportInboxToolInput struct {
	Page  int64
	Limit int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listSupportInboxToolName:
		output, err := m.executeListSupportInboxTool(ctx, input)
		return output, true, err
	case getSupportUnreadCountToolName:
		output, err := m.executeGetSupportUnreadCountTool(ctx)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListSupportInboxTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := supportInboxToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := parseListSupportInboxToolInput(input)
	if err != nil {
		return nil, err
	}

	items, total, err := m.service.ListInbox(ctx, businessID, int(parsed.Page), int(parsed.Limit))
	if err != nil {
		return nil, err
	}
	unreadCount, err := m.service.UnreadCount(ctx, businessID)
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, supportInboxItemToolPayload(item))
	}

	return map[string]any{
		"inbox":        payloads,
		"total":        total,
		"returned":     len(payloads),
		"unread_count": unreadCount,
		"page":         parsed.Page,
		"page_size":    parsed.Limit,
		"has_more":     parsed.Page*parsed.Limit < int64(total),
	}, nil
}

func (m *Module) executeGetSupportUnreadCountTool(ctx context.Context) (map[string]any, error) {
	businessID, err := supportInboxToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	count, err := m.service.UnreadCount(ctx, businessID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"unread_count": count,
	}, nil
}

func parseListSupportInboxToolInput(input map[string]any) (listSupportInboxToolInput, error) {
	page, err := supportInboxToolNonNegativeInt64(firstSupportInboxToolValue(input, "page"))
	if err != nil {
		return listSupportInboxToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := supportInboxToolNonNegativeInt64(firstSupportInboxToolValue(input, "limit", "page_size"))
	if err != nil {
		return listSupportInboxToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListSupportInboxToolLimit
	}
	if limit > maxListSupportInboxToolLimit {
		limit = maxListSupportInboxToolLimit
	}

	return listSupportInboxToolInput{Page: page, Limit: limit}, nil
}

func supportInboxToolBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User != nil && device.User.BusinessID > 0 {
		return device.User.BusinessID, nil
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if ok && authUser != nil && authUser.BusinessID > 0 {
		return authUser.BusinessID, nil
	}
	return 0, errors.New("authentication is required")
}

func supportInboxItemToolPayload(item domain.InboxItem) map[string]any {
	return map[string]any{
		"ticket_id":       item.TicketID,
		"customer_name":   item.CustomerName,
		"customer_email":  item.CustomerEmail,
		"subject":         item.Subject,
		"status":          item.Status,
		"priority":        item.Priority,
		"channel":         item.Channel,
		"last_message":    item.LastMessage,
		"last_message_at": formatSupportInboxToolTime(item.LastMessageAt),
		"message_count":   item.MessageCount,
	}
}

func firstSupportInboxToolValue(input map[string]any, keys ...string) any {
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

func supportInboxToolNonNegativeInt64(value any) (int64, error) {
	out, err := supportInboxToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func supportInboxToolInt64(value any) (int64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		if math.Trunc(v) != v {
			return 0, errors.New("must be an integer")
		}
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		return strconv.ParseInt(raw, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func formatSupportInboxToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(supportInboxToolTimestampDate)
}
