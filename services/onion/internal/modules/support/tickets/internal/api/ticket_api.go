package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/support/tickets/internal/application"
	"github.com/voxire/onion/internal/modules/support/tickets/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	supportticketsv1 "github.com/voxire/voxie/gen/go/support/tickets/v1"
)

// TicketAPI implements support.tickets.v1 for the signed-in business.
type TicketAPI struct{ Service *application.Service }

func NewTicketAPI(service *application.Service) *TicketAPI { return &TicketAPI{Service: service} }

func businessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return device.User.BusinessID, nil
}

func (a *TicketAPI) ListTickets(ctx context.Context, req *supportticketsv1.ListTicketsRequest) (*supportticketsv1.ListTicketsResponse, error) {
	business, err := businessID(ctx)
	if err != nil {
		return nil, err
	}
	page, pageSize := int(req.GetPage()), int(req.GetPageSize())
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	items, total, err := a.Service.List(ctx, business, req.GetStatus(), page, pageSize)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list support tickets")
	}
	tickets := make([]*supportticketsv1.Ticket, len(items))
	for i := range items {
		tickets[i] = ticketToProto(items[i])
	}
	return &supportticketsv1.ListTicketsResponse{
		Tickets:  tickets,
		Total:    int32(total),    // #nosec G115 -- ticket counts are bounded by the database page query.
		Page:     int32(page),     // #nosec G115 -- page comes from an int32 request field.
		PageSize: int32(pageSize), // #nosec G115 -- page size comes from an int32 request field.
	}, nil
}

func (a *TicketAPI) GetTicket(ctx context.Context, req *supportticketsv1.GetTicketRequest) (*supportticketsv1.GetTicketResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	business, err := businessID(ctx)
	if err != nil {
		return nil, err
	}
	ticket, err := a.Service.GetByID(ctx, business, req.GetId())
	if err != nil {
		return nil, supportTicketLookupError(err)
	}
	msgs, err := a.Service.ListMessages(ctx, business, req.GetId())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list ticket messages")
	}
	messages := make([]*supportticketsv1.TicketMessage, len(msgs))
	for i, message := range msgs {
		messages[i] = &supportticketsv1.TicketMessage{
			Id:         message.ID,
			TicketId:   message.TicketID,
			SenderType: message.SenderType,
			SenderName: message.SenderName,
			Body:       message.Body,
			IsInternal: message.IsInternal,
			CreatedAt:  formatTime(message.CreatedAt),
		}
	}
	var out *supportticketsv1.Ticket
	if ticket != nil {
		out = ticketToProto(*ticket)
	}
	return &supportticketsv1.GetTicketResponse{Ticket: out, Messages: messages}, nil
}

func (a *TicketAPI) CreateTicket(ctx context.Context, req *supportticketsv1.CreateTicketRequest) (*supportticketsv1.CreateTicketResponse, error) {
	if strings.TrimSpace(req.GetSubject()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Subject is required")
	}
	if strings.TrimSpace(req.GetCustomerEmail()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Customer email is required")
	}
	business, err := businessID(ctx)
	if err != nil {
		return nil, err
	}
	ticketID, err := a.Service.Create(ctx, domain.TicketInput{
		BusinessID:    business,
		CustomerName:  req.GetCustomerName(),
		CustomerEmail: req.GetCustomerEmail(),
		Subject:       req.GetSubject(),
		Priority:      req.GetPriority(),
		Channel:       req.GetChannel(),
		AssignedTo:    req.AssignedTo,
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create support ticket")
	}
	if body := req.GetBody(); body != "" {
		senderName := req.GetSenderName()
		if senderName == "" {
			senderName = req.GetCustomerName()
		}
		if _, msgErr := a.Service.CreateMessage(ctx, domain.MessageInput{
			TicketID:   ticketID,
			SenderType: "customer",
			SenderName: senderName,
			Body:       body,
		}); msgErr != nil {
			// The ticket was created successfully; the initial message is a
			// non-critical attachment. Log the failure with the ticket and
			// business context instead of silently discarding it, and still
			// return success for the ticket itself.
			logger.Ctx(ctx).Warn("Failed to create initial support ticket message",
				logger.Any("ticket_id", ticketID),
				logger.Any("business_id", business),
				logger.Error(msgErr),
			)
		}
	}
	return &supportticketsv1.CreateTicketResponse{TicketId: ticketID, Message: "Support ticket created"}, nil
}

func (a *TicketAPI) UpdateTicket(ctx context.Context, req *supportticketsv1.UpdateTicketRequest) (*supportticketsv1.UpdateTicketResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	business, err := businessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.Update(ctx, domain.TicketUpdateInput{
		BusinessID: business,
		ID:         req.GetId(),
		Subject:    req.GetSubject(),
		Status:     req.GetStatus(),
		Priority:   req.GetPriority(),
		AssignedTo: req.AssignedTo,
		Channel:    req.GetChannel(),
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update support ticket")
	}
	return &supportticketsv1.UpdateTicketResponse{Message: "Support ticket updated"}, nil
}

// supportTicketLookupError maps a ticket lookup failure to the HTTP error the
// client should see: a missing ticket is a 404, not a server failure.
func supportTicketLookupError(err error) error {
	if errors.Is(err, domain.ErrTicketNotFound) {
		return onekiterrors.New(http.StatusNotFound, "Support ticket not found")
	}
	return onekiterrors.New(http.StatusInternalServerError, "Failed to get support ticket")
}

func ticketToProto(ticket domain.Ticket) *supportticketsv1.Ticket {
	return &supportticketsv1.Ticket{
		Id:            ticket.ID,
		BusinessId:    ticket.BusinessID,
		CustomerName:  ticket.CustomerName,
		CustomerEmail: ticket.CustomerEmail,
		Subject:       ticket.Subject,
		Status:        ticket.Status,
		Priority:      ticket.Priority,
		AssignedTo:    ticket.AssignedTo,
		Channel:       ticket.Channel,
		CreatedAt:     formatTime(ticket.CreatedAt),
		UpdatedAt:     formatTime(ticket.UpdatedAt),
	}
}

// formatTime keeps the timestamp format the legacy JSON encoder produced.
func formatTime(value time.Time) string {
	return value.Format(time.RFC3339Nano)
}
