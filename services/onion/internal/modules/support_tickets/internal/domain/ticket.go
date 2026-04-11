package domain

import "time"

type Ticket struct {
	ID            int64     `json:"id"`
	BusinessID    int64     `json:"businessId"`
	CustomerName  string    `json:"customerName"`
	CustomerEmail string    `json:"customerEmail"`
	Subject       string    `json:"subject"`
	Status        string    `json:"status"`
	Priority      string    `json:"priority"`
	AssignedTo    *int64    `json:"assignedTo"`
	Channel       string    `json:"channel"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
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
	TicketID   int64     `json:"ticketId"`
	SenderType string    `json:"senderType"`
	SenderName string    `json:"senderName"`
	Body       string    `json:"body"`
	IsInternal bool      `json:"isInternal"`
	CreatedAt  time.Time `json:"createdAt"`
}

type MessageInput struct {
	TicketID   int64
	SenderType string
	SenderName string
	Body       string
	IsInternal bool
}
