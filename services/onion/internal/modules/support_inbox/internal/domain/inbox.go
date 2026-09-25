package domain

import "time"

type InboxItem struct {
	TicketID      int64     `json:"ticket_id"`
	CustomerName  string    `json:"customer_name"`
	CustomerEmail string    `json:"customer_email"`
	Subject       string    `json:"subject"`
	Status        string    `json:"status"`
	Priority      string    `json:"priority"`
	Channel       string    `json:"channel"`
	LastMessage   string    `json:"last_message"`
	LastMessageAt time.Time `json:"last_message_at"`
	MessageCount  int       `json:"message_count"`
}

type UnreadCount struct {
	Count int `json:"count"`
}
