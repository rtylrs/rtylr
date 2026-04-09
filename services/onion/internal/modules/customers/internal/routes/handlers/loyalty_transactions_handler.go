package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
)

func LoyaltyTransactionsRoutes() []string {
	return []string{"/crm/loyalty/transactions"}
}

type ListLoyaltyTransactionsHandler struct {
	CRMService *application.Service
}

type ListLoyaltyTransactionsRequest struct {
	CustomerID int64 `json:"customerId"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
}

type LoyaltyTransactionPayload struct {
	ID            int64   `json:"id"`
	OwnerID       int64   `json:"ownerId"`
	CustomerID    int64   `json:"customerId"`
	Type          string  `json:"type"`
	Points        int     `json:"points"`
	BalanceBefore int     `json:"balanceBefore"`
	BalanceAfter  int     `json:"balanceAfter"`
	Description   string  `json:"description"`
	ReferenceID   *int64  `json:"referenceId"`
	ReferenceType string  `json:"referenceType"`
	CreatedAt     string  `json:"createdAt"`
}

type ListLoyaltyTransactionsResponse struct {
	Transactions []LoyaltyTransactionPayload `json:"transactions"`
	Total        int                         `json:"total"`
	Page         int                         `json:"page"`
	PageSize     int                         `json:"pageSize"`
}

func newLoyaltyTransactionPayload(tx domain.LoyaltyTransaction) LoyaltyTransactionPayload {
	return LoyaltyTransactionPayload{
		ID:            tx.ID,
		OwnerID:       tx.OwnerID,
		CustomerID:    tx.CustomerID,
		Type:          tx.Type,
		Points:        tx.Points,
		BalanceBefore: tx.BalanceBefore,
		BalanceAfter:  tx.BalanceAfter,
		Description:   tx.Description,
		ReferenceID:   tx.ReferenceID,
		ReferenceType: tx.ReferenceType,
		CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *ListLoyaltyTransactionsHandler) Validate(_ context.Context, req *ListLoyaltyTransactionsRequest) *errorModels.ServerError {
	if req.CustomerID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Customer ID is required",
		}
	}

	return nil
}

func (h *ListLoyaltyTransactionsHandler) Serve(ctx context.Context, req *ListLoyaltyTransactionsRequest) (ListLoyaltyTransactionsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListLoyaltyTransactionsResponse{}, errorModels.PermissionDeniedServerError()
	}

	if req.Page < 1 {
		req.Page = 1
	}

	if req.PageSize < 1 {
		req.PageSize = 20
	}

	txs, total, err := h.CRMService.ListLoyaltyTransactions(ctx, device.User.UserID, req.CustomerID, req.Page, req.PageSize)
	if err != nil {
		return ListLoyaltyTransactionsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list loyalty transactions",
		}
	}

	payloads := make([]LoyaltyTransactionPayload, len(txs))
	for i, tx := range txs {
		payloads[i] = newLoyaltyTransactionPayload(tx)
	}

	return ListLoyaltyTransactionsResponse{
		Transactions: payloads,
		Total:        total,
		Page:         req.Page,
		PageSize:     req.PageSize,
	}, nil
}
