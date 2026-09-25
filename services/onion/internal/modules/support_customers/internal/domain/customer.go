package domain

import (
	"errors"
	"time"
)

var ErrCustomerNotFound = errors.New("support customer not found")

type Customer struct {
	ID           int64      `json:"id"`
	BusinessID   int64      `json:"business_id"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	TotalTickets int        `json:"total_tickets"`
	LastTicketAt *time.Time `json:"last_ticket_at"`
	CreatedAt    time.Time  `json:"created_at"`
}
