package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/integrations/internal/application"
	"github.com/voxire/onion/internal/modules/integrations/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/tenant"
	hubintegrationv1 "github.com/voxire/voxie/gen/go/hub/integration/v1"
)

func userWith(access *authorization.EffectiveAccess) *authPkg.AuthUser {
	return &authPkg.AuthUser{UserID: 5, BusinessID: 1, EffectiveAccess: access}
}

func TestConfigIsHiddenFromOrdinaryMembers(t *testing.T) {
	config := json.RawMessage(`{"secret_key":"sk_live_x","mode":"live"}`)
	if got := configFor(userWith(&authorization.EffectiveAccess{}), config); got != nil {
		t.Fatalf("a member received integration config: %s", got)
	}
}

func TestConfigIsRedactedForAdmins(t *testing.T) {
	config := json.RawMessage(`{"secret_key":"sk_live_x","mode":"live"}`)
	got := configFor(userWith(&authorization.EffectiveAccess{IsAdmin: true}), config)
	if string(got) != `{"mode":"live","secret_key":"[redacted]"}` {
		t.Fatalf("admin config = %s", got)
	}
}

type businessScope struct{}

func (businessScope) ActiveBusinessOwner(context.Context, int64) (int64, error) { return 1, nil }
func (businessScope) UserBusiness(_ context.Context, businessID, _ int64) (int64, error) {
	return businessID, nil
}

type missingRepo struct{}

func (missingRepo) List(context.Context, int64) ([]domain.Integration, error) { return nil, nil }
func (missingRepo) Connect(context.Context, int64, string, json.RawMessage) (int64, error) {
	return 1, nil
}
func (missingRepo) Disconnect(context.Context, int64, int64) error { return domain.ErrNotFound }

func TestDisconnectingAMissingIntegrationIs404(t *testing.T) {
	t.Cleanup(tenant.SetScopeResolver(businessScope{}))
	api := NewIntegrationAPI(application.NewService(context.Background(), missingRepo{}))
	ctx := (&authPkg.Device{User: userWith(&authorization.EffectiveAccess{IsAdmin: true})}).Ctx(context.Background())

	_, err := api.DisconnectIntegration(ctx, &hubintegrationv1.DisconnectIntegrationRequest{Id: ptr(int64(99))})
	var statusErr interface{ HTTPStatusCode() int }
	if !errors.As(err, &statusErr) || statusErr.HTTPStatusCode() != http.StatusNotFound {
		t.Fatalf("err = %v, want a 404", err)
	}
}

func TestConnectingAnUnsupportedProviderIs400(t *testing.T) {
	t.Cleanup(tenant.SetScopeResolver(businessScope{}))
	api := NewIntegrationAPI(application.NewService(context.Background(), missingRepo{}))
	ctx := (&authPkg.Device{User: userWith(&authorization.EffectiveAccess{IsAdmin: true})}).Ctx(context.Background())

	_, err := api.ConnectIntegration(ctx, &hubintegrationv1.ConnectIntegrationRequest{Provider: "shopify"})
	var statusErr interface{ HTTPStatusCode() int }
	if !errors.As(err, &statusErr) || statusErr.HTTPStatusCode() != http.StatusBadRequest {
		t.Fatalf("err = %v, want a 400", err)
	}
}

func ptr[T any](v T) *T { return &v }
