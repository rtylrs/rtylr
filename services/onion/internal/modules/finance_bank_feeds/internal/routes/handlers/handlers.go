package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/application"
	"github.com/voxire/onion/internal/modules/finance_bank_feeds/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ---------------------------------------------------------------------------
// List Bank Feed Transactions
// ---------------------------------------------------------------------------

func ListRoutes() []string { return []string{"/finance/bank-feeds"} }

type ListHandler struct{ Service *application.Service }

type ListRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type BankFeedTransactionPayload struct {
	ID               int64     `json:"id"`
	OwnerID          int64     `json:"owner_id"`
	BankAccountID    int64     `json:"bank_account_id"`
	AmountMinorUnits int64     `json:"amount_minor_units"`
	Currency         string    `json:"currency"`
	OccurredAt       time.Time `json:"occurred_at"`
	Description      string    `json:"description"`
	ExternalID       string    `json:"external_id"`
	MatchedExpenseID *int64    `json:"matched_expense_id"`
	MatchedInvoiceID *int64    `json:"matched_invoice_id"`
	CreatedAt        time.Time `json:"created_at"`
}

type ListResponse struct {
	Transactions []BankFeedTransactionPayload `json:"transactions"`
	Total        int64                        `json:"total"`
	Limit        int32                        `json:"limit"`
	Offset       int32                        `json:"offset"`
}

func toPayload(t domain.BankFeedTransaction) BankFeedTransactionPayload {
	return BankFeedTransactionPayload{
		ID:               t.ID,
		OwnerID:          t.OwnerID,
		BankAccountID:    t.BankAccountID,
		AmountMinorUnits: t.AmountMinorUnits,
		Currency:         t.Currency,
		OccurredAt:       t.OccurredAt,
		Description:      t.Description,
		ExternalID:       t.ExternalID,
		MatchedExpenseID: t.MatchedExpenseID,
		MatchedInvoiceID: t.MatchedInvoiceID,
		CreatedAt:        t.CreatedAt,
	}
}

func (h *ListHandler) Serve(ctx context.Context, req *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	items, total, err := h.Service.List(ctx, businessID, limit, offset)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list bank feed transactions"}
	}

	payloads := make([]BankFeedTransactionPayload, 0, len(items))
	for _, t := range items {
		payloads = append(payloads, toPayload(t))
	}
	return ListResponse{Transactions: payloads, Total: total, Limit: limit, Offset: offset}, nil
}

// ---------------------------------------------------------------------------
// Sync Bank Feeds
// ---------------------------------------------------------------------------

func SyncRoutes() []string { return []string{"/finance/bank-feeds/sync"} }

type (
	SyncHandler  struct{ Service *application.Service }
	SyncRequest  struct{}
	SyncResponse struct {
		Imported int `json:"imported"`
	}
)

func (h *SyncHandler) Serve(ctx context.Context, _ *SyncRequest) (SyncResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return SyncResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return SyncResponse{}, errorModels.PermissionDeniedServerError()
	}
	imported, err := h.Service.Sync(ctx, businessID)
	if err != nil {
		return SyncResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to sync bank feeds"}
	}
	return SyncResponse{Imported: imported}, nil
}
