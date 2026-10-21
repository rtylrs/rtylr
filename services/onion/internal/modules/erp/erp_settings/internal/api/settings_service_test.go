package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/erp/erp_settings/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	erpsettingsv1 "github.com/voxire/voxie/gen/go/erp/settings/v1"
)

func TestSettingsAPIRequiresExactPermissionBeforeServiceAccess(t *testing.T) {
	user := &authPkg.AuthUser{
		UserID:     7,
		BusinessID: 9,
		EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID:       authorization.AppERP,
			Licensed:    true,
			HasAccess:   true,
			Permissions: []string{"erp:settings:view"},
		}}},
	}
	ctx := authPkg.WithAuthUser(context.Background(), user)
	api := NewSettingsAPI(nil)

	assertForbidden := func(t *testing.T, err error) {
		t.Helper()
		var httpErr *onekithttp.Error
		if !errors.As(err, &httpErr) || int(httpErr.GetStatusCode()) != http.StatusForbidden {
			t.Fatalf("error = %v, want forbidden", err)
		}
	}

	t.Run("read without view", func(t *testing.T) {
		user.EffectiveAccess.Apps[0].Permissions = nil
		_, err := api.GetErpSettings(ctx, &erpsettingsv1.GetErpSettingsRequest{})
		assertForbidden(t, err)
	})
	t.Run("write with view only", func(t *testing.T) {
		user.EffectiveAccess.Apps[0].Permissions = []string{"erp:settings:view"}
		_, err := api.UpdateErpSettings(ctx, &erpsettingsv1.UpdateErpSettingsRequest{
			Currency:           "USD",
			TaxRate:            testutil.Float64Ptr(0),
			LowStockThreshold:  testutil.Float64Ptr(0),
			DefaultWarehouseId: testutil.Int64Ptr(1),
		})
		assertForbidden(t, err)
	})
}

func TestSettingsToProtoPreservesDisabledNotificationFlags(t *testing.T) {
	settings := settingsToProto(domain.Settings{
		ID:                 3,
		BusinessID:         42,
		Currency:           "EUR",
		TaxRate:            20,
		LowStockThreshold:  4,
		DefaultWarehouseID: 7,
		NotifyOnLowStock:   false,
		NotifyOnExpiry:     false,
	})

	if settings.NotifyOnLowStock || settings.NotifyOnExpiry {
		t.Fatalf("disabled notification flags were changed: %+v", settings)
	}
	if settings.TaxRate == nil || settings.LowStockThreshold == nil || settings.DefaultWarehouseId == nil {
		t.Fatalf("required settings pointers were not populated: %+v", settings)
	}
	if settings.DefaultWarehouseId == nil || *settings.DefaultWarehouseId != 7 {
		t.Fatalf("default warehouse = %v, want 7", settings.DefaultWarehouseId)
	}
	if err := settings.Validate(); err != nil {
		t.Fatalf("generated contract validation failed: %v", err)
	}
}
