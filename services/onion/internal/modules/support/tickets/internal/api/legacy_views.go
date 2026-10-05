package api

import (
	"github.com/voxire/onion/internal/modules/support/tickets/internal/domain"
	"time"
)

type ticketPayload struct {
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

func payloadTicket(value domain.Ticket) ticketPayload { return ticketPayload(value) }

func payloadTicketPointer(value *domain.Ticket) *ticketPayload {
	if value == nil {
		return nil
	}
	result := payloadTicket(*value)
	return &result
}

func payloadTicketList(values []domain.Ticket) []ticketPayload {
	if values == nil {
		return nil
	}
	result := make([]ticketPayload, len(values))
	for i, value := range values {
		result[i] = payloadTicket(value)
	}
	return result
}

type messagePayload struct {
	ID         int64     `json:"id"`
	TicketID   int64     `json:"ticket_id"`
	SenderType string    `json:"sender_type"`
	SenderName string    `json:"sender_name"`
	Body       string    `json:"body"`
	IsInternal bool      `json:"is_internal"`
	CreatedAt  time.Time `json:"created_at"`
}

func payloadMessage(value domain.Message) messagePayload { return messagePayload(value) }

func payloadMessagePointer(value *domain.Message) *messagePayload {
	if value == nil {
		return nil
	}
	result := payloadMessage(*value)
	return &result
}

func payloadMessageList(values []domain.Message) []messagePayload {
	if values == nil {
		return nil
	}
	result := make([]messagePayload, len(values))
	for i, value := range values {
		result[i] = payloadMessage(value)
	}
	return result
}
