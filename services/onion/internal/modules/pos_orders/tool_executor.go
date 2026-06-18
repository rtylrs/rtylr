package pos_orders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const refundAmountTolerance = 0.000001

type refundOrderToolInput struct {
	OrderID int64
	Amount  float64
	Reason  string
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != "pos.refund_order" {
		return nil, false, nil
	}
	output, err := m.executeRefundOrderTool(ctx, input)
	return output, true, err
}

func (m *Module) executeRefundOrderTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseRefundOrderToolInput(input)
	if err != nil {
		return nil, err
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	order, err := m.service.GetOrder(ctx, ownerID, parsed.OrderID)
	if err != nil {
		return nil, err
	}
	if err := validateRefundableOrder(order, parsed.Amount); err != nil {
		return nil, err
	}

	refundID, err := m.service.ProcessRefund(ctx, ownerID, authUser.UserID, domain.RefundInput{
		OrderID: parsed.OrderID,
		Amount:  parsed.Amount,
		Reason:  parsed.Reason,
	})
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"refund": map[string]any{
			"id":                      refundID,
			"order_id":                parsed.OrderID,
			"order_number":            order.OrderNumber,
			"amount":                  parsed.Amount,
			"reason":                  parsed.Reason,
			"previous_status":         string(order.Status),
			"previous_payment_status": string(order.PaymentStatus),
		},
	}, nil
}

func parseRefundOrderToolInput(input map[string]any) (refundOrderToolInput, error) {
	orderIDValue := input["order_id"]
	if orderIDValue == nil {
		orderIDValue = input["id"]
	}
	orderID, err := posToolInt64(orderIDValue)
	if err != nil {
		return refundOrderToolInput{}, fmt.Errorf("invalid order_id: %w", err)
	}
	if orderID <= 0 {
		return refundOrderToolInput{}, errors.New("order_id is required")
	}

	amountValue := input["amount"]
	if amountValue == nil {
		amountValue = input["refund_amount"]
	}
	amount, err := posToolFloat64(amountValue)
	if err != nil {
		return refundOrderToolInput{}, fmt.Errorf("invalid amount: %w", err)
	}
	if amount <= 0 {
		return refundOrderToolInput{}, errors.New("refund amount must be positive")
	}

	reason := strings.TrimSpace(posToolString(input["reason"]))
	if reason == "" {
		reason = strings.TrimSpace(posToolString(input["notes"]))
	}
	if reason == "" {
		return refundOrderToolInput{}, errors.New("reason is required")
	}

	return refundOrderToolInput{
		OrderID: orderID,
		Amount:  amount,
		Reason:  reason,
	}, nil
}

func validateRefundableOrder(order *domain.Order, amount float64) error {
	if order == nil {
		return domain.ErrOrderNotFound
	}
	if order.Status == domain.OrderStatusRefunded || order.PaymentStatus == domain.PaymentStatusRefunded {
		return errors.New("order is already refunded")
	}
	if order.Status == domain.OrderStatusVoided || order.Status == domain.OrderStatusCancelled {
		return errors.New("order cannot be refunded in its current status")
	}
	if order.AmountPaid <= 0 && order.PaymentStatus != domain.PaymentStatusPaid {
		return errors.New("order does not have a paid amount to refund")
	}

	maxRefund := order.AmountPaid
	if maxRefund <= 0 {
		maxRefund = order.Total
	}
	if maxRefund <= 0 {
		return errors.New("order does not have a paid amount to refund")
	}
	if amount-maxRefund > refundAmountTolerance {
		return fmt.Errorf("refund amount cannot exceed %.2f", maxRefund)
	}
	return nil
}

func posToolInt64(value any) (int64, error) {
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

func posToolFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case json.Number:
		return v.Float64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		return strconv.ParseFloat(raw, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func posToolString(value any) string {
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
