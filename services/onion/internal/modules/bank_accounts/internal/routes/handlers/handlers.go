package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/application"
	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// bankAccountsCacheScope namespaces the per-owner bank accounts list cache.
// finance_bank_accounts is written only within this module (Create), so a
// short TTL plus invalidation on create keeps reads consistent.
const bankAccountsCacheScope = "bank_accounts"

func invalidateBankAccountsCache(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(bankAccountsCacheScope, businessID))
	}
}

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
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	res, cacheErr := cacher.Of[ListResponse]().GetOrComputeSingle(
		ctx,
		cacheKeys.Scoped(bankAccountsCacheScope, businessID, "list"),
		func(ctx context.Context) (ListResponse, error) {
			return h.computeList(ctx, businessID)
		},
		cacher.WithExpiry(ctx, 3*time.Minute),
	)
	if cacheErr == nil {
		return res, nil
	}

	res, computeErr := h.computeList(ctx, businessID)
	if computeErr != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list bank accounts"}
	}
	return res, nil
}

func (h *ListHandler) computeList(ctx context.Context, businessID int64) (ListResponse, error) {
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, err
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
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.BankAccountInput{
		Name: req.Name, Institution: req.Institution, AccountType: req.AccountType,
		Balance: req.Balance, Currency: req.Currency, IsConnected: req.IsConnected,
	}
	id, err := h.Service.Create(ctx, businessID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create bank account"}
	}
	invalidateBankAccountsCache(ctx, businessID)
	return CreateResponse{ID: id, Message: "Bank account created successfully"}, nil
}
