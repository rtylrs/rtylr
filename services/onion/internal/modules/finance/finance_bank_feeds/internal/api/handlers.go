package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/finance/finance_bank_feeds/internal/application"
	"github.com/voxire/onion/internal/modules/finance/finance_bank_feeds/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ---------------------------------------------------------------------------
// List Bank Feed Transactions
// ---------------------------------------------------------------------------

func ListRoutes() []string { return []string{"/finance/bank-feeds"} }

type ListHandler struct{ Service *application.Service }

type ListRequest struct {
	BankAccountID int64 `json:"bank_account_id"`
	AccountID     int64 `json:"account_id"`
	Limit         int32 `json:"limit"`
	Offset        int32 `json:"offset"`
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

func toViewerPayload(t domain.BankFeedTransaction, includeSensitive bool) BankFeedTransactionPayload {
	payload := toPayload(t)
	if !includeSensitive {
		payload.BankAccountID = 0
		payload.ExternalID = ""
		payload.MatchedExpenseID = nil
		payload.MatchedInvoiceID = nil
	}
	return payload
}

func (h *ListHandler) Serve(ctx context.Context, req *ListRequest) (ListResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(device.User, authorization.AppFinance, "finance:bank_feeds:view") {
		return ListResponse{}, onekiterrors.PermissionDenied()
	}
	includeSensitive := authPkg.HasAppPermission(device.User, authorization.AppFinance, "finance:bank_feeds:sensitive_view")

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListResponse{}, onekiterrors.PermissionDenied()
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

	bankAccountID := req.BankAccountID
	if bankAccountID != 0 && req.AccountID != 0 && bankAccountID != req.AccountID {
		return ListResponse{}, onekiterrors.New(http.StatusBadRequest, "account_id and bank_account_id must match when both are provided")
	}
	if bankAccountID == 0 {
		bankAccountID = req.AccountID
	}
	if bankAccountID < 0 {
		return ListResponse{}, onekiterrors.New(http.StatusBadRequest, "bank_account_id must be positive")
	}

	var items []domain.BankFeedTransaction
	var total int64
	var err error
	if bankAccountID > 0 {
		items, total, err = h.Service.ListByAccount(ctx, businessID, bankAccountID, limit, offset)
	} else {
		items, total, err = h.Service.List(ctx, businessID, limit, offset)
	}
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInvalidBankAccount):
			return ListResponse{}, onekiterrors.New(http.StatusBadRequest, "bank_account_id must be positive")
		case errors.Is(err, application.ErrBankAccountNotFound):
			return ListResponse{}, onekiterrors.New(http.StatusNotFound, "Bank account not found")
		default:
			return ListResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list bank feed transactions")
		}
	}

	payloads := make([]BankFeedTransactionPayload, 0, len(items))
	for _, t := range items {
		payloads = append(payloads, toViewerPayload(t, includeSensitive))
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

const (
	// A feed sync is a state-changing bank operation, not a read of the feed.
	// A view-only grant must never satisfy this check.
	bankFeedSyncPermission           = "finance:bank_feeds:sync"
	bankFeedReconciliationPermission = "finance:reconciliation:manage"
)

func canSyncBankFeeds(user *authPkg.AuthUser) bool {
	if user == nil || !authPkg.HasAppAccess(user, authorization.AppFinance) {
		return false
	}
	return authPkg.HasAppPermission(user, authorization.AppFinance, bankFeedSyncPermission) ||
		authPkg.HasAppPermission(user, authorization.AppFinance, bankFeedReconciliationPermission)
}

func (h *SyncHandler) Serve(ctx context.Context, _ *SyncRequest) (SyncResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return SyncResponse{}, onekiterrors.PermissionDenied()
	}
	if !canSyncBankFeeds(device.User) {
		return SyncResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return SyncResponse{}, onekiterrors.PermissionDenied()
	}
	imported, err := h.Service.Sync(ctx, businessID)
	if err != nil {
		return SyncResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to sync bank feeds")
	}
	return SyncResponse{Imported: imported}, nil
}
