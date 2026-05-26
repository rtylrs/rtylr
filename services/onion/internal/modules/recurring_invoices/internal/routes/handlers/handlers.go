package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/application"
	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ---- Payload ----

type RecurringInvoicePayload struct {
	ID                int64     `json:"id"`
	OwnerID           int64     `json:"owner_id"`
	TemplateInvoiceID int64     `json:"template_invoice_id"`
	Frequency         string    `json:"frequency"`
	NextInvoiceDate   time.Time `json:"next_invoice_date"`
	EndDate           time.Time `json:"end_date"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
}

func newPayload(r domain.RecurringInvoice) RecurringInvoicePayload {
	return RecurringInvoicePayload{
		ID: r.ID, OwnerID: r.OwnerID, TemplateInvoiceID: r.TemplateInvoiceID,
		Frequency: r.Frequency, NextInvoiceDate: r.NextInvoiceDate,
		EndDate: r.EndDate, IsActive: r.IsActive, CreatedAt: r.CreatedAt,
	}
}

// ---- List ----

func ListRoutes() []string { return []string{"/finance/recurring-invoices"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		RecurringInvoices []RecurringInvoicePayload `json:"recurring_invoices"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}

	items, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list recurring invoices"}
	}

	payloads := make([]RecurringInvoicePayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, newPayload(item))
	}

	return ListResponse{RecurringInvoices: payloads}, nil
}

// ---- Create ----

func CreateRoutes() []string { return []string{"/finance/recurring-invoices"} }

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		TemplateInvoiceID int64  `json:"template_invoice_id"`
		Frequency         string `json:"frequency"`
		NextInvoiceDate   string `json:"next_invoice_date"`
		EndDate           string `json:"end_date"`
		IsActive          bool   `json:"is_active"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.TemplateInvoiceID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "templateInvoiceId is required"}
	}
	if req.Frequency == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "frequency is required"}
	}
	if req.NextInvoiceDate == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "nextInvoiceDate is required"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.RecurringInvoiceInput{
		TemplateInvoiceID: req.TemplateInvoiceID,
		Frequency:         req.Frequency,
		NextInvoiceDate:   req.NextInvoiceDate,
		EndDate:           req.EndDate,
		IsActive:          req.IsActive,
	}

	id, err := h.Service.Create(ctx, device.User.UserID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create recurring invoice"}
	}

	return CreateResponse{ID: id, Message: "Recurring invoice created successfully"}, nil
}

// ---- Update ----

func UpdateRoutes() []string { return []string{"/finance/recurring-invoices"} }

type (
	UpdateHandler struct{ Service *application.Service }
	UpdateRequest struct {
		ID                int64  `json:"id"`
		TemplateInvoiceID int64  `json:"template_invoice_id"`
		Frequency         string `json:"frequency"`
		NextInvoiceDate   string `json:"next_invoice_date"`
		EndDate           string `json:"end_date"`
		IsActive          bool   `json:"is_active"`
	}
)

type UpdateResponse struct {
	Message string `json:"message"`
}

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	if req.Frequency == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "frequency is required"}
	}
	if req.NextInvoiceDate == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "nextInvoiceDate is required"}
	}
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.RecurringInvoiceInput{
		TemplateInvoiceID: req.TemplateInvoiceID,
		Frequency:         req.Frequency,
		NextInvoiceDate:   req.NextInvoiceDate,
		EndDate:           req.EndDate,
		IsActive:          req.IsActive,
	}

	if err := h.Service.Update(ctx, device.User.UserID, req.ID, input); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update recurring invoice"}
	}

	return UpdateResponse{Message: "Recurring invoice updated successfully"}, nil
}
