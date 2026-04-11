package domain

import "time"

type Customer struct {
	ID           int64      `json:"id"`
	BusinessID   int64      `json:"businessId"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	TotalTickets int        `json:"totalTickets"`
	LastTicketAt *time.Time `json:"lastTicketAt"`
	CreatedAt    time.Time  `json:"createdAt"`
}
