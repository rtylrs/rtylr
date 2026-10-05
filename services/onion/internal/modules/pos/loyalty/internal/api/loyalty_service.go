package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos/loyalty/internal/application"
	"github.com/voxire/onion/internal/modules/pos/loyalty/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	posloyaltyv1 "github.com/voxire/voxie/gen/go/pos/loyalty/v1"
)

type LoyaltyAPI struct {
	Service *application.Service
}

func NewLoyaltyAPI(service *application.Service) *LoyaltyAPI {
	return &LoyaltyAPI{Service: service}
}

func loyaltyBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func (a *LoyaltyAPI) GetLoyaltyConfig(ctx context.Context, _ *posloyaltyv1.GetLoyaltyConfigRequest) (*posloyaltyv1.GetLoyaltyConfigResponse, error) {
	businessID, err := loyaltyBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	cfg, getErr := a.Service.GetConfig(ctx, businessID)
	if getErr != nil {
		if errors.Is(getErr, domain.ErrConfigNotFound) {
			return &posloyaltyv1.GetLoyaltyConfigResponse{Config: nil}, nil
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get loyalty config")
	}

	return &posloyaltyv1.GetLoyaltyConfigResponse{Config: &posloyaltyv1.LoyaltyConfig{
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

func (a *LoyaltyAPI) UpdateLoyaltyConfig(ctx context.Context, req *posloyaltyv1.UpdateLoyaltyConfigRequest) (*posloyaltyv1.UpdateLoyaltyConfigResponse, error) {
	businessID, err := loyaltyBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetPointsPerDollar() < 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "points_per_dollar must be non-negative")
	}
	if req.GetRedemptionRate() < 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "redemption_rate must be non-negative")
	}
	if req.GetMinRedeemPoints() < 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "min_redeem_points must be non-negative")
	}

	if updateErr := a.Service.UpdateConfig(ctx, domain.UpdateLoyaltyConfigInput{
		BusinessID:      businessID,
		IsEnabled:       req.GetIsEnabled(),
		PointsPerDollar: req.GetPointsPerDollar(),
		RedemptionRate:  req.GetRedemptionRate(),
		MinRedeemPoints: req.GetMinRedeemPoints(),
	}); updateErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update loyalty config")
	}

	return &posloyaltyv1.UpdateLoyaltyConfigResponse{Message: "Loyalty config updated successfully"}, nil
}

func (a *LoyaltyAPI) ListLoyaltyTransactions(ctx context.Context, req *posloyaltyv1.ListLoyaltyTransactionsRequest) (*posloyaltyv1.ListLoyaltyTransactionsResponse, error) {
	businessID, err := loyaltyBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetCustomerId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "customer_id is required")
	}

	txns, listErr := a.Service.ListTransactions(ctx, businessID, req.GetCustomerId())
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list transactions")
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
	return &posloyaltyv1.ListLoyaltyTransactionsResponse{Transactions: payloads}, nil
}

func (a *LoyaltyAPI) EarnPoints(ctx context.Context, req *posloyaltyv1.EarnPointsRequest) (*posloyaltyv1.EarnPointsResponse, error) {
	businessID, err := loyaltyBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetCustomerId() == 0 || req.GetOrderId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "customer_id and order_id are required")
	}

	orderID := req.GetOrderId()

	id, earnErr := a.Service.EarnPoints(ctx, domain.EarnPointsInput{
		BusinessID: businessID,
		CustomerID: req.GetCustomerId(),
		OrderID:    &orderID,
		Points:     req.GetPoints(), // validated for wire compatibility; service derives the award
	})
	if earnErr != nil {
		if errors.Is(earnErr, domain.ErrOrderRequired) || errors.Is(earnErr, domain.ErrOrderNotPaid) || errors.Is(earnErr, domain.ErrOrderAlreadyEarned) {
			return nil, onekiterrors.New(http.StatusConflict, earnErr.Error())
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to earn points")
	}
	return &posloyaltyv1.EarnPointsResponse{TransactionId: id, Message: "Points earned successfully"}, nil
}

func (a *LoyaltyAPI) RedeemPoints(ctx context.Context, req *posloyaltyv1.RedeemPointsRequest) (*posloyaltyv1.RedeemPointsResponse, error) {
	businessID, err := loyaltyBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetCustomerId() == 0 || req.GetPoints() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "customerId and positive points are required")
	}

	id, redeemErr := a.Service.RedeemPoints(ctx, domain.RedeemPointsInput{
		BusinessID: businessID,
		CustomerID: req.GetCustomerId(),
		Points:     req.GetPoints(),
	})
	if redeemErr != nil {
		if errors.Is(redeemErr, domain.ErrLoyaltyNotEnabled) {
			return nil, onekiterrors.New(http.StatusConflict, "Loyalty program is disabled")
		}
		if errors.Is(redeemErr, domain.ErrMinRedeemNotReached) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Minimum redeem points not reached")
		}
		if errors.Is(redeemErr, domain.ErrInsufficientPoints) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Insufficient loyalty points")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to redeem points")
	}
	return &posloyaltyv1.RedeemPointsResponse{TransactionId: id, Message: "Points redeemed successfully"}, nil
}

func (a *LoyaltyAPI) GetLoyaltyBalance(ctx context.Context, req *posloyaltyv1.GetLoyaltyBalanceRequest) (*posloyaltyv1.GetLoyaltyBalanceResponse, error) {
	businessID, err := loyaltyBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetCustomerId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "customer_id is required")
	}

	balance, balanceErr := a.Service.GetBalance(ctx, businessID, req.GetCustomerId())
	if balanceErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get balance")
	}
	return &posloyaltyv1.GetLoyaltyBalanceResponse{CustomerId: req.GetCustomerId(), Balance: balance}, nil
}
