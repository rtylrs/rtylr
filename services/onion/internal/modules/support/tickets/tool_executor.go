package supporttickets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/support/tickets/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const (
	listSupportTicketsToolName         = "support.list_tickets"
	getSupportTicketToolName           = "support.get_ticket"
	defaultListSupportTicketsToolLimit = int64(20)
	maxListSupportTicketsToolLimit     = int64(50)
	supportTicketToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type listSupportTicketsToolInput struct {
	Status string
	Page   int64
	Limit  int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listSupportTicketsToolName:
		output, err := m.executeListSupportTicketsTool(ctx, input)
		return output, true, err
	case getSupportTicketToolName:
		output, err := m.executeGetSupportTicketTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListSupportTicketsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := supportTicketsToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := parseListSupportTicketsToolInput(input)
	if err != nil {
		return nil, err
	}

	tickets, total, err := m.service.List(ctx, businessID, parsed.Status, int(parsed.Page), int(parsed.Limit))
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(tickets))
	for _, ticket := range tickets {
		payloads = append(payloads, supportTicketToolPayload(ticket))
	}

	return map[string]any{
		"tickets":   payloads,
		"total":     total,
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  parsed.Page*parsed.Limit < int64(total),
		"filters": map[string]any{
			"status": parsed.Status,
		},
	}, nil
}

func (m *Module) executeGetSupportTicketTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := supportTicketsToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	ticketID, err := supportTicketToolInt64(firstSupportTicketToolValue(input, "ticket_id", "id"))
	if err != nil {
		return nil, fmt.Errorf("invalid ticket_id: %w", err)
	}
	if ticketID <= 0 {
		return nil, errors.New("ticket_id is required")
	}

	ticket, err := m.service.GetByID(ctx, businessID, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, errors.New("ticket not found")
	}

	messages, err := m.service.ListMessages(ctx, businessID, ticketID)
	if err != nil {
		return nil, err
	}

	messagePayloads := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		messagePayloads = append(messagePayloads, supportMessageToolPayload(message))
	}

	return map[string]any{
		"ticket":   supportTicketToolPayload(*ticket),
		"messages": messagePayloads,
	}, nil
}

func parseListSupportTicketsToolInput(input map[string]any) (listSupportTicketsToolInput, error) {
	statusRaw := firstSupportTicketToolValue(input, "status")
	status := normalizeSupportTicketToolStatus(statusRaw)
	if rawStatus := normalizeSupportTicketToolStatusToken(statusRaw); rawStatus != "" && status == "" && rawStatus != "all" && rawStatus != "any" {
		return listSupportTicketsToolInput{}, errors.New("status must be one of open, pending, resolved, closed, or all")
	}

	page, err := supportTicketToolNonNegativeInt64(firstSupportTicketToolValue(input, "page"))
	if err != nil {
		return listSupportTicketsToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := supportTicketToolNonNegativeInt64(firstSupportTicketToolValue(input, "limit", "page_size"))
	if err != nil {
		return listSupportTicketsToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListSupportTicketsToolLimit
	}
	if limit > maxListSupportTicketsToolLimit {
		limit = maxListSupportTicketsToolLimit
	}

	return listSupportTicketsToolInput{
		Status: status,
		Page:   page,
		Limit:  limit,
	}, nil
}

func supportTicketsToolBusinessID(ctx context.Context) (int64, error) {
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

func normalizeSupportTicketToolStatus(value any) string {
	status := normalizeSupportTicketToolStatusToken(value)
	switch status {
	case "", "all", "any":
		return ""
	case "open", "pending", "resolved", "closed":
		return status
	default:
		return ""
	}
}

func normalizeSupportTicketToolStatusToken(value any) string {
	status := strings.ToLower(strings.TrimSpace(supportTicketToolString(value)))
	status = strings.ReplaceAll(status, "-", "_")
	status = strings.ReplaceAll(status, " ", "_")
	return status
}

func supportTicketToolPayload(ticket domain.Ticket) map[string]any {
	return map[string]any{
		"id":             ticket.ID,
		"business_id":    ticket.BusinessID,
		"customer_name":  ticket.CustomerName,
		"customer_email": ticket.CustomerEmail,
		"subject":        ticket.Subject,
		"status":         ticket.Status,
		"priority":       ticket.Priority,
		"assigned_to":    supportTicketToolOptionalInt64(ticket.AssignedTo),
		"channel":        ticket.Channel,
		"created_at":     formatSupportTicketToolTime(ticket.CreatedAt),
		"updated_at":     formatSupportTicketToolTime(ticket.UpdatedAt),
	}
}

func supportMessageToolPayload(message domain.Message) map[string]any {
	return map[string]any{
		"id":          message.ID,
		"ticket_id":   message.TicketID,
		"sender_type": message.SenderType,
		"sender_name": message.SenderName,
		"body":        message.Body,
		"is_internal": message.IsInternal,
		"created_at":  formatSupportTicketToolTime(message.CreatedAt),
	}
}

func firstSupportTicketToolValue(input map[string]any, keys ...string) any {
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

func supportTicketToolNonNegativeInt64(value any) (int64, error) {
	out, err := supportTicketToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func supportTicketToolInt64(value any) (int64, error) {
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

func supportTicketToolString(value any) string {
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

func supportTicketToolOptionalInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func formatSupportTicketToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(supportTicketToolTimestampDate)
}
