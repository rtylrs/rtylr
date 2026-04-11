package domain

import "time"

type InboxItem struct {
	TicketID      int64     `json:"ticketId"`
	CustomerName  string    `json:"customerName"`
	CustomerEmail string    `json:"customerEmail"`
	Subject       string    `json:"subject"`
	Status        string    `json:"status"`
	Priority      string    `json:"priority"`
	Channel       string    `json:"channel"`
	LastMessage   string    `json:"lastMessage"`
	LastMessageAt time.Time `json:"lastMessageAt"`
	MessageCount  int       `json:"messageCount"`
}

type UnreadCount struct {
	Count int `json:"count"`
}
