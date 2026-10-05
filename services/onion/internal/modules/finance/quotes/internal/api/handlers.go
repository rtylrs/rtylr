package api

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/finance/quotes/internal/application"
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

type QuotePayload struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	CustomerName  string    `json:"customer_name"`
	CustomerEmail string    `json:"customer_email"`
	Status        string    `json:"status"`
	Items         string    `json:"items"`
	Subtotal      float64   `json:"subtotal"`
	Tax           float64   `json:"tax"`
	Discount      float64   `json:"discount"`
	Total         float64   `json:"total"`
	ValidUntil    time.Time `json:"valid_until"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func newPayload(q domain.Quote) QuotePayload {
	return QuotePayload{
		ID: q.ID, OwnerID: q.OwnerID, CustomerName: q.CustomerName,
		CustomerEmail: q.CustomerEmail, Status: q.Status, Items: q.Items,
		Subtotal: q.Subtotal, Tax: q.Tax, Discount: q.Discount, Total: q.Total,
		ValidUntil: q.ValidUntil, Notes: q.Notes, CreatedAt: q.CreatedAt, UpdatedAt: q.UpdatedAt,
	}
}

// ---- List ----

type (
	ListHandler struct{ Service *application.Service }
	ListRequest struct {
		Page     int64 `json:"page"`
		PageSize int64 `json:"page_size"`
	}
	ListResponse struct {
		Quotes   []QuotePayload `json:"quotes"`
		Page     int64          `json:"page"`
		PageSize int64          `json:"page_size"`
		HasMore  bool           `json:"has_more"`
	}
)

// ---- Create ----

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		CustomerName  string  `json:"customer_name"`
		CustomerEmail string  `json:"customer_email"`
		Status        string  `json:"status"`
		Items         string  `json:"items"`
		Subtotal      float64 `json:"subtotal"`
		Tax           float64 `json:"tax"`
		Discount      float64 `json:"discount"`
		Total         float64 `json:"total"`
		ValidUntil    string  `json:"valid_until"`
		Notes         string  `json:"notes"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) error {
	err := domain.ValidateQuoteInput(domain.QuoteInput{
		CustomerName: req.CustomerName,
		Status:       req.Status,
		Subtotal:     req.Subtotal,
		Tax:          req.Tax,
		Discount:     req.Discount,
		Total:        req.Total,
		ValidUntil:   req.ValidUntil,
	})
	if err != nil {
		return onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return nil
}

// ---- Update ----

type (
	UpdateHandler struct{ Service *application.Service }
	UpdateRequest struct {
		ID            int64   `json:"id"`
		CustomerName  string  `json:"customer_name"`
		CustomerEmail string  `json:"customer_email"`
		Status        string  `json:"status"`
		Items         string  `json:"items"`
		Subtotal      float64 `json:"subtotal"`
		Tax           float64 `json:"tax"`
		Discount      float64 `json:"discount"`
		Total         float64 `json:"total"`
		ValidUntil    string  `json:"valid_until"`
		Notes         string  `json:"notes"`
	}
)

type UpdateResponse struct {
	Message string `json:"message"`
}

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	err := domain.ValidateQuoteInput(domain.QuoteInput{
		CustomerName: req.CustomerName,
		Status:       req.Status,
		Subtotal:     req.Subtotal,
		Tax:          req.Tax,
		Discount:     req.Discount,
		Total:        req.Total,
		ValidUntil:   req.ValidUntil,
	})
	if err != nil {
		return onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return nil
}

// ---- Delete ----

type (
	DeleteHandler struct{ Service *application.Service }
	DeleteRequest struct {
		ID int64 `json:"id"`
	}
)

type DeleteResponse struct {
	Message string `json:"message"`
}
