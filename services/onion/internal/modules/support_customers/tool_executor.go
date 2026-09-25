package supportcustomers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/support_customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const (
	listSupportCustomersToolName         = "support.list_customers"
	getSupportCustomerToolName           = "support.get_customer"
	defaultListSupportCustomersToolLimit = int64(20)
	maxListSupportCustomersToolLimit     = int64(50)
	supportCustomerToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type listSupportCustomersToolInput struct {
	Page  int64
	Limit int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listSupportCustomersToolName:
		output, err := m.executeListSupportCustomersTool(ctx, input)
		return output, true, err
	case getSupportCustomerToolName:
		output, err := m.executeGetSupportCustomerTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListSupportCustomersTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := supportCustomersToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := parseListSupportCustomersToolInput(input)
	if err != nil {
		return nil, err
	}

	customers, total, err := m.service.List(ctx, businessID, int(parsed.Page), int(parsed.Limit))
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(customers))
	for _, customer := range customers {
		payloads = append(payloads, supportCustomerToolPayload(customer))
	}

	return map[string]any{
		"customers": payloads,
		"total":     total,
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  parsed.Page*parsed.Limit < int64(total),
	}, nil
}

func (m *Module) executeGetSupportCustomerTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := supportCustomersToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	customerID, err := supportCustomerToolInt64(firstSupportCustomerToolValue(input, "customer_id", "id"))
	if err != nil {
		return nil, fmt.Errorf("invalid customer_id: %w", err)
	}
	if customerID <= 0 {
		return nil, errors.New("customer_id is required")
	}

	customer, err := m.service.GetByID(ctx, businessID, customerID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, errors.New("customer not found")
	}

	return map[string]any{
		"customer": supportCustomerToolPayload(*customer),
	}, nil
}

func parseListSupportCustomersToolInput(input map[string]any) (listSupportCustomersToolInput, error) {
	page, err := supportCustomerToolNonNegativeInt64(firstSupportCustomerToolValue(input, "page"))
	if err != nil {
		return listSupportCustomersToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := supportCustomerToolNonNegativeInt64(firstSupportCustomerToolValue(input, "limit", "page_size"))
	if err != nil {
		return listSupportCustomersToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListSupportCustomersToolLimit
	}
	if limit > maxListSupportCustomersToolLimit {
		limit = maxListSupportCustomersToolLimit
	}

	return listSupportCustomersToolInput{Page: page, Limit: limit}, nil
}

func supportCustomersToolBusinessID(ctx context.Context) (int64, error) {
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

func supportCustomerToolPayload(customer domain.Customer) map[string]any {
	return map[string]any{
		"id":            customer.ID,
		"business_id":   customer.BusinessID,
		"name":          customer.Name,
		"email":         customer.Email,
		"phone":         customer.Phone,
		"total_tickets": customer.TotalTickets,
		"last_ticket_at": formatSupportCustomerToolOptionalTime(
			customer.LastTicketAt,
		),
		"created_at": formatSupportCustomerToolTime(customer.CreatedAt),
	}
}

func firstSupportCustomerToolValue(input map[string]any, keys ...string) any {
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

func supportCustomerToolNonNegativeInt64(value any) (int64, error) {
	out, err := supportCustomerToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func supportCustomerToolInt64(value any) (int64, error) {
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

func formatSupportCustomerToolOptionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatSupportCustomerToolTime(*value)
}

func formatSupportCustomerToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(supportCustomerToolTimestampDate)
}
