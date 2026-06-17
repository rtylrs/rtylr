package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_settings/internal/application"
	"github.com/voxire/onion/internal/modules/pos_settings/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func GetSettingsRoutes() []string    { return []string{"/pos/settings"} }
func UpdateSettingsRoutes() []string { return []string{"/pos/settings"} }

type (
	GetSettingsHandler struct{ Service *application.Service }
	GetSettingsRequest struct{}
)

type SettingsPayload struct {
	ID              int64                  `json:"id"`
	BusinessID      int64                  `json:"business_id"`
	BusinessProfile string                 `json:"business_profile"`
	Theme           string                 `json:"theme"`
	Features        domain.FeatureSettings `json:"features"`
	ReceiptFooter   string                 `json:"receipt_footer"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
}

type GetSettingsResponse struct {
	Settings SettingsPayload `json:"settings"`
}

func (h *GetSettingsHandler) Serve(ctx context.Context, _ *GetSettingsRequest) (GetSettingsResponse, *errorModels.ServerError) {
	businessID, serverErr := businessIDFromContext(ctx)
	if serverErr != nil {
		return GetSettingsResponse{}, serverErr
	}
	settings, err := h.Service.GetSettings(ctx, businessID)
	if err != nil {
		return GetSettingsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get POS settings"}
	}
	return GetSettingsResponse{Settings: toPayload(settings)}, nil
}

type UpdateSettingsHandler struct{ Service *application.Service }

type UpdateSettingsRequest struct {
	BusinessProfile string                  `json:"business_profile"`
	Theme           string                  `json:"theme"`
	Features        *domain.FeatureSettings `json:"features"`
	ReceiptFooter   string                  `json:"receipt_footer"`
}

type UpdateSettingsResponse struct {
	Settings SettingsPayload `json:"settings"`
	Message  string          `json:"message"`
}

func (h *UpdateSettingsHandler) Validate(_ context.Context, req *UpdateSettingsRequest) *errorModels.ServerError {
	if req.BusinessProfile != "" && domain.NormalizeProfile(req.BusinessProfile) != req.BusinessProfile {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Unsupported POS business profile"}
	}
	if req.Theme != "" && domain.NormalizeTheme(req.Theme) != req.Theme {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Unsupported POS theme"}
	}
	return nil
}

func (h *UpdateSettingsHandler) Serve(ctx context.Context, req *UpdateSettingsRequest) (UpdateSettingsResponse, *errorModels.ServerError) {
	businessID, serverErr := businessIDFromContext(ctx)
	if serverErr != nil {
		return UpdateSettingsResponse{}, serverErr
	}

	existing, err := h.Service.GetSettings(ctx, businessID)
	if err != nil {
		return UpdateSettingsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get existing POS settings"}
	}

	profile := existing.Profile
	if req.BusinessProfile != "" {
		profile = req.BusinessProfile
	}
	theme := existing.Theme
	if req.Theme != "" {
		theme = req.Theme
	}
	features := existing.Features
	if req.BusinessProfile != "" && req.Features == nil {
		features = domain.DefaultFeatures(domain.NormalizeProfile(profile))
	}
	if req.Features != nil {
		features = *req.Features
	}
	receiptFooter := existing.ReceiptFooter
	if req.ReceiptFooter != "" {
		receiptFooter = req.ReceiptFooter
	}

	settings, err := h.Service.UpdateSettings(ctx, domain.UpdateSettingsInput{
		BusinessID:    businessID,
		Profile:       profile,
		Theme:         theme,
		Features:      features,
		HasFeatures:   req.Features != nil || req.BusinessProfile != "",
		ReceiptFooter: receiptFooter,
	})
	if err != nil {
		return UpdateSettingsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update POS settings"}
	}

	return UpdateSettingsResponse{
		Settings: toPayload(settings),
		Message:  "POS settings updated successfully",
	}, nil
}

func businessIDFromContext(ctx context.Context) (int64, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, errorModels.PermissionDeniedServerError()
	}
	if device.User.BusinessID == 0 {
		return 0, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Business context is required"}
	}
	return device.User.BusinessID, nil
}

func toPayload(settings domain.Settings) SettingsPayload {
	return SettingsPayload{
		ID:              settings.ID,
		BusinessID:      settings.BusinessID,
		BusinessProfile: settings.Profile,
		Theme:           settings.Theme,
		Features:        settings.Features,
		ReceiptFooter:   settings.ReceiptFooter,
		CreatedAt:       settings.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       settings.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
