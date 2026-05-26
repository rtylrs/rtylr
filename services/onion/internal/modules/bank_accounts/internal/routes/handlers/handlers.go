package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/application"
	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type BankAccountPayload struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	Name         string    `json:"name"`
	Institution  string    `json:"institution"`
	AccountType  string    `json:"account_type"`
	Balance      float64   `json:"balance"`
	Currency     string    `json:"currency"`
	IsConnected  bool      `json:"is_connected"`
	LastSyncedAt time.Time `json:"last_synced_at"`
	CreatedAt    time.Time `json:"created_at"`
}

func newPayload(b domain.BankAccount) BankAccountPayload {
	return BankAccountPayload{
		ID: b.ID, OwnerID: b.OwnerID, Name: b.Name, Institution: b.Institution,
		AccountType: b.AccountType, Balance: b.Balance, Currency: b.Currency,
		IsConnected: b.IsConnected, LastSyncedAt: b.LastSyncedAt, CreatedAt: b.CreatedAt,
	}
}

func ListRoutes() []string { return []string{"/finance/bank-accounts"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Accounts []BankAccountPayload `json:"accounts"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list bank accounts"}
	}
	payloads := make([]BankAccountPayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, newPayload(item))
	}
	return ListResponse{Accounts: payloads}, nil
}

func CreateRoutes() []string { return []string{"/finance/bank-accounts"} }

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		Name        string  `json:"name"`
		Institution string  `json:"institution"`
		AccountType string  `json:"account_type"`
		Balance     float64 `json:"balance"`
		Currency    string  `json:"currency"`
		IsConnected bool    `json:"is_connected"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.BankAccountInput{
		Name: req.Name, Institution: req.Institution, AccountType: req.AccountType,
		Balance: req.Balance, Currency: req.Currency, IsConnected: req.IsConnected,
	}
	id, err := h.Service.Create(ctx, device.User.UserID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create bank account"}
	}
	return CreateResponse{ID: id, Message: "Bank account created successfully"}, nil
}
