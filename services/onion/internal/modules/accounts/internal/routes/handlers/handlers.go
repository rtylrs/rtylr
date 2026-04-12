package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/accounts/internal/application"
	"github.com/voxire/onion/internal/modules/accounts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type AccountPayload struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"owner_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Currency  string    `json:"currency"`
	Balance   float64   `json:"balance"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func newPayload(a domain.Account) AccountPayload {
	return AccountPayload{
		ID: a.ID, OwnerID: a.OwnerID, Name: a.Name, Type: a.Type,
		Currency: a.Currency, Balance: a.Balance, IsActive: a.IsActive, CreatedAt: a.CreatedAt,
	}
}

func ListRoutes() []string { return []string{"/finance/accounts"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}
type ListResponse struct {
	Accounts []AccountPayload `json:"accounts"`
}

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list accounts"}
	}
	payloads := make([]AccountPayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, newPayload(item))
	}
	return ListResponse{Accounts: payloads}, nil
}

func CreateRoutes() []string { return []string{"/finance/accounts"} }

type CreateHandler struct{ Service *application.Service }
type CreateRequest struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	IsActive bool    `json:"is_active"`
}
type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	if req.Type == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "type is required"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.AccountInput{
		Name: req.Name, Type: req.Type, Currency: req.Currency,
		Balance: req.Balance, IsActive: req.IsActive,
	}
	id, err := h.Service.Create(ctx, device.User.UserID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create account"}
	}
	return CreateResponse{ID: id, Message: "Account created successfully"}, nil
}
