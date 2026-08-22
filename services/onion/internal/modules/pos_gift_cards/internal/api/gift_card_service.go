package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/application"
	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	posgiftcardv1 "github.com/voxire/voxie/gen/go/pos/gift_card/v1"
)

type GiftCardAPI struct {
	Service *application.Service
}

func NewGiftCardAPI(service *application.Service) *GiftCardAPI {
	return &GiftCardAPI{Service: service}
}

func (a *GiftCardAPI) ListGiftCards(ctx context.Context, _ *posgiftcardv1.ListGiftCardsRequest) (*posgiftcardv1.ListGiftCardsResponse, error) {
	businessID, err := authenticatedGiftCardBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	items, listErr := a.Service.List(ctx, businessID)
	if listErr != nil {
		logger.Ctx(ctx).Error("failed to list gift cards", logger.Int("business_id", int(businessID)), logger.Error(listErr))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list gift cards")
	}

	payloads := make([]*posgiftcardv1.GiftCard, len(items))
	for i, gc := range items {
		payloads[i] = toProtoGiftCard(gc)
	}
	return &posgiftcardv1.ListGiftCardsResponse{GiftCards: payloads}, nil
}

func (a *GiftCardAPI) CreateGiftCard(ctx context.Context, req *posgiftcardv1.CreateGiftCardRequest) (*posgiftcardv1.CreateGiftCardResponse, error) {
	if err := requireGiftCardManager(ctx); err != nil {
		return nil, err
	}
	businessID, err := authenticatedGiftCardBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.GetCode()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Code is required")
	}
	if req.GetInitialAmount() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Initial amount must be greater than zero")
	}
	if expiresAt := req.GetExpiresAt(); expiresAt != "" {
		if _, parseErr := time.Parse("2006-01-02", expiresAt); parseErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "expires_at must be a valid date in YYYY-MM-DD format")
		}
	}

	device := authPkg.DeviceCtx(ctx)
	id, createErr := a.Service.CreateForOwner(ctx, businessID, device.User.UserID, domain.GiftCardInput{
		PaidOrderID:   req.GetPaidOrderId(),
		Code:          req.GetCode(),
		InitialAmount: req.GetInitialAmount(),
		CustomerName:  req.GetCustomerName(),
		CustomerEmail: req.GetCustomerEmail(),
		ExpiresAt:     req.GetExpiresAt(),
	})
	if createErr != nil {
		if errors.Is(createErr, domain.ErrInvalidTender) {
			return nil, onekiterrors.New(http.StatusBadRequest, "paid order is invalid or not fully paid")
		}
		logger.Ctx(ctx).Error("failed to create gift card", logger.Int("business_id", int(businessID)), logger.Error(createErr))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create gift card")
	}
	return &posgiftcardv1.CreateGiftCardResponse{GiftCardId: id, Message: "Gift card created successfully"}, nil
}

func (a *GiftCardAPI) GetGiftCardByCode(ctx context.Context, req *posgiftcardv1.GetGiftCardByCodeRequest) (*posgiftcardv1.GetGiftCardByCodeResponse, error) {
	businessID, err := authenticatedGiftCardBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	code := strings.TrimSpace(req.GetCode())
	if code == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Code is required")
	}

	gc, getErr := a.Service.GetByCode(ctx, businessID, code)
	if getErr != nil {
		if errors.Is(getErr, domain.ErrGiftCardNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Gift card not found")
		}
		logger.Ctx(ctx).Error("failed to fetch gift card", logger.Int("business_id", int(businessID)), logger.String("code", code), logger.Error(getErr))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to fetch gift card")
	}
	return &posgiftcardv1.GetGiftCardByCodeResponse{GiftCard: toProtoGiftCard(*gc)}, nil
}

func (a *GiftCardAPI) RedeemGiftCard(ctx context.Context, req *posgiftcardv1.RedeemGiftCardRequest) (*posgiftcardv1.RedeemGiftCardResponse, error) {
	if err := requireGiftCardManager(ctx); err != nil {
		return nil, err
	}
	businessID, err := authenticatedGiftCardBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	code := strings.TrimSpace(req.GetCode())
	if code == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Code is required")
	}
	if req.GetAmount() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Amount must be greater than zero")
	}

	device := authPkg.DeviceCtx(ctx)
	gc, redeemErr := a.Service.RedeemForOrder(ctx, businessID, device.User.UserID, domain.RedeemInput{
		OrderID: req.GetOrderId(),
		Code:    code,
		Amount:  req.GetAmount(),
	})
	if redeemErr != nil {
		if errors.Is(redeemErr, domain.ErrGiftCardNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Gift card not found")
		}
		if errors.Is(redeemErr, domain.ErrGiftCardInactive) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Gift card is inactive")
		}
		if errors.Is(redeemErr, domain.ErrGiftCardExpired) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Gift card has expired")
		}
		if errors.Is(redeemErr, domain.ErrInsufficientBalance) {
			return nil, onekiterrors.New(http.StatusConflict, "Insufficient gift card balance")
		}
		if errors.Is(redeemErr, domain.ErrInvalidOrder) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid gift card order reference")
		}
		logger.Ctx(ctx).Error("failed to redeem gift card", logger.Int("business_id", int(businessID)), logger.String("code", code), logger.Error(redeemErr))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to redeem gift card")
	}
	return &posgiftcardv1.RedeemGiftCardResponse{GiftCard: toProtoGiftCard(*gc), Message: "Gift card redeemed successfully"}, nil
}

func authenticatedGiftCardBusinessID(ctx context.Context) (int64, error) {
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

func requireGiftCardManager(ctx context.Context) error {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || !authPkg.HasAnyRole(authUser, "owner", "admin", "console_admin", "pos_manager") {
		return onekiterrors.PermissionDenied()
	}
	return nil
}

func toProtoGiftCard(gc domain.GiftCard) *posgiftcardv1.GiftCard {
	expiresAt := ""
	if gc.ExpiresAt != nil {
		expiresAt = gc.ExpiresAt.Format("2006-01-02T15:04:05Z")
	}
	return &posgiftcardv1.GiftCard{
		Id:             gc.ID,
		OwnerId:        gc.OwnerID,
		BusinessId:     gc.BusinessID,
		CustomerId:     gc.CustomerID,
		Code:           gc.Code,
		InitialAmount:  gc.InitialAmount,
		CurrentBalance: gc.CurrentBalance,
		CustomerName:   gc.CustomerName,
		CustomerEmail:  gc.CustomerEmail,
		IsActive:       gc.IsActive,
		ExpiresAt:      expiresAt,
		CreatedAt:      gc.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      gc.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
