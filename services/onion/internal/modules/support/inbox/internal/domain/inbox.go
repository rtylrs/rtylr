package domain

import "time"

type InboxItem struct {
	TicketID      int64
	CustomerName  string
	CustomerEmail string
	Subject       string
	Status        string
	Priority      string
	Channel       string
	LastMessage   string
	LastMessageAt time.Time
	MessageCount  int
}

type UnreadCount struct {
	Count int
}
