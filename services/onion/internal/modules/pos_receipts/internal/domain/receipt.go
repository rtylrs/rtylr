package domain

import (
	"errors"
	"time"
)

var ErrReceiptNotFound = errors.New("receipt not found")

type Receipt struct {
	ID            int64     `json:"id"`
	OrderNumber   string    `json:"order_number"`
	CustomerName  string    `json:"customer_name"`
	Total         float64   `json:"total"`
	PaymentMethod string    `json:"payment_method"`
	PaidAt        time.Time `json:"paid_at"`
	BusinessID    int64     `json:"business_id"`
}
