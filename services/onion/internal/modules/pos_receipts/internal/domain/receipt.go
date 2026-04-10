package domain

import "time"

type Receipt struct {
	ID            int64     `json:"id"`
	OrderNumber   string    `json:"orderNumber"`
	CustomerName  string    `json:"customerName"`
	Total         float64   `json:"total"`
	PaymentMethod string    `json:"paymentMethod"`
	PaidAt        time.Time `json:"paidAt"`
}
