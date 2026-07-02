package handlers

import (
	"context"
	"net/http"

	possettingsv1 "github.com/voxire/voxie/gen/go/pos/settings/v1"

	"github.com/voxire/onion/internal/modules/pos_settings/internal/application"
	"github.com/voxire/onion/internal/modules/pos_settings/internal/domain"
	contractHttp "github.com/voxire/onion/internal/platform/contracts"
)

type SettingsAPI struct {
	Service *application.Service
}

func NewSettingsAPI(service *application.Service) *SettingsAPI {
	return &SettingsAPI{Service: service}
}

func (a *SettingsAPI) GetPOSSettings(ctx context.Context, _ *possettingsv1.GetPOSSettingsRequest) (*possettingsv1.GetPOSSettingsResponse, error) {
	businessID, serverErr := businessIDFromContext(ctx)
	if serverErr != nil {
		return nil, contractHttp.NewServiceError(serverErr)
	}

	settings, err := a.Service.GetSettings(ctx, businessID)
	if err != nil {
		return nil, contractHttp.NewStatusError(http.StatusInternalServerError, "Failed to get POS settings")
	}

	return &possettingsv1.GetPOSSettingsResponse{Settings: settingsToProto(settings)}, nil
}

func (a *SettingsAPI) UpdatePOSSettings(ctx context.Context, req *possettingsv1.UpdatePOSSettingsRequest) (*possettingsv1.UpdatePOSSettingsResponse, error) {
	businessID, serverErr := businessIDFromContext(ctx)
	if serverErr != nil {
		return nil, contractHttp.NewServiceError(serverErr)
	}
	if req.GetBusinessProfile() != "" && domain.NormalizeProfile(req.GetBusinessProfile()) != req.GetBusinessProfile() {
		return nil, contractHttp.NewStatusError(http.StatusBadRequest, "Unsupported POS business profile")
	}
	if req.GetTheme() != "" && domain.NormalizeTheme(req.GetTheme()) != req.GetTheme() {
		return nil, contractHttp.NewStatusError(http.StatusBadRequest, "Unsupported POS theme")
	}

	existing, err := a.Service.GetSettings(ctx, businessID)
	if err != nil {
		return nil, contractHttp.NewStatusError(http.StatusInternalServerError, "Failed to get existing POS settings")
	}

	profile := existing.Profile
	if req.GetBusinessProfile() != "" {
		profile = req.GetBusinessProfile()
	}
	theme := existing.Theme
	if req.GetTheme() != "" {
		theme = req.GetTheme()
	}
	features := existing.Features
	hasFeatures := req.Features != nil
	if req.GetBusinessProfile() != "" && req.Features == nil {
		features = domain.DefaultFeatures(domain.NormalizeProfile(profile))
		hasFeatures = true
	}
	if req.Features != nil {
		features = featuresFromProto(req.GetFeatures())
	}
	receiptFooter := existing.ReceiptFooter
	if req.GetReceiptFooter() != "" {
		receiptFooter = req.GetReceiptFooter()
	}

	settings, err := a.Service.UpdateSettings(ctx, domain.UpdateSettingsInput{
		BusinessID:    businessID,
		Profile:       profile,
		Theme:         theme,
		Features:      features,
		HasFeatures:   hasFeatures,
		ReceiptFooter: receiptFooter,
	})
	if err != nil {
		return nil, contractHttp.NewStatusError(http.StatusInternalServerError, "Failed to update POS settings")
	}

	return &possettingsv1.UpdatePOSSettingsResponse{Settings: settingsToProto(settings), Message: "POS settings updated successfully"}, nil
}

func settingsToProto(settings domain.Settings) *possettingsv1.POSBusinessSettings {
	return &possettingsv1.POSBusinessSettings{
		Id:              settings.ID,
		BusinessId:      settings.BusinessID,
		BusinessProfile: settings.Profile,
		Theme:           settings.Theme,
		Features:        featuresToProto(settings.Features),
		ReceiptFooter:   settings.ReceiptFooter,
		CreatedAt:       settings.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       settings.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func featuresToProto(features domain.FeatureSettings) *possettingsv1.POSFeatureSettings {
	return &possettingsv1.POSFeatureSettings{
		Tables:         features.Tables,
		KitchenDisplay: features.KitchenDisplay,
		Modifiers:      features.Modifiers,
		Combos:         features.Combos,
		Tips:           features.Tips,
		GiftCards:      features.GiftCards,
		Customers:      features.Customers,
		Inventory:      features.Inventory,
		CashDrawer:     features.CashDrawer,
	}
}

func featuresFromProto(features *possettingsv1.POSFeatureSettings) domain.FeatureSettings {
	if features == nil {
		return domain.FeatureSettings{}
	}
	return domain.FeatureSettings{
		Tables:         features.GetTables(),
		KitchenDisplay: features.GetKitchenDisplay(),
		Modifiers:      features.GetModifiers(),
		Combos:         features.GetCombos(),
		Tips:           features.GetTips(),
		GiftCards:      features.GetGiftCards(),
		Customers:      features.GetCustomers(),
		Inventory:      features.GetInventory(),
		CashDrawer:     features.GetCashDrawer(),
	}
}
