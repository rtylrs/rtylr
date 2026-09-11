package infrastructure

import (
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	sqlc "github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure/persistence/sqlc"
)

// nullPaymentMethod maps a domain payment method onto the nullable ENUM column.
//
// pos_orders.payment_method is ENUM('cash','card','online','split') with no
// empty member, so an order that has not been paid yet has to be written as
// NULL. Marking the value Valid regardless sends ” and MySQL answers
// "Error 1265 (01000): Data truncated for column 'payment_method'" (#1185).
//
// That is what SplitOrder did: it copied the source order's method and marked
// it Valid unconditionally, so splitting an unpaid order always failed. The
// normal create path already had the guard, which is why only the split broke.
//
// A non-empty value that is not an enum member is deliberately passed through
// rather than nulled. Writing NULL for a typo would silently drop what the
// caller asked for; letting the column reject it keeps the mistake visible.
func nullPaymentMethod(method domain.PaymentMethod) sqlc.NullPosOrdersPaymentMethod {
	return sqlc.NullPosOrdersPaymentMethod{
		PosOrdersPaymentMethod: sqlc.PosOrdersPaymentMethod(method),
		Valid:                  method != "",
	}
}
