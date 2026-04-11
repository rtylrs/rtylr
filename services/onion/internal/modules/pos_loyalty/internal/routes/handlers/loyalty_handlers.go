package handlers

import (
	"context"
	"errors"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/application"
	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/domain"
)

// ─── Get Config ──────────────────────────────────────────────────────────────

func GetConfigRoutes() []string { return []string{"/pos/loyalty/config"} }

type GetConfigHandler struct{ Service *application.Service }
type GetConfigRequest struct{}
type GetConfigResponse struct {
	Config *LoyaltyConfigPayload `json:"config"`
}

type LoyaltyConfigPayload struct {
	ID              int64   `json:"id"`
	BusinessID      int64   `json:"businessId"`
	IsEnabled       bool    `json:"isEnabled"`
	PointsPerDollar float64 `json:"pointsPerDollar"`
	RedemptionRate  float64 `json:"redemptionRate"`
	MinRedeemPoints int32   `json:"minRedeemPoints"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

func (h *GetConfigHandler) Serve(ctx context.Context, _ *GetConfigRequest) (GetConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetConfigResponse{}, errorModels.PermissionDeniedServerError()
	}
	cfg, err := h.Service.GetConfig(ctx, device.User.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			return GetConfigResponse{Config: nil}, nil
		}
		return GetConfigResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get loyalty config"}
	}
	payload := &LoyaltyConfigPayload{
		ID:              cfg.ID,
		BusinessID:      cfg.BusinessID,
		IsEnabled:       cfg.IsEnabled,
		PointsPerDollar: cfg.PointsPerDollar,
		RedemptionRate:  cfg.RedemptionRate,
		MinRedeemPoints: cfg.MinRedeemPoints,
		CreatedAt:       cfg.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       cfg.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	return GetConfigResponse{Config: payload}, nil
}

// ─── Update Config ───────────────────────────────────────────────────────────

func UpdateConfigRoutes() []string { return []string{"/pos/loyalty/config"} }

type UpdateConfigHandler struct{ Service *application.Service }

type UpdateConfigRequest struct {
	IsEnabled       bool    `json:"isEnabled"`
	PointsPerDollar float64 `json:"pointsPerDollar"`
	RedemptionRate  float64 `json:"redemptionRate"`
	MinRedeemPoints int32   `json:"minRedeemPoints"`
}

type UpdateConfigResponse struct {
	Message string `json:"message"`
}

func (h *UpdateConfigHandler) Serve(ctx context.Context, req *UpdateConfigRequest) (UpdateConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateConfigResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.UpdateConfig(ctx, domain.UpdateLoyaltyConfigInput{
		BusinessID:      device.User.UserID,
		IsEnabled:       req.IsEnabled,
		PointsPerDollar: req.PointsPerDollar,
		RedemptionRate:  req.RedemptionRate,
		MinRedeemPoints: req.MinRedeemPoints,
	}); err != nil {
		return UpdateConfigResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update loyalty config"}
	}
	return UpdateConfigResponse{Message: "Loyalty config updated successfully"}, nil
}

// ─── List Transactions ───────────────────────────────────────────────────────

func ListTransactionsRoutes() []string { return []string{"/pos/loyalty/transactions"} }

type ListTransactionsHandler struct{ Service *application.Service }
type ListTransactionsRequest struct {
	CustomerID int64 `json:"customer_id"`
}
type ListTransactionsResponse struct {
	Transactions []LoyaltyTransactionPayload `json:"transactions"`
}

type LoyaltyTransactionPayload struct {
	ID           int64   `json:"id"`
	BusinessID   int64   `json:"businessId"`
	CustomerID   int64   `json:"customerId"`
	OrderID      *int64  `json:"orderId"`
	Type         string  `json:"type"`
	Points       int32   `json:"points"`
	BalanceAfter int32   `json:"balanceAfter"`
	Note         *string `json:"note"`
	CreatedAt    string  `json:"createdAt"`
}

func (h *ListTransactionsHandler) Serve(ctx context.Context, req *ListTransactionsRequest) (ListTransactionsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListTransactionsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.CustomerID == 0 {
		return ListTransactionsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "customer_id is required"}
	}
	txns, err := h.Service.ListTransactions(ctx, device.User.UserID, req.CustomerID)
	if err != nil {
		return ListTransactionsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list transactions"}
	}
	payloads := make([]LoyaltyTransactionPayload, len(txns))
	for i, t := range txns {
		payloads[i] = LoyaltyTransactionPayload{
			ID:           t.ID,
			BusinessID:   t.BusinessID,
			CustomerID:   t.CustomerID,
			OrderID:      t.OrderID,
			Type:         t.Type,
			Points:       t.Points,
			BalanceAfter: t.BalanceAfter,
			Note:         t.Note,
			CreatedAt:    t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	return ListTransactionsResponse{Transactions: payloads}, nil
}

// ─── Earn Points ─────────────────────────────────────────────────────────────

func EarnPointsRoutes() []string { return []string{"/pos/loyalty/earn"} }

type EarnPointsHandler struct{ Service *application.Service }

type EarnPointsRequest struct {
	CustomerID int64  `json:"customerId"`
	OrderID    *int64 `json:"orderId"`
	Points     int32  `json:"points"`
}

type EarnPointsResponse struct {
	TransactionID int64  `json:"transactionId"`
	Message       string `json:"message"`
}

func (h *EarnPointsHandler) Serve(ctx context.Context, req *EarnPointsRequest) (EarnPointsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return EarnPointsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.CustomerID == 0 || req.Points <= 0 {
		return EarnPointsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "customerId and positive points are required"}
	}
	id, err := h.Service.EarnPoints(ctx, domain.EarnPointsInput{
		BusinessID: device.User.UserID,
		CustomerID: req.CustomerID,
		OrderID:    req.OrderID,
		Points:     req.Points,
	})
	if err != nil {
		return EarnPointsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to earn points"}
	}
	return EarnPointsResponse{TransactionID: id, Message: "Points earned successfully"}, nil
}

// ─── Redeem Points ───────────────────────────────────────────────────────────

func RedeemPointsRoutes() []string { return []string{"/pos/loyalty/redeem"} }

type RedeemPointsHandler struct{ Service *application.Service }

type RedeemPointsRequest struct {
	CustomerID int64 `json:"customerId"`
	Points     int32 `json:"points"`
}

type RedeemPointsResponse struct {
	TransactionID int64  `json:"transactionId"`
	Message       string `json:"message"`
}

func (h *RedeemPointsHandler) Serve(ctx context.Context, req *RedeemPointsRequest) (RedeemPointsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return RedeemPointsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.CustomerID == 0 || req.Points <= 0 {
		return RedeemPointsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "customerId and positive points are required"}
	}
	id, err := h.Service.RedeemPoints(ctx, domain.RedeemPointsInput{
		BusinessID: device.User.UserID,
		CustomerID: req.CustomerID,
		Points:     req.Points,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientPoints) {
			return RedeemPointsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Insufficient loyalty points"}
		}
		return RedeemPointsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to redeem points"}
	}
	return RedeemPointsResponse{TransactionID: id, Message: "Points redeemed successfully"}, nil
}

// ─── Get Balance ─────────────────────────────────────────────────────────────

func GetBalanceRoutes() []string { return []string{"/pos/loyalty/balance"} }

type GetBalanceHandler struct{ Service *application.Service }
type GetBalanceRequest struct {
	CustomerID int64 `json:"customer_id"`
}
type GetBalanceResponse struct {
	CustomerID int64 `json:"customerId"`
	Balance    int32 `json:"balance"`
}

func (h *GetBalanceHandler) Serve(ctx context.Context, req *GetBalanceRequest) (GetBalanceResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetBalanceResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.CustomerID == 0 {
		return GetBalanceResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "customer_id is required"}
	}
	balance, err := h.Service.GetBalance(ctx, device.User.UserID, req.CustomerID)
	if err != nil {
		return GetBalanceResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get balance"}
	}
	return GetBalanceResponse{CustomerID: req.CustomerID, Balance: balance}, nil
}
