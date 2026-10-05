package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_equipment/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrequipmentv1 "github.com/voxire/voxie/gen/go/hr/equipment/v1"
)

type CreateAssetHandler struct{ Service *application.Service }

type CreateAssetRequest struct {
	Name                    string `json:"name"`
	Category                string `json:"category"`
	SerialNumber            string `json:"serial_number"`
	LocationID              int64  `json:"location_id"`
	Condition               string `json:"condition"`
	PurchaseDate            string `json:"purchase_date"`
	PurchasePriceMinorUnits int64  `json:"purchase_price_minor_units"`
	Notes                   string `json:"notes"`
}

type CreateAssetResponse struct {
	Asset *hrequipmentv1.AssetPayload `json:"asset"`
}

func (h *CreateAssetHandler) Validate(_ context.Context, req *CreateAssetRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "name is required")
	}
	req.Category = strings.TrimSpace(req.Category)
	if req.Category == "" {
		return onekiterrors.New(http.StatusBadRequest, "category is required")
	}
	req.Condition = strings.TrimSpace(req.Condition)
	if req.Condition == "" {
		req.Condition = domain.AssetConditionGood
	}
	if !domain.IsValidAssetCondition(req.Condition) {
		return onekiterrors.New(http.StatusBadRequest, "condition must be one of new, good, fair, damaged")
	}
	if req.PurchasePriceMinorUnits < 0 {
		return onekiterrors.New(http.StatusBadRequest, "purchase_price_minor_units must be non-negative")
	}
	return nil
}
