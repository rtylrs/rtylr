package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/loyalty/internal/application"
	"github.com/voxire/onion/internal/modules/loyalty/internal/domain"
)

// ─── List Transactions ────────────────────────────────────────────────────────

func LoyaltyTransactionsRoutes() []string { return []string{"/crm/loyalty/transactions"} }

type ListTransactionsHandler struct{ Service *application.Service }

type ListTransactionsRequest struct {
	CustomerID int64 `json:"customerId"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
}

type TransactionPayload struct {
	ID            int64  `json:"id"`
	OwnerID       int64  `json:"ownerId"`
	CustomerID    int64  `json:"customerId"`
	Points        int    `json:"points"`
	Type          string `json:"type"`
	Description   string `json:"description"`
	ReferenceID   *int64 `json:"referenceId"`
	ReferenceType string `json:"referenceType"`
	CreatedAt     string `json:"createdAt"`
}

type ListTransactionsResponse struct {
	Transactions []TransactionPayload `json:"transactions"`
	Total        int                  `json:"total"`
	Page         int                  `json:"page"`
	PageSize     int                  `json:"pageSize"`
}

func (h *ListTransactionsHandler) Serve(ctx context.Context, req *ListTransactionsRequest) (ListTransactionsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListTransactionsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	txs, total, err := h.Service.ListTransactions(ctx, device.User.UserID, req.CustomerID, req.Page, req.PageSize)
	if err != nil {
		return ListTransactionsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list loyalty transactions"}
	}
	payloads := make([]TransactionPayload, len(txs))
	for i, tx := range txs {
		payloads[i] = TransactionPayload{
			ID:            tx.ID,
			OwnerID:       tx.OwnerID,
			CustomerID:    tx.CustomerID,
			Points:        tx.Points,
			Type:          tx.Type,
			Description:   tx.Description,
			ReferenceID:   tx.ReferenceID,
			ReferenceType: tx.ReferenceType,
			CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	return ListTransactionsResponse{Transactions: payloads, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// ─── Award Points ─────────────────────────────────────────────────────────────

func LoyaltyAwardRoutes() []string { return []string{"/crm/loyalty/award"} }

type AwardPointsHandler struct{ Service *application.Service }

type AwardPointsRequest struct {
	CustomerID    int64  `json:"customerId"`
	Points        int    `json:"points"`
	Description   string `json:"description"`
	ReferenceID   *int64 `json:"referenceId"`
	ReferenceType string `json:"referenceType"`
}

type AwardPointsResponse struct {
	Message string `json:"message"`
}

func (h *AwardPointsHandler) Validate(_ context.Context, req *AwardPointsRequest) *errorModels.ServerError {
	if req.CustomerID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Customer ID is required"}
	}
	if req.Points <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Points must be greater than zero"}
	}
	return nil
}

func (h *AwardPointsHandler) Serve(ctx context.Context, req *AwardPointsRequest) (AwardPointsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return AwardPointsResponse{}, errorModels.PermissionDeniedServerError()
	}
	err := h.Service.AwardPoints(ctx, domain.AwardInput{
		OwnerID:       device.User.UserID,
		CustomerID:    req.CustomerID,
		Points:        req.Points,
		Description:   req.Description,
		ReferenceID:   req.ReferenceID,
		ReferenceType: req.ReferenceType,
	})
	if err != nil {
		return AwardPointsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to award points"}
	}
	return AwardPointsResponse{Message: "Points awarded successfully"}, nil
}

// ─── Redeem Points ────────────────────────────────────────────────────────────

func LoyaltyRedeemRoutes() []string { return []string{"/crm/loyalty/redeem"} }

type RedeemPointsHandler struct{ Service *application.Service }

type RedeemPointsRequest struct {
	CustomerID    int64  `json:"customerId"`
	Points        int    `json:"points"`
	Description   string `json:"description"`
	ReferenceID   *int64 `json:"referenceId"`
	ReferenceType string `json:"referenceType"`
}

type RedeemPointsResponse struct {
	Message string `json:"message"`
}

func (h *RedeemPointsHandler) Validate(_ context.Context, req *RedeemPointsRequest) *errorModels.ServerError {
	if req.CustomerID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Customer ID is required"}
	}
	if req.Points <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Points must be greater than zero"}
	}
	return nil
}

func (h *RedeemPointsHandler) Serve(ctx context.Context, req *RedeemPointsRequest) (RedeemPointsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return RedeemPointsResponse{}, errorModels.PermissionDeniedServerError()
	}
	err := h.Service.RedeemPoints(ctx, domain.RedeemInput{
		OwnerID:       device.User.UserID,
		CustomerID:    req.CustomerID,
		Points:        req.Points,
		Description:   req.Description,
		ReferenceID:   req.ReferenceID,
		ReferenceType: req.ReferenceType,
	})
	if err != nil {
		return RedeemPointsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to redeem points"}
	}
	return RedeemPointsResponse{Message: "Points redeemed successfully"}, nil
}

// ─── Get Config ───────────────────────────────────────────────────────────────

func LoyaltyConfigGetRoutes() []string { return []string{"/crm/loyalty/config"} }

type GetLoyaltyConfigHandler struct{ Service *application.Service }

type GetLoyaltyConfigRequest struct{}

type LoyaltyConfigPayload struct {
	ID                  int64   `json:"id"`
	OwnerID             int64   `json:"ownerId"`
	PointsPerDollar     float64 `json:"pointsPerDollar"`
	MinRedemptionPoints int     `json:"minRedemptionPoints"`
	PointsValue         float64 `json:"pointsValue"`
	IsActive            bool    `json:"isActive"`
	CreatedAt           string  `json:"createdAt"`
	UpdatedAt           string  `json:"updatedAt"`
}

type GetLoyaltyConfigResponse struct {
	Config LoyaltyConfigPayload `json:"config"`
}

func (h *GetLoyaltyConfigHandler) Serve(ctx context.Context, _ *GetLoyaltyConfigRequest) (GetLoyaltyConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetLoyaltyConfigResponse{}, errorModels.PermissionDeniedServerError()
	}
	cfg, err := h.Service.GetConfig(ctx, device.User.UserID)
	if err != nil {
		return GetLoyaltyConfigResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get loyalty config"}
	}
	return GetLoyaltyConfigResponse{Config: LoyaltyConfigPayload{
		ID:                  cfg.ID,
		OwnerID:             cfg.OwnerID,
		PointsPerDollar:     cfg.PointsPerDollar,
		MinRedemptionPoints: cfg.MinRedemptionPoints,
		PointsValue:         cfg.PointsValue,
		IsActive:            cfg.IsActive,
		CreatedAt:           cfg.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:           cfg.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}}, nil
}

// ─── Update Config ────────────────────────────────────────────────────────────

func LoyaltyConfigUpdateRoutes() []string { return []string{"/crm/loyalty/config"} }

type UpdateLoyaltyConfigHandler struct{ Service *application.Service }

type UpdateLoyaltyConfigRequest struct {
	PointsPerDollar     float64 `json:"pointsPerDollar"`
	MinRedemptionPoints int     `json:"minRedemptionPoints"`
	PointsValue         float64 `json:"pointsValue"`
	IsActive            bool    `json:"isActive"`
}

type UpdateLoyaltyConfigResponse struct {
	Message string `json:"message"`
}

func (h *UpdateLoyaltyConfigHandler) Serve(ctx context.Context, req *UpdateLoyaltyConfigRequest) (UpdateLoyaltyConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateLoyaltyConfigResponse{}, errorModels.PermissionDeniedServerError()
	}
	err := h.Service.UpdateConfig(ctx, domain.LoyaltyConfigInput{
		OwnerID:             device.User.UserID,
		PointsPerDollar:     req.PointsPerDollar,
		MinRedemptionPoints: req.MinRedemptionPoints,
		PointsValue:         req.PointsValue,
		IsActive:            req.IsActive,
	})
	if err != nil {
		return UpdateLoyaltyConfigResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update loyalty config"}
	}
	return UpdateLoyaltyConfigResponse{Message: "Loyalty config updated successfully"}, nil
}
