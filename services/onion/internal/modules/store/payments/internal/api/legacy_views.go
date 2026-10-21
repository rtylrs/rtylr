package api

import (
	"github.com/voxire/onion/internal/modules/store/payments/internal/domain"
	"time"
)

type paymentMethodPayload struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	Provider   string    `json:"provider"`
	Config     string    `json:"-"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

func payloadPaymentMethod(value domain.PaymentMethod) paymentMethodPayload {
	return paymentMethodPayload(value)
}

func payloadPaymentMethodPointer(value *domain.PaymentMethod) *paymentMethodPayload {
	if value == nil {
		return nil
	}
	result := payloadPaymentMethod(*value)
	return &result
}

func payloadPaymentMethodList(values []domain.PaymentMethod) []paymentMethodPayload {
	if values == nil {
		return nil
	}
	result := make([]paymentMethodPayload, len(values))
	for i, value := range values {
		result[i] = payloadPaymentMethod(value)
	}
	return result
}
