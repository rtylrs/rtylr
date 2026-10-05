package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/support/tickets/internal/application"
	"github.com/voxire/onion/internal/modules/support/tickets/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

// -- LIST --------------------------------------------------------------------

func ListTicketsRoutes() []string { return []string{"/support/tickets"} }

type ListTicketsHandler struct{ Service *application.Service }

type ListTicketsRequest struct {
	Status   string `json:"status"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type ListTicketsResponse struct {
	Tickets  []ticketPayload `json:"tickets"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

func (h *ListTicketsHandler) Serve(ctx context.Context, req *ListTicketsRequest) (ListTicketsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListTicketsResponse{}, onekiterrors.PermissionDenied()
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	items, total, err := h.Service.List(ctx, device.User.BusinessID, req.Status, req.Page, req.PageSize)
	if err != nil {
		return ListTicketsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list support tickets")
	}
	return ListTicketsResponse{Tickets: payloadTicketList(items), Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// -- GET DETAIL --------------------------------------------------------------

func GetTicketRoutes() []string { return []string{"/support/tickets/detail"} }

type GetTicketHandler struct{ Service *application.Service }

type GetTicketRequest struct {
	ID int64 `json:"id"`
}

type GetTicketResponse struct {
	Ticket   *ticketPayload   `json:"ticket"`
	Messages []messagePayload `json:"messages"`
}

func (h *GetTicketHandler) Validate(_ context.Context, req *GetTicketRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	return nil
}

func (h *GetTicketHandler) Serve(ctx context.Context, req *GetTicketRequest) (GetTicketResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetTicketResponse{}, onekiterrors.PermissionDenied()
	}
	ticket, err := h.Service.GetByID(ctx, device.User.BusinessID, req.ID)
	if err != nil {
		return GetTicketResponse{}, supportTicketLookupError(err)
	}
	msgs, err := h.Service.ListMessages(ctx, device.User.BusinessID, req.ID)
	if err != nil {
		return GetTicketResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list ticket messages")
	}
	return GetTicketResponse{Ticket: payloadTicketPointer(ticket), Messages: payloadMessageList(msgs)}, nil
}

// supportTicketLookupError maps a ticket lookup failure to the HTTP error the
// client should see: a missing ticket is a 404, not a server failure.
func supportTicketLookupError(err error) error {
	if errors.Is(err, domain.ErrTicketNotFound) {
		return onekiterrors.New(http.StatusNotFound, "Support ticket not found")
	}
	return onekiterrors.New(http.StatusInternalServerError, "Failed to get support ticket")
}

// -- CREATE ------------------------------------------------------------------

func CreateTicketRoutes() []string { return []string{"/support/tickets"} }

type CreateTicketHandler struct{ Service *application.Service }

type CreateTicketRequest struct {
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
	Subject       string `json:"subject"`
	Priority      string `json:"priority"`
	Channel       string `json:"channel"`
	AssignedTo    *int64 `json:"assigned_to"`
	Body          string `json:"body"`
	SenderName    string `json:"sender_name"`
}

type CreateTicketResponse struct {
	TicketID int64  `json:"ticket_id"`
	Message  string `json:"message"`
}

func (h *CreateTicketHandler) Validate(_ context.Context, req *CreateTicketRequest) error {
	if strings.TrimSpace(req.Subject) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Subject is required")
	}
	if strings.TrimSpace(req.CustomerEmail) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Customer email is required")
	}
	return nil
}

func (h *CreateTicketHandler) Serve(ctx context.Context, req *CreateTicketRequest) (CreateTicketResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateTicketResponse{}, onekiterrors.PermissionDenied()
	}
	ticketID, err := h.Service.Create(ctx, domain.TicketInput{
		BusinessID:    device.User.BusinessID,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		Subject:       req.Subject,
		Priority:      req.Priority,
		Channel:       req.Channel,
		AssignedTo:    req.AssignedTo,
	})
	if err != nil {
		return CreateTicketResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create support ticket")
	}
	if req.Body != "" {
		senderName := req.SenderName
		if senderName == "" {
			senderName = req.CustomerName
		}
		if _, msgErr := h.Service.CreateMessage(ctx, domain.MessageInput{
			TicketID:   ticketID,
			SenderType: "customer",
			SenderName: senderName,
			Body:       req.Body,
		}); msgErr != nil {
			// The ticket was created successfully; the initial message is a
			// non-critical attachment. Log the failure with the ticket and
			// business context instead of silently discarding it, and still
			// return success for the ticket itself.
			logger.Ctx(ctx).Warn("Failed to create initial support ticket message",
				logger.Any("ticket_id", ticketID),
				logger.Any("business_id", device.User.BusinessID),
				logger.Error(msgErr),
			)
		}
	}
	return CreateTicketResponse{TicketID: ticketID, Message: "Support ticket created"}, nil
}

// -- UPDATE ------------------------------------------------------------------

func UpdateTicketRoutes() []string { return []string{"/support/tickets"} }

type UpdateTicketHandler struct{ Service *application.Service }

type UpdateTicketRequest struct {
	ID         int64  `json:"id"`
	Subject    string `json:"subject"`
	Status     string `json:"status"`
	Priority   string `json:"priority"`
	AssignedTo *int64 `json:"assigned_to"`
	Channel    string `json:"channel"`
}

type UpdateTicketResponse struct {
	Message string `json:"message"`
}

func (h *UpdateTicketHandler) Validate(_ context.Context, req *UpdateTicketRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	return nil
}

func (h *UpdateTicketHandler) Serve(ctx context.Context, req *UpdateTicketRequest) (UpdateTicketResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateTicketResponse{}, onekiterrors.PermissionDenied()
	}
	err := h.Service.Update(ctx, domain.TicketUpdateInput{
		BusinessID: device.User.BusinessID,
		ID:         req.ID,
		Subject:    req.Subject,
		Status:     req.Status,
		Priority:   req.Priority,
		AssignedTo: req.AssignedTo,
		Channel:    req.Channel,
	})
	if err != nil {
		return UpdateTicketResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update support ticket")
	}
	return UpdateTicketResponse{Message: "Support ticket updated"}, nil
}
