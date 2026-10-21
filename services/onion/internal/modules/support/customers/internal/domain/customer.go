package domain

import (
	"errors"
	"time"
)

var ErrCustomerNotFound = errors.New("support customer not found")

type Customer struct {
	ID           int64
	BusinessID   int64
	Name         string
	Email        string
	Phone        string
	TotalTickets int
	LastTicketAt *time.Time
	CreatedAt    time.Time
}
