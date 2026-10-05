package domain

import (
	"errors"
	"time"
)

var ErrTicketNotFound = errors.New("support ticket not found")

type Ticket struct {
	ID            int64
	BusinessID    int64
	CustomerName  string
	CustomerEmail string
	Subject       string
	Status        string
	Priority      string
	AssignedTo    *int64
	Channel       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
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
	ID         int64
	TicketID   int64
	SenderType string
	SenderName string
	Body       string
	IsInternal bool
	CreatedAt  time.Time
}

type MessageInput struct {
	TicketID   int64
	SenderType string
	SenderName string
	Body       string
	IsInternal bool
}
