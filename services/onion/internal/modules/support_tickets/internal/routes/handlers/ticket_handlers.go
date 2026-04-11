package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/support_tickets/internal/application"
	"github.com/voxire/onion/internal/modules/support_tickets/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type TicketPayload struct {
	ID            int64  `json:"id"`
	BusinessID    int64  `json:"businessId"`
	CustomerName  string `json:"customerName"`
	CustomerEmail string `json:"customerEmail"`
	Subject       string `json:"subject"`
	Status        string `json:"status"`
	Priority      string `json:"priority"`
	AssignedTo    *int64 `json:"assignedTo"`
	Channel       string `json:"channel"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type MessagePayload struct {
	ID         int64  `json:"id"`
	TicketID   int64  `json:"ticketId"`
	SenderType string `json:"senderType"`
	SenderName string `json:"senderName"`
	Body       string `json:"body"`
	IsInternal bool   `json:"isInternal"`
	CreatedAt  string `json:"createdAt"`
}

func toTicketPayload(t domain.Ticket) TicketPayload {
	return TicketPayload{
		ID: t.ID, BusinessID: t.BusinessID, CustomerName: t.CustomerName,
		CustomerEmail: t.CustomerEmail, Subject: t.Subject, Status: t.Status,
		Priority: t.Priority, AssignedTo: t.AssignedTo, Channel: t.Channel,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// -- LIST --------------------------------------------------------------------

func ListTicketsRoutes() []string { return []string{"/support/tickets"} }

type ListTicketsHandler struct{ Service *application.Service }

type ListTicketsRequest struct {
	Status   string `json:"status"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

type ListTicketsResponse struct {
	Tickets  []TicketPayload `json:"tickets"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

func (h *ListTicketsHandler) Serve(ctx context.Context, req *ListTicketsRequest) (ListTicketsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListTicketsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	items, total, err := h.Service.List(ctx, device.User.UserID, req.Status, req.Page, req.PageSize)
	if err != nil {
		return ListTicketsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list support tickets"}
	}
	payloads := make([]TicketPayload, len(items))
	for i, t := range items {
		payloads[i] = toTicketPayload(t)
	}
	return ListTicketsResponse{Tickets: payloads, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// -- GET DETAIL --------------------------------------------------------------

func GetTicketRoutes() []string { return []string{"/support/tickets/detail"} }

type GetTicketHandler struct{ Service *application.Service }

type GetTicketRequest struct {
	ID int64 `json:"id"`
}

type GetTicketResponse struct {
	Ticket   TicketPayload    `json:"ticket"`
	Messages []MessagePayload `json:"messages"`
}

func (h *GetTicketHandler) Validate(_ context.Context, req *GetTicketRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "ID is required"}
	}
	return nil
}

func (h *GetTicketHandler) Serve(ctx context.Context, req *GetTicketRequest) (GetTicketResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetTicketResponse{}, errorModels.PermissionDeniedServerError()
	}
	ticket, err := h.Service.GetByID(ctx, device.User.UserID, req.ID)
	if err != nil {
		return GetTicketResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get support ticket"}
	}
	msgs, err := h.Service.ListMessages(ctx, req.ID)
	if err != nil {
		return GetTicketResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list ticket messages"}
	}
	msgPayloads := make([]MessagePayload, len(msgs))
	for i, m := range msgs {
		msgPayloads[i] = MessagePayload{
			ID: m.ID, TicketID: m.TicketID, SenderType: m.SenderType,
			SenderName: m.SenderName, Body: m.Body, IsInternal: m.IsInternal,
			CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	return GetTicketResponse{Ticket: toTicketPayload(*ticket), Messages: msgPayloads}, nil
}

// -- CREATE ------------------------------------------------------------------

func CreateTicketRoutes() []string { return []string{"/support/tickets"} }

type CreateTicketHandler struct{ Service *application.Service }

type CreateTicketRequest struct {
	CustomerName  string `json:"customerName"`
	CustomerEmail string `json:"customerEmail"`
	Subject       string `json:"subject"`
	Priority      string `json:"priority"`
	Channel       string `json:"channel"`
	AssignedTo    *int64 `json:"assignedTo"`
	Body          string `json:"body"`
	SenderName    string `json:"senderName"`
}

type CreateTicketResponse struct {
	TicketID int64  `json:"ticketId"`
	Message  string `json:"message"`
}

func (h *CreateTicketHandler) Validate(_ context.Context, req *CreateTicketRequest) *errorModels.ServerError {
	if req.Subject == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Subject is required"}
	}
	if req.CustomerEmail == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Customer email is required"}
	}
	return nil
}

func (h *CreateTicketHandler) Serve(ctx context.Context, req *CreateTicketRequest) (CreateTicketResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateTicketResponse{}, errorModels.PermissionDeniedServerError()
	}
	ticketID, err := h.Service.Create(ctx, domain.TicketInput{
		BusinessID:    device.User.UserID,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		Subject:       req.Subject,
		Priority:      req.Priority,
		Channel:       req.Channel,
		AssignedTo:    req.AssignedTo,
	})
	if err != nil {
		return CreateTicketResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create support ticket"}
	}
	if req.Body != "" {
		senderName := req.SenderName
		if senderName == "" {
			senderName = req.CustomerName
		}
		_, _ = h.Service.CreateMessage(ctx, domain.MessageInput{
			TicketID:   ticketID,
			SenderType: "customer",
			SenderName: senderName,
			Body:       req.Body,
		})
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
	AssignedTo *int64 `json:"assignedTo"`
	Channel    string `json:"channel"`
}

type UpdateTicketResponse struct {
	Message string `json:"message"`
}

func (h *UpdateTicketHandler) Validate(_ context.Context, req *UpdateTicketRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "ID is required"}
	}
	return nil
}

func (h *UpdateTicketHandler) Serve(ctx context.Context, req *UpdateTicketRequest) (UpdateTicketResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateTicketResponse{}, errorModels.PermissionDeniedServerError()
	}
	err := h.Service.Update(ctx, domain.TicketUpdateInput{
		BusinessID: device.User.UserID,
		ID:         req.ID,
		Subject:    req.Subject,
		Status:     req.Status,
		Priority:   req.Priority,
		AssignedTo: req.AssignedTo,
		Channel:    req.Channel,
	})
	if err != nil {
		return UpdateTicketResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update support ticket"}
	}
	return UpdateTicketResponse{Message: "Support ticket updated"}, nil
}
