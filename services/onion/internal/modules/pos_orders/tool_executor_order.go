package pos_orders

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
)

func parseGetOrderToolInput(input map[string]any) (int64, error) {
	orderID, err := posOrderIDToolValue(firstPosToolValue(input, "order_id", "order", "id"))
	if err != nil {
		return 0, fmt.Errorf("invalid order_id: %w", err)
	}
	if orderID <= 0 {
		return 0, errors.New("order_id is required")
	}
	return orderID, nil
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

func normalizeOrderToolStatus(value any) string {
	status := strings.ToLower(strings.TrimSpace(posToolString(value)))
	if status == "" {
		return ""
	}
	status = strings.ReplaceAll(status, "-", "_")
	status = strings.ReplaceAll(status, " ", "_")
	switch domain.OrderStatus(status) {
	case domain.OrderStatusPending,
		domain.OrderStatusConfirmed,
		domain.OrderStatusPreparing,
		domain.OrderStatusReady,
		domain.OrderStatusServed,
		domain.OrderStatusPaid,
		domain.OrderStatusCancelled,
		domain.OrderStatusRefunded,
		domain.OrderStatusVoided,
		domain.OrderStatusHeld:
		return status
	default:
		return ""
	}
}

func validateOrderToolDateRange(dateFrom, dateTo string) error {
	if dateFrom == "" || dateTo == "" {
		return nil
	}
	parsedFrom, err := time.Parse(posToolCalendarDate, dateFrom)
	if err != nil {
		return fmt.Errorf("invalid date_from: %w", err)
	}
	parsedTo, err := time.Parse(posToolCalendarDate, dateTo)
	if err != nil {
		return fmt.Errorf("invalid date_to: %w", err)
	}
	if parsedTo.Before(parsedFrom) {
		return errors.New("date_to must be on or after date_from")
	}
	return nil
}

func orderDetailToolPayload(order domain.Order) map[string]any {
	payload := orderToolPayload(order)
	payload["notes"] = order.Notes
	payload["kitchen_notes"] = order.KitchenNotes
	payload["updated_at"] = formatOrderToolTime(order.UpdatedAt)
	payload["terminal_id"] = order.TerminalID
	payload["branch_id"] = order.BranchID
	payload["tip_amount"] = order.TipAmount
	payload["void_reason"] = order.VoidReason
	payload["voided_by"] = order.VoidedBy
	payload["voided_at"] = formatOrderToolOptionalTime(order.VoidedAt)
	payload["exchange_order_id"] = order.ExchangeOrderID
	return payload
}

func orderToolPayload(order domain.Order) map[string]any {
	return map[string]any{
		"id":              order.ID,
		"session_id":      order.SessionID,
		"table_id":        order.TableID,
		"location_id":     order.LocationID,
		"order_number":    order.OrderNumber,
		"type":            string(order.Type),
		"status":          string(order.Status),
		"customer_id":     order.CustomerID,
		"customer_name":   order.CustomerName,
		"staff_id":        order.StaffID,
		"subtotal":        order.Subtotal,
		"discount_amount": order.DiscountAmount,
		"tax_amount":      order.TaxAmount,
		"total":           order.Total,
		"payment_method":  string(order.PaymentMethod),
		"payment_status":  string(order.PaymentStatus),
		"amount_paid":     order.AmountPaid,
		"change_due":      order.ChangeDue,
		"item_count":      len(order.Items),
		"created_at":      formatOrderToolTime(order.CreatedAt),
		"paid_at":         formatOrderToolOptionalTime(order.PaidAt),
	}
}

func orderItemToolPayload(item domain.OrderItem) map[string]any {
	return map[string]any{
		"id":              item.ID,
		"order_id":        item.OrderID,
		"menu_item_id":    item.MenuItemID,
		"name":            item.Name,
		"sku":             item.SKU,
		"quantity":        item.Quantity,
		"unit_price":      item.UnitPrice,
		"discount_pct":    item.DiscountPct,
		"discount_amount": item.DiscountAmount,
		"tax_rate":        item.TaxRate,
		"total":           item.Total,
		"notes":           item.Notes,
		"status":          string(item.Status),
		"created_at":      formatOrderToolTime(item.CreatedAt),
	}
}

func formatOrderToolTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(posToolTimestampDate)
}

func formatOrderToolOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return formatOrderToolTime(*value)
}
