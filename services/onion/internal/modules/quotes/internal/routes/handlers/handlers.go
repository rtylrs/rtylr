package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/quotes/internal/application"
	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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

func ListRoutes() []string { return []string{"/finance/quotes"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Quotes []QuotePayload `json:"quotes"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}

	items, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list quotes"}
	}

	payloads := make([]QuotePayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, newPayload(item))
	}

	return ListResponse{Quotes: payloads}, nil
}

// ---- Create ----

func CreateRoutes() []string { return []string{"/finance/quotes"} }

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

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.CustomerName == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "customerName is required"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.QuoteInput{
		CustomerName: req.CustomerName, CustomerEmail: req.CustomerEmail,
		Status: req.Status, Items: req.Items, Subtotal: req.Subtotal,
		Tax: req.Tax, Discount: req.Discount, Total: req.Total,
		ValidUntil: req.ValidUntil, Notes: req.Notes,
	}

	id, err := h.Service.Create(ctx, device.User.UserID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create quote"}
	}

	return CreateResponse{ID: id, Message: "Quote created successfully"}, nil
}

// ---- Update ----

func UpdateRoutes() []string { return []string{"/finance/quotes"} }

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

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.QuoteInput{
		CustomerName: req.CustomerName, CustomerEmail: req.CustomerEmail,
		Status: req.Status, Items: req.Items, Subtotal: req.Subtotal,
		Tax: req.Tax, Discount: req.Discount, Total: req.Total,
		ValidUntil: req.ValidUntil, Notes: req.Notes,
	}

	if err := h.Service.Update(ctx, device.User.UserID, req.ID, input); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update quote"}
	}

	return UpdateResponse{Message: "Quote updated successfully"}, nil
}

// ---- Delete ----

func DeleteRoutes() []string { return []string{"/finance/quotes"} }

type (
	DeleteHandler struct{ Service *application.Service }
	DeleteRequest struct {
		ID int64 `json:"id"`
	}
)

type DeleteResponse struct {
	Message string `json:"message"`
}

func (h *DeleteHandler) Validate(_ context.Context, req *DeleteRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	return nil
}

func (h *DeleteHandler) Serve(ctx context.Context, req *DeleteRequest) (DeleteResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.Delete(ctx, device.User.UserID, req.ID); err != nil {
		return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete quote"}
	}

	return DeleteResponse{Message: "Quote deleted successfully"}, nil
}
