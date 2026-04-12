package handlers

import (
	posv1 "github.com/voxire/voxie/gen/go/pos/v1"

	"context"
	"errors"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/application"
	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/domain"
)

func giftCardToProto(gc domain.GiftCard) *posv1.GiftCard {
	expiresAt := ""
	if gc.ExpiresAt != nil {
		expiresAt = gc.ExpiresAt.Format("2006-01-02T15:04:05Z")
	}
	return &posv1.GiftCard{
		Id:             gc.ID,
		Code:           gc.Code,
		InitialBalance: gc.InitialAmount,
		CurrentBalance: gc.CurrentBalance,
		CustomerName:   gc.CustomerName,
		IsActive:       gc.IsActive,
		ExpiresAt:      expiresAt,
		CreatedAt:      gc.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      gc.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- List ---

func ListGiftCardsRoutes() []string { return []string{"/pos/gift-cards"} }

type ListGiftCardsHandler struct{ Service *application.Service }
type ListGiftCardsRequest struct{}
type ListGiftCardsResponse struct {
	GiftCards []*posv1.GiftCard `json:"gift_cards"`
}

func (h *ListGiftCardsHandler) Serve(ctx context.Context, _ *ListGiftCardsRequest) (ListGiftCardsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListGiftCardsResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListGiftCardsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list gift cards"}
	}
	payloads := make([]*posv1.GiftCard, len(items))
	for i, gc := range items {
		payloads[i] = giftCardToProto(gc)
	}
	return ListGiftCardsResponse{GiftCards: payloads}, nil
}

// --- Create ---

func CreateGiftCardRoutes() []string { return []string{"/pos/gift-cards"} }

type CreateGiftCardHandler struct{ Service *application.Service }

type CreateGiftCardRequest struct {
	Code          string  `json:"code"`
	InitialAmount float64 `json:"initial_amount"`
	CustomerName  string  `json:"customer_name"`
	CustomerEmail string  `json:"customer_email"`
	ExpiresAt     string  `json:"expires_at"`
}

type CreateGiftCardResponse struct {
	GiftCardID int64  `json:"gift_card_id"`
	Message    string `json:"message"`
}

func (h *CreateGiftCardHandler) Validate(_ context.Context, req *CreateGiftCardRequest) *errorModels.ServerError {
	if req.Code == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Code is required"}
	}
	if req.InitialAmount <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Initial amount must be greater than zero"}
	}
	return nil
}

func (h *CreateGiftCardHandler) Serve(ctx context.Context, req *CreateGiftCardRequest) (CreateGiftCardResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateGiftCardResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, device.User.UserID, domain.GiftCardInput{
		Code:          req.Code,
		InitialAmount: req.InitialAmount,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		ExpiresAt:     req.ExpiresAt,
	})
	if err != nil {
		return CreateGiftCardResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create gift card"}
	}
	return CreateGiftCardResponse{GiftCardID: id, Message: "Gift card created successfully"}, nil
}

// --- Redeem / Check Balance (PUT) ---

func RedeemGiftCardRoutes() []string { return []string{"/pos/gift-cards"} }

type RedeemGiftCardHandler struct{ Service *application.Service }

type RedeemGiftCardRequest struct {
	Code   string  `json:"code"`
	Amount float64 `json:"amount"`
}

type RedeemGiftCardResponse struct {
	GiftCard *posv1.GiftCard `json:"gift_card"`
	Message  string          `json:"message"`
}

func (h *RedeemGiftCardHandler) Validate(_ context.Context, req *RedeemGiftCardRequest) *errorModels.ServerError {
	if req.Code == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Code is required"}
	}
	return nil
}

func (h *RedeemGiftCardHandler) Serve(ctx context.Context, req *RedeemGiftCardRequest) (RedeemGiftCardResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return RedeemGiftCardResponse{}, errorModels.PermissionDeniedServerError()
	}

	// If amount is 0, just check balance
	if req.Amount == 0 {
		gc, err := h.Service.GetByCode(ctx, device.User.UserID, req.Code)
		if err != nil {
			if errors.Is(err, domain.ErrGiftCardNotFound) {
				return RedeemGiftCardResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Gift card not found"}
			}
			return RedeemGiftCardResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to check gift card"}
		}
		return RedeemGiftCardResponse{GiftCard: *giftCardToProto(*gc), Message: "Balance retrieved"}, nil
	}

	gc, err := h.Service.Redeem(ctx, device.User.UserID, domain.RedeemInput{
		Code:   req.Code,
		Amount: req.Amount,
	})
	if err != nil {
		if errors.Is(err, domain.ErrGiftCardNotFound) {
			return RedeemGiftCardResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Gift card not found"}
		}
		if errors.Is(err, domain.ErrGiftCardInactive) {
			return RedeemGiftCardResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Gift card is inactive"}
		}
		if errors.Is(err, domain.ErrGiftCardExpired) {
			return RedeemGiftCardResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Gift card has expired"}
		}
		if errors.Is(err, domain.ErrInsufficientBalance) {
			return RedeemGiftCardResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Insufficient gift card balance"}
		}
		return RedeemGiftCardResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to redeem gift card"}
	}
	return RedeemGiftCardResponse{GiftCard: *giftCardToProto(*gc), Message: "Gift card redeemed successfully"}, nil
}
