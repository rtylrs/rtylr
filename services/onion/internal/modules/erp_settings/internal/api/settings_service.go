package api

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/erp_settings/internal/application"
	"github.com/voxire/onion/internal/modules/erp_settings/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	erpsettingsv1 "github.com/voxire/voxie/gen/go/erp/settings/v1"
)

type SettingsAPI struct {
	service *application.Service
}

func NewSettingsAPI(service *application.Service) *SettingsAPI {
	return &SettingsAPI{service: service}
}

func (a *SettingsAPI) GetErpSettings(ctx context.Context, _ *erpsettingsv1.GetErpSettingsRequest) (*erpsettingsv1.GetErpSettingsResponse, error) {
	businessID, err := businessIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	settings, err := a.service.GetSettings(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get ERP settings")
	}

	return &erpsettingsv1.GetErpSettingsResponse{Settings: settingsToProto(settings)}, nil
}

func (a *SettingsAPI) UpdateErpSettings(ctx context.Context, req *erpsettingsv1.UpdateErpSettingsRequest) (*erpsettingsv1.UpdateErpSettingsResponse, error) {
	if req == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid ERP settings")
	}
	if err := req.Validate(); err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
	}

	businessID, err := businessIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	settings, err := a.service.UpdateSettings(ctx, domain.SettingsInput{
		BusinessID:         businessID,
		Currency:           req.GetCurrency(),
		TaxRate:            req.GetTaxRate(),
		LowStockThreshold:  req.GetLowStockThreshold(),
		DefaultWarehouseID: req.GetDefaultWarehouseId(),
		NotifyOnLowStock:   req.GetNotifyOnLowStock(),
		NotifyOnExpiry:     req.GetNotifyOnExpiry(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update ERP settings")
	}

	return &erpsettingsv1.UpdateErpSettingsResponse{
		Settings: settingsToProto(settings),
		Message:  "ERP settings updated successfully",
	}, nil
}

func businessIDFromContext(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}

	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func settingsToProto(settings domain.Settings) *erpsettingsv1.ErpBusinessSettings {
	taxRate := settings.TaxRate
	lowStockThreshold := settings.LowStockThreshold
	defaultWarehouseID := settings.DefaultWarehouseID

	return &erpsettingsv1.ErpBusinessSettings{
		Id:                 settings.ID,
		BusinessId:         settings.BusinessID,
		Currency:           settings.Currency,
		TaxRate:            &taxRate,
		LowStockThreshold:  &lowStockThreshold,
		DefaultWarehouseId: &defaultWarehouseID,
		NotifyOnLowStock:   settings.NotifyOnLowStock,
		NotifyOnExpiry:     settings.NotifyOnExpiry,
		CreatedAt:          formatTime(settings.CreatedAt),
		UpdatedAt:          formatTime(settings.UpdatedAt),
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
