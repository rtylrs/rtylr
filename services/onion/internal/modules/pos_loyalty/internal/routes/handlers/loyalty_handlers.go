package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/application"
	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	posloyaltyv1 "github.com/voxire/voxie/gen/go/pos/loyalty/v1"
)

// ─── Get Config ──────────────────────────────────────────────────────────────

func GetConfigRoutes() []string { return []string{"/pos/loyalty/config"} }

type (
	GetConfigHandler  struct{ Service *application.Service }
	GetConfigRequest  struct{}
	GetConfigResponse struct {
		Config *posloyaltyv1.LoyaltyConfig `json:"config"`
	}
)

func (h *GetConfigHandler) Serve(ctx context.Context, _ *GetConfigRequest) (GetConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetConfigResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetConfigResponse{}, errorModels.PermissionDeniedServerError()
	}
	cfg, err := h.Service.GetConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			return GetConfigResponse{Config: nil}, nil
		}
		return GetConfigResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get loyalty config"}
	}
	return GetConfigResponse{Config: &posloyaltyv1.LoyaltyConfig{
		Id:              cfg.ID,
		BusinessId:      cfg.BusinessID,
		IsEnabled:       cfg.IsEnabled,
		PointsPerDollar: cfg.PointsPerDollar,
		RedemptionRate:  cfg.RedemptionRate,
		MinRedeemPoints: cfg.MinRedeemPoints,
		CreatedAt:       cfg.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       cfg.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}}, nil
}

// ─── Update Config ───────────────────────────────────────────────────────────

func UpdateConfigRoutes() []string { return []string{"/pos/loyalty/config"} }

type UpdateConfigHandler struct{ Service *application.Service }

type UpdateConfigRequest struct {
	IsEnabled       bool    `json:"is_enabled"`
	PointsPerDollar float64 `json:"points_per_dollar"`
	RedemptionRate  float64 `json:"redemption_rate"`
	MinRedeemPoints int32   `json:"min_redeem_points"`
}

type UpdateConfigResponse struct {
	Message string `json:"message"`
}

func (h *UpdateConfigHandler) Validate(_ context.Context, req *UpdateConfigRequest) *errorModels.ServerError {
	if req.PointsPerDollar < 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "points_per_dollar must be non-negative"}
	}
	if req.RedemptionRate < 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "redemption_rate must be non-negative"}
	}
	if req.MinRedeemPoints < 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "min_redeem_points must be non-negative"}
	}
	return nil
}

func (h *UpdateConfigHandler) Serve(ctx context.Context, req *UpdateConfigRequest) (UpdateConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateConfigResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateConfigResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.UpdateConfig(ctx, domain.UpdateLoyaltyConfigInput{
		BusinessID:      businessID,
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

type (
	ListTransactionsHandler struct{ Service *application.Service }
	ListTransactionsRequest struct {
		CustomerID int64 `json:"customer_id"`
	}
)

type ListTransactionsResponse struct {
	Transactions []*posloyaltyv1.LoyaltyTransaction `json:"transactions"`
}

func (h *ListTransactionsHandler) Serve(ctx context.Context, req *ListTransactionsRequest) (ListTransactionsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListTransactionsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListTransactionsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.CustomerID == 0 {
		return ListTransactionsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "customer_id is required"}
	}
	txns, err := h.Service.ListTransactions(ctx, businessID, req.CustomerID)
	if err != nil {
		return ListTransactionsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list transactions"}
	}
	payloads := make([]*posloyaltyv1.LoyaltyTransaction, len(txns))
	for i, t := range txns {
		p := &posloyaltyv1.LoyaltyTransaction{
			Id:           t.ID,
			BusinessId:   t.BusinessID,
			CustomerId:   t.CustomerID,
			Type:         t.Type,
			Points:       t.Points,
			BalanceAfter: t.BalanceAfter,
			CreatedAt:    t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if t.OrderID != nil {
			p.OrderId = *t.OrderID
		}
		if t.Note != nil {
			p.Note = *t.Note
		}
		payloads[i] = p
	}
	return ListTransactionsResponse{Transactions: payloads}, nil
}

// ─── Earn Points ─────────────────────────────────────────────────────────────

func EarnPointsRoutes() []string { return []string{"/pos/loyalty/earn"} }

type EarnPointsHandler struct{ Service *application.Service }

type EarnPointsRequest struct {
	CustomerID int64  `json:"customer_id"`
	OrderID    *int64 `json:"order_id"`
	Points     int32  `json:"points"`
}

type EarnPointsResponse struct {
	TransactionID int64  `json:"transaction_id"`
	Message       string `json:"message"`
}

func (h *EarnPointsHandler) Serve(ctx context.Context, req *EarnPointsRequest) (EarnPointsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return EarnPointsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return EarnPointsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.CustomerID == 0 || req.Points <= 0 {
		return EarnPointsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "customerId and positive points are required"}
	}
	id, err := h.Service.EarnPoints(ctx, domain.EarnPointsInput{
		BusinessID: businessID,
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
	CustomerID int64 `json:"customer_id"`
	Points     int32 `json:"points"`
}

type RedeemPointsResponse struct {
	TransactionID int64  `json:"transaction_id"`
	Message       string `json:"message"`
}

func (h *RedeemPointsHandler) Serve(ctx context.Context, req *RedeemPointsRequest) (RedeemPointsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return RedeemPointsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return RedeemPointsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.CustomerID == 0 || req.Points <= 0 {
		return RedeemPointsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "customerId and positive points are required"}
	}
	id, err := h.Service.RedeemPoints(ctx, domain.RedeemPointsInput{
		BusinessID: businessID,
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

type (
	GetBalanceHandler struct{ Service *application.Service }
	GetBalanceRequest struct {
		CustomerID int64 `json:"customer_id"`
	}
)

type GetBalanceResponse struct {
	CustomerID int64 `json:"customer_id"`
	Balance    int32 `json:"balance"`
}

func (h *GetBalanceHandler) Serve(ctx context.Context, req *GetBalanceRequest) (GetBalanceResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetBalanceResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetBalanceResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.CustomerID == 0 {
		return GetBalanceResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "customer_id is required"}
	}
	balance, err := h.Service.GetBalance(ctx, businessID, req.CustomerID)
	if err != nil {
		return GetBalanceResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get balance"}
	}
	return GetBalanceResponse{CustomerID: req.CustomerID, Balance: balance}, nil
}
