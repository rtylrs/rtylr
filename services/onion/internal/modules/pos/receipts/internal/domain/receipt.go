package domain

import (
	"errors"
	"time"
)

var ErrReceiptNotFound = errors.New("receipt not found")

type Receipt struct {
	ID            int64
	OrderNumber   string
	CustomerName  string
	Total         float64
	PaymentMethod string
	PaidAt        time.Time
	BusinessID    int64
}
