package api

import (
	"github.com/voxire/onion/internal/modules/store/shipping/internal/domain"
	"time"
)

type shippingMethodPayload struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	BusinessID    int64     `json:"business_id"`
	Name          string    `json:"name"`
	Price         float64   `json:"price"`
	EstimatedDays int       `json:"estimated_days"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

func payloadShippingMethod(value domain.ShippingMethod) shippingMethodPayload {
	return shippingMethodPayload(value)
}

func payloadShippingMethodPointer(value *domain.ShippingMethod) *shippingMethodPayload {
	if value == nil {
		return nil
	}
	result := payloadShippingMethod(*value)
	return &result
}

func payloadShippingMethodList(values []domain.ShippingMethod) []shippingMethodPayload {
	if values == nil {
		return nil
	}
	result := make([]shippingMethodPayload, len(values))
	for i, value := range values {
		result[i] = payloadShippingMethod(value)
	}
	return result
}
