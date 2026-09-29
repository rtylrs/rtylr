package infrastructure

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/money"
)

// ─── Refund money ─────────────────────────────────────────────────────────────
//
// A refund is bounded by how much cash the order actually left the merchant
// with, and every comparison is made in integer cents. The previous float64
// form failed closed: 0.03 + 0.27 evaluates to 0.30000000000000004 in binary
// floating point, which is greater than 0.30, so the final cent of a perfectly
// legitimate full refund was refused with a 400.

// refundCapCents is the most that may ever be returned against an order.
//
// The cap is the net cash the merchant kept, GREATEST(total, amount_paid -
// change_due):
//
//   - amount_paid alone is wrong in both directions. Cash tender carries the
//     change the customer was handed straight back, so a $20 payment on a $10
//     order retained $10, and capping at $20 would pay out money the drawer
//     never had.
//   - total alone is wrong too. pos_orders.tip_amount is never written by any
//     query, so a tip-inclusive tender lands in amount_paid and not in total;
//     capping at total makes the tip permanently unrefundable (a legitimate
//     tip-inclusive refund answers 400). amount_paid - change_due is the tip
//     plus the sale, so it is the cap that survives the tip being persisted
//     later without change.
//   - total is the floor for orders settled without a cash amount, such as a
//     Paddle payment that writes payment_status but leaves amount_paid at its
//     column default of 0. Without the floor those orders could never be
//     refunded at all.
//
// A zero cap is returned as an error: an order whose money columns cannot be
// read must fail closed rather than accept an unbounded refund.
func refundCapCents(order sqlc.GetOrderForRefundRow) (money.Amount, error) {
	total, err := decimalColumnCents(order.Total)
	if err != nil {
		return 0, fmt.Errorf("order total: %w", err)
	}
	amountPaid, err := decimalColumnCents(order.AmountPaid)
	if err != nil {
		return 0, fmt.Errorf("order amount_paid: %w", err)
	}
	changeDue, err := decimalColumnCents(order.ChangeDue)
	if err != nil {
		return 0, fmt.Errorf("order change_due: %w", err)
	}

	capCents := amountPaid - changeDue
	if total > capCents {
		capCents = total
	}
	if capCents <= 0 {
		return 0, errors.New("order has no refundable amount")
	}
	return capCents, nil
}

// sumRefundCents reads a SUM over pos_refunds.amount back as integer cents.
// SumRefundsForOrder is typed as interface{} because it aggregates a DECIMAL
// with a literal zero, and the driver hands that back as a []byte, a string or
// a float64 depending on the value; all three are parsed, never summed as
// floats.
func sumRefundCents(sum any) (money.Amount, error) {
	if sum == nil {
		return 0, nil
	}
	var text string
	switch v := sum.(type) {
	case []byte:
		text = string(v)
	case string:
		text = v
	case float64:
		cents, err := money.FromFloat64(v)
		if err != nil {
			return 0, err
		}
		return cents, nil
	case int64:
		return money.Amount(v * 100), nil
	default:
		return 0, fmt.Errorf("unsupported refund sum type %T", sum)
	}
	if strings.TrimSpace(text) == "" {
		return 0, nil
	}
	return money.FromString(text)
}

// decimalColumnCents parses a nullable DECIMAL(10,2) column into integer cents.
// An absent or empty column is zero, matching the column defaults; anything
// present but unparseable is an error so the caller fails closed.
func decimalColumnCents(value sql.NullString) (money.Amount, error) {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return 0, nil
	}
	return money.FromString(value.String)
}

// refundPaymentStatusFor is the payment_status an order carries once
// totalRefundedCents more has been returned against a cap of capCents.
//
// Reaching the cap retires the sale. Falling short of it leaves the order a
// live paid sale: GetSessionTotals, GetDashboardSalesToday, the day summary,
// receipts and insights all count payment_status = 'paid', so an
// unconditional write meant a $0.01 refund deleted a $100 sale from the books
// while pos_refunds recorded only $0.01.
//
// The enum's 'partial' value is deliberately not returned. It reads as
// "part of the order is settled", but every revenue query filters on
// payment_status = 'paid', so writing 'partial' would drop the order from
// precisely the same reports the fix is meant to keep it in.
func refundPaymentStatusFor(totalRefundedCents, capCents money.Amount) domain.PaymentStatus {
	if totalRefundedCents >= capCents {
		return domain.PaymentStatusRefunded
	}
	return domain.PaymentStatusPaid
}
