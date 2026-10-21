package api

import (
	"context"
	"math"
	"net/http"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/application"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmloyaltyv1 "github.com/voxire/voxie/gen/go/crm/loyalty/v1"
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

type GetLoyaltyConfigResponse struct {
	Config *crmloyaltyv1.LoyaltyConfig `json:"config"`
}

func newLoyaltyConfigPayload(cfg domain.LoyaltyConfig) *crmloyaltyv1.LoyaltyConfig {
	return &crmloyaltyv1.LoyaltyConfig{
		Id:                    cfg.ID,
		OwnerId:               cfg.OwnerID,
		IsEnabled:             cfg.IsEnabled,
		PointsPerCurrencyUnit: cfg.PointsPerCurrencyUnit,
		CurrencyPerPoint:      cfg.CurrencyPerPoint,
		BronzeMin:             int64(cfg.BronzeMin),
		SilverMin:             int64(cfg.SilverMin),
		GoldMin:               int64(cfg.GoldMin),
		PlatinumMin:           int64(cfg.PlatinumMin),
		BronzeDiscountPct:     cfg.BronzeDiscountPct,
		SilverDiscountPct:     cfg.SilverDiscountPct,
		GoldDiscountPct:       cfg.GoldDiscountPct,
		PlatinumDiscountPct:   cfg.PlatinumDiscountPct,
		PointsExpiryDays:      int64(cfg.PointsExpiryDays),
		UpdatedAt:             cfg.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *GetLoyaltyConfigHandler) Serve(ctx context.Context, _ *GetLoyaltyConfigRequest) (GetLoyaltyConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetLoyaltyConfigResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetLoyaltyConfigResponse{}, onekiterrors.PermissionDenied()
	}

	cfg, err := h.CRMService.GetLoyaltyConfig(ctx, businessID)
	if err != nil {
		return GetLoyaltyConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get loyalty config")
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

func validateLoyaltyConfigBounds(bronzeMin, silverMin, goldMin, platinumMin, pointsExpiryDays int64) error {
	values := []struct {
		name  string
		value int64
	}{
		{name: "bronze_min", value: bronzeMin},
		{name: "silver_min", value: silverMin},
		{name: "gold_min", value: goldMin},
		{name: "platinum_min", value: platinumMin},
		{name: "points_expiry_days", value: pointsExpiryDays},
	}
	for _, item := range values {
		if item.value < 0 || item.value > math.MaxInt32 {
			return onekiterrors.New(http.StatusBadRequest, item.name+" must be between 0 and 2147483647")
		}
	}
	if bronzeMin > silverMin || silverMin > goldMin || goldMin > platinumMin {
		return onekiterrors.New(http.StatusBadRequest, "loyalty tier thresholds must be in ascending order")
	}
	return nil
}

func (h *UpdateLoyaltyConfigHandler) Validate(_ context.Context, req *UpdateLoyaltyConfigRequest) error {
	if req == nil {
		return onekiterrors.New(http.StatusBadRequest, "Loyalty config is required")
	}
	if req.PointsPerCurrencyUnit < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Points per currency unit must not be negative")
	}
	if req.CurrencyPerPoint < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Currency per point must not be negative")
	}
	if req.BronzeDiscountPct < 0 || req.SilverDiscountPct < 0 || req.GoldDiscountPct < 0 || req.PlatinumDiscountPct < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Discount percentages must not be negative")
	}
	if err := validateLoyaltyConfigBounds(
		int64(req.BronzeMin),
		int64(req.SilverMin),
		int64(req.GoldMin),
		int64(req.PlatinumMin),
		int64(req.PointsExpiryDays),
	); err != nil {
		return err
	}
	return nil
}

func (h *UpdateLoyaltyConfigHandler) Serve(ctx context.Context, req *UpdateLoyaltyConfigRequest) (UpdateLoyaltyConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateLoyaltyConfigResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:loyalty:manage"); err != nil {
		return UpdateLoyaltyConfigResponse{}, err
	}
	if err := h.Validate(ctx, req); err != nil {
		return UpdateLoyaltyConfigResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateLoyaltyConfigResponse{}, onekiterrors.PermissionDenied()
	}

	input := domain.LoyaltyConfigInput{
		OwnerID:               device.User.UserID,
		BusinessID:            businessID,
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
		return UpdateLoyaltyConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update loyalty config")
	}

	return UpdateLoyaltyConfigResponse{Message: "Loyalty config updated successfully"}, nil
}
