package domain

import "time"

type Receipt struct {
	ID            int64     `json:"id"`
	OrderNumber   string    `json:"order_number"`
	CustomerName  string    `json:"customer_name"`
	Total         float64   `json:"total"`
	PaymentMethod string    `json:"payment_method"`
	PaidAt        time.Time `json:"paid_at"`
}
