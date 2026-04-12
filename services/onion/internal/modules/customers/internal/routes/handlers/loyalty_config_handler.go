package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
)

func LoyaltyConfigGetRoutes() []string {
	return []string{"/crm/loyalty/config"}
}

func LoyaltyConfigUpdateRoutes() []string {
	return []string{"/crm/loyalty/config"}
}

// ─── GET config ───────────────────────────────────────────────────────────────

type GetLoyaltyConfigHandler struct {
	CRMService *application.Service
}

type GetLoyaltyConfigRequest struct{}

type LoyaltyConfigPayload struct {
	ID                    int64   `json:"id"`
	OwnerID               int64   `json:"owner_id"`
	IsEnabled             bool    `json:"is_enabled"`
	PointsPerCurrencyUnit float64 `json:"points_per_currency_unit"`
	CurrencyPerPoint      float64 `json:"currency_per_point"`
	BronzeMin             int     `json:"bronze_min"`
	SilverMin             int     `json:"silver_min"`
	GoldMin               int     `json:"gold_min"`
	PlatinumMin           int     `json:"platinum_min"`
	BronzeDiscountPct     float64 `json:"bronze_discount_pct"`
	SilverDiscountPct     float64 `json:"silver_discount_pct"`
	GoldDiscountPct       float64 `json:"gold_discount_pct"`
	PlatinumDiscountPct   float64 `json:"platinum_discount_pct"`
	PointsExpiryDays      int     `json:"points_expiry_days"`
	UpdatedAt             string  `json:"updated_at"`
}

type GetLoyaltyConfigResponse struct {
	Config LoyaltyConfigPayload `json:"config"`
}

func newLoyaltyConfigPayload(cfg domain.LoyaltyConfig) LoyaltyConfigPayload {
	return LoyaltyConfigPayload{
		ID:                    cfg.ID,
		OwnerID:               cfg.OwnerID,
		IsEnabled:             cfg.IsEnabled,
		PointsPerCurrencyUnit: cfg.PointsPerCurrencyUnit,
		CurrencyPerPoint:      cfg.CurrencyPerPoint,
		BronzeMin:             cfg.BronzeMin,
		SilverMin:             cfg.SilverMin,
		GoldMin:               cfg.GoldMin,
		PlatinumMin:           cfg.PlatinumMin,
		BronzeDiscountPct:     cfg.BronzeDiscountPct,
		SilverDiscountPct:     cfg.SilverDiscountPct,
		GoldDiscountPct:       cfg.GoldDiscountPct,
		PlatinumDiscountPct:   cfg.PlatinumDiscountPct,
		PointsExpiryDays:      cfg.PointsExpiryDays,
		UpdatedAt:             cfg.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *GetLoyaltyConfigHandler) Serve(ctx context.Context, _ *GetLoyaltyConfigRequest) (GetLoyaltyConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetLoyaltyConfigResponse{}, errorModels.PermissionDeniedServerError()
	}

	cfg, err := h.CRMService.GetLoyaltyConfig(ctx, device.User.UserID)
	if err != nil {
		return GetLoyaltyConfigResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get loyalty config",
		}
	}

	return GetLoyaltyConfigResponse{Config: newLoyaltyConfigPayload(cfg)}, nil
}

// ─── PUT config ───────────────────────────────────────────────────────────────

type UpdateLoyaltyConfigHandler struct {
	CRMService *application.Service
}

type UpdateLoyaltyConfigRequest struct {
	IsEnabled             bool    `json:"is_enabled"`
	PointsPerCurrencyUnit float64 `json:"points_per_currency_unit"`
	CurrencyPerPoint      float64 `json:"currency_per_point"`
	BronzeMin             int     `json:"bronze_min"`
	SilverMin             int     `json:"silver_min"`
	GoldMin               int     `json:"gold_min"`
	PlatinumMin           int     `json:"platinum_min"`
	BronzeDiscountPct     float64 `json:"bronze_discount_pct"`
	SilverDiscountPct     float64 `json:"silver_discount_pct"`
	GoldDiscountPct       float64 `json:"gold_discount_pct"`
	PlatinumDiscountPct   float64 `json:"platinum_discount_pct"`
	PointsExpiryDays      int     `json:"points_expiry_days"`
}

type UpdateLoyaltyConfigResponse struct {
	Message string `json:"message"`
}

func (h *UpdateLoyaltyConfigHandler) Serve(ctx context.Context, req *UpdateLoyaltyConfigRequest) (UpdateLoyaltyConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateLoyaltyConfigResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.LoyaltyConfigInput{
		OwnerID:               device.User.UserID,
		IsEnabled:             req.IsEnabled,
		PointsPerCurrencyUnit: req.PointsPerCurrencyUnit,
		CurrencyPerPoint:      req.CurrencyPerPoint,
		BronzeMin:             req.BronzeMin,
		SilverMin:             req.SilverMin,
		GoldMin:               req.GoldMin,
		PlatinumMin:           req.PlatinumMin,
		BronzeDiscountPct:     req.BronzeDiscountPct,
		SilverDiscountPct:     req.SilverDiscountPct,
		GoldDiscountPct:       req.GoldDiscountPct,
		PlatinumDiscountPct:   req.PlatinumDiscountPct,
		PointsExpiryDays:      req.PointsExpiryDays,
	}

	if err := h.CRMService.UpsertLoyaltyConfig(ctx, input); err != nil {
		return UpdateLoyaltyConfigResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update loyalty config",
		}
	}

	return UpdateLoyaltyConfigResponse{Message: "Loyalty config updated successfully"}, nil
}
