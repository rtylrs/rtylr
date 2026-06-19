package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/application"
	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/platform/logger"
)

type GiftCardPayload struct {
	ID             int64   `json:"id"`
	OwnerID        int64   `json:"owner_id"`
	BusinessID     int64   `json:"business_id"`
	CustomerID     int64   `json:"customer_id"`
	Code           string  `json:"code"`
	InitialAmount  float64 `json:"initial_amount"`
	CurrentBalance float64 `json:"current_balance"`
	CustomerName   string  `json:"customer_name"`
	CustomerEmail  string  `json:"customer_email"`
	IsActive       bool    `json:"is_active"`
	ExpiresAt      string  `json:"expires_at"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

func newGiftCardPayload(gc domain.GiftCard) GiftCardPayload {
	expiresAt := ""
	if gc.ExpiresAt != nil {
		expiresAt = gc.ExpiresAt.Format("2006-01-02T15:04:05Z")
	}
	return GiftCardPayload{
		ID:             gc.ID,
		OwnerID:        gc.OwnerID,
		BusinessID:     gc.BusinessID,
		CustomerID:     gc.CustomerID,
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

// --- List ---

func ListGiftCardsRoutes() []string { return []string{"/pos/gift-cards"} }

type (
	ListGiftCardsHandler  struct{ Service *application.Service }
	ListGiftCardsRequest  struct{}
	ListGiftCardsResponse struct {
		GiftCards []GiftCardPayload `json:"gift_cards"`
	}
)

func (h *ListGiftCardsHandler) Serve(ctx context.Context, _ *ListGiftCardsRequest) (ListGiftCardsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListGiftCardsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListGiftCardsResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list gift cards", logger.Int("business_id", int(businessID)), logger.Error(err))
		return ListGiftCardsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list gift cards"}
	}
	payloads := make([]GiftCardPayload, len(items))
	for i, gc := range items {
		payloads[i] = newGiftCardPayload(gc)
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
	req.Code = strings.TrimSpace(req.Code)
	if req.Code == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Code is required"}
	}
	if req.InitialAmount <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Initial amount must be greater than zero"}
	}
	if req.ExpiresAt != "" {
		if _, err := time.Parse("2006-01-02", req.ExpiresAt); err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "expires_at must be a valid date in YYYY-MM-DD format"}
		}
	}
	return nil
}

func (h *CreateGiftCardHandler) Serve(ctx context.Context, req *CreateGiftCardRequest) (CreateGiftCardResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateGiftCardResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateGiftCardResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, businessID, domain.GiftCardInput{
		Code:          req.Code,
		InitialAmount: req.InitialAmount,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		ExpiresAt:     req.ExpiresAt,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create gift card", logger.Int("business_id", int(businessID)), logger.Error(err))
		return CreateGiftCardResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create gift card"}
	}
	return CreateGiftCardResponse{GiftCardID: id, Message: "Gift card created successfully"}, nil
}

// --- Get Balance By Code ---

func GetGiftCardByCodeRoutes() []string { return []string{"/pos/gift-cards/{code}"} }

type (
	GetGiftCardByCodeHandler  struct{ Service *application.Service }
	GetGiftCardByCodeRequest  struct{}
	GetGiftCardByCodeResponse struct {
		GiftCard GiftCardPayload `json:"gift_card"`
	}
)

func (h *GetGiftCardByCodeHandler) Serve(ctx context.Context, _ *GetGiftCardByCodeRequest) (GetGiftCardByCodeResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetGiftCardByCodeResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetGiftCardByCodeResponse{}, errorModels.PermissionDeniedServerError()
	}
	code := chi.URLParamFromCtx(ctx, "code")
	if code == "" {
		return GetGiftCardByCodeResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Code is required"}
	}
	gc, err := h.Service.GetByCode(ctx, businessID, code)
	if err != nil {
		if errors.Is(err, domain.ErrGiftCardNotFound) {
			return GetGiftCardByCodeResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Gift card not found"}
		}
		logger.Ctx(ctx).Error("failed to fetch gift card", logger.Int("business_id", int(businessID)), logger.String("code", code), logger.Error(err))
		return GetGiftCardByCodeResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to fetch gift card"}
	}
	return GetGiftCardByCodeResponse{GiftCard: newGiftCardPayload(*gc)}, nil
}

// --- Redeem By Code (path) ---

func RedeemGiftCardByCodeRoutes() []string { return []string{"/pos/gift-cards/{code}/redeem"} }

type (
	RedeemGiftCardByCodeHandler struct{ Service *application.Service }
	RedeemGiftCardByCodeRequest struct {
		Amount float64 `json:"amount"`
	}
)

type RedeemGiftCardByCodeResponse struct {
	GiftCard GiftCardPayload `json:"gift_card"`
	Message  string          `json:"message"`
}

func (h *RedeemGiftCardByCodeHandler) Validate(_ context.Context, req *RedeemGiftCardByCodeRequest) *errorModels.ServerError {
	if req.Amount <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Amount must be greater than zero"}
	}
	return nil
}

func (h *RedeemGiftCardByCodeHandler) Serve(ctx context.Context, req *RedeemGiftCardByCodeRequest) (RedeemGiftCardByCodeResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return RedeemGiftCardByCodeResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return RedeemGiftCardByCodeResponse{}, errorModels.PermissionDeniedServerError()
	}
	code := chi.URLParamFromCtx(ctx, "code")
	if code == "" {
		return RedeemGiftCardByCodeResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Code is required"}
	}
	gc, err := h.Service.Redeem(ctx, businessID, domain.RedeemInput{
		Code:   code,
		Amount: req.Amount,
	})
	if err != nil {
		if errors.Is(err, domain.ErrGiftCardNotFound) {
			return RedeemGiftCardByCodeResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Gift card not found"}
		}
		if errors.Is(err, domain.ErrGiftCardInactive) {
			return RedeemGiftCardByCodeResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Gift card is inactive"}
		}
		if errors.Is(err, domain.ErrGiftCardExpired) {
			return RedeemGiftCardByCodeResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Gift card has expired"}
		}
		if errors.Is(err, domain.ErrInsufficientBalance) {
			return RedeemGiftCardByCodeResponse{}, &errorModels.ServerError{HTTPCode: http.StatusConflict, Code: http.StatusConflict, Message: "Insufficient gift card balance"}
		}
		logger.Ctx(ctx).Error("failed to redeem gift card", logger.Int("business_id", int(businessID)), logger.String("code", code), logger.Error(err))
		return RedeemGiftCardByCodeResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to redeem gift card"}
	}
	return RedeemGiftCardByCodeResponse{GiftCard: newGiftCardPayload(*gc), Message: "Gift card redeemed successfully"}, nil
}

// --- Redeem / Check Balance (PUT) ---

func RedeemGiftCardRoutes() []string { return []string{"/pos/gift-cards"} }

type RedeemGiftCardHandler struct{ Service *application.Service }

type RedeemGiftCardRequest struct {
	Code   string  `json:"code"`
	Amount float64 `json:"amount"`
}

type RedeemGiftCardResponse struct {
	GiftCard GiftCardPayload `json:"gift_card"`
	Message  string          `json:"message"`
}

func (h *RedeemGiftCardHandler) Validate(_ context.Context, req *RedeemGiftCardRequest) *errorModels.ServerError {
	req.Code = strings.TrimSpace(req.Code)
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
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return RedeemGiftCardResponse{}, errorModels.PermissionDeniedServerError()
	}

	// If amount is 0, just check balance
	if req.Amount == 0 {
		gc, err := h.Service.GetByCode(ctx, businessID, req.Code)
		if err != nil {
			if errors.Is(err, domain.ErrGiftCardNotFound) {
				return RedeemGiftCardResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Gift card not found"}
			}
			logger.Ctx(ctx).Error("failed to check gift card", logger.Int("business_id", int(businessID)), logger.String("code", req.Code), logger.Error(err))
			return RedeemGiftCardResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to check gift card"}
		}
		return RedeemGiftCardResponse{GiftCard: newGiftCardPayload(*gc), Message: "Balance retrieved"}, nil
	}

	gc, err := h.Service.Redeem(ctx, businessID, domain.RedeemInput{
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
		logger.Ctx(ctx).Error("failed to redeem gift card", logger.Int("business_id", int(businessID)), logger.String("code", req.Code), logger.Error(err))
		return RedeemGiftCardResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to redeem gift card"}
	}
	return RedeemGiftCardResponse{GiftCard: newGiftCardPayload(*gc), Message: "Gift card redeemed successfully"}, nil
}
