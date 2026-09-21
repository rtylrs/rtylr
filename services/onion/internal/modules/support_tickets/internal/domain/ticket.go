package domain

import (
	"errors"
	"time"
)

var ErrTicketNotFound = errors.New("support ticket not found")

type Ticket struct {
	ID            int64     `json:"id"`
	BusinessID    int64     `json:"business_id"`
	CustomerName  string    `json:"customer_name"`
	CustomerEmail string    `json:"customer_email"`
	Subject       string    `json:"subject"`
	Status        string    `json:"status"`
	Priority      string    `json:"priority"`
	AssignedTo    *int64    `json:"assigned_to"`
	Channel       string    `json:"channel"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type TicketInput struct {
	BusinessID    int64
	CustomerName  string
	CustomerEmail string
	Subject       string
	Status        string
	Priority      string
	AssignedTo    *int64
	Channel       string
}

type TicketUpdateInput struct {
	BusinessID int64
	ID         int64
	Subject    string
	Status     string
	Priority   string
	AssignedTo *int64
	Channel    string
}

type Message struct {
	ID         int64     `json:"id"`
	TicketID   int64     `json:"ticket_id"`
	SenderType string    `json:"sender_type"`
	SenderName string    `json:"sender_name"`
	Body       string    `json:"body"`
	IsInternal bool      `json:"is_internal"`
	CreatedAt  time.Time `json:"created_at"`
}

type MessageInput struct {
	TicketID   int64
	SenderType string
	SenderName string
	Body       string
	IsInternal bool
}
