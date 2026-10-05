package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/application"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmloyaltyv1 "github.com/voxire/voxie/gen/go/crm/loyalty/v1"
)

func LoyaltyTransactionsRoutes() []string {
	return []string{"/crm/loyalty/transactions"}
}

type ListLoyaltyTransactionsHandler struct {
	CRMService *application.Service
}

type ListLoyaltyTransactionsRequest struct {
	CustomerID int64 `json:"customer_id"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
}

type ListLoyaltyTransactionsResponse struct {
	Transactions []*crmloyaltyv1.LoyaltyTransaction `json:"transactions"`
	Total        int                                `json:"total"`
	Page         int                                `json:"page"`
	PageSize     int                                `json:"page_size"`
}

func newLoyaltyTransactionPayload(tx domain.LoyaltyTransaction) *crmloyaltyv1.LoyaltyTransaction {
	return &crmloyaltyv1.LoyaltyTransaction{
		Id:            tx.ID,
		OwnerId:       tx.OwnerID,
		CustomerId:    tx.CustomerID,
		Type:          tx.Type,
		Points:        int64(tx.Points),
		BalanceBefore: int64(tx.BalanceBefore),
		BalanceAfter:  int64(tx.BalanceAfter),
		Description:   tx.Description,
		ReferenceId:   tx.ReferenceID,
		ReferenceType: tx.ReferenceType,
		CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *ListLoyaltyTransactionsHandler) Validate(_ context.Context, _ *ListLoyaltyTransactionsRequest) error {
	return nil
}

func (h *ListLoyaltyTransactionsHandler) Serve(ctx context.Context, req *ListLoyaltyTransactionsRequest) (ListLoyaltyTransactionsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListLoyaltyTransactionsResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListLoyaltyTransactionsResponse{}, onekiterrors.PermissionDenied()
	}
	if req == nil {
		return ListLoyaltyTransactionsResponse{}, onekiterrors.New(http.StatusBadRequest, "Loyalty transaction request is required")
	}

	page, pageSize, err := normalizeLoyaltyTransactionPage(int64(req.Page), int64(req.PageSize))
	if err != nil {
		return ListLoyaltyTransactionsResponse{}, err
	}
	req.Page = int(page)
	req.PageSize = int(pageSize)

	txs, total, err := h.CRMService.ListLoyaltyTransactions(ctx, businessID, req.CustomerID, req.Page, req.PageSize)
	if err != nil {
		return ListLoyaltyTransactionsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list loyalty transactions")
	}

	payloads := make([]*crmloyaltyv1.LoyaltyTransaction, len(txs))
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
