package api

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/hubdeveloper"
	"github.com/voxire/onion/internal/platform/tenant"
	hubdeveloperv1 "github.com/voxire/voxie/gen/go/hub/developer/v1"
)

type rotatingWebhooks struct {
	hubdeveloper.WebhooksProvider
	rotatedID int64
}

func (r *rotatingWebhooks) RotateSecret(_ context.Context, _, webhookID int64) (string, error) {
	r.rotatedID = webhookID
	return "new-secret", nil
}

type ownBusiness struct{}

func (ownBusiness) ActiveBusinessOwner(context.Context, int64) (int64, error) { return 1, nil }
func (ownBusiness) UserBusiness(_ context.Context, businessID, _ int64) (int64, error) {
	return businessID, nil
}

func TestRotateWebhookSecretNeedsDeveloperAccessAndRevealsTheSecretOnce(t *testing.T) {
	t.Cleanup(tenant.SetScopeResolver(ownBusiness{}))
	provider := &rotatingWebhooks{}
	api := &DeveloperAPI{WebhooksService: provider}

	member := (&authPkg.Device{User: &authPkg.AuthUser{UserID: 5, BusinessID: 1, EffectiveAccess: &authorization.EffectiveAccess{}}}).Ctx(context.Background())
	if _, err := api.RotateWebhookSecret(member, &hubdeveloperv1.RotateWebhookSecretRequest{Id: 3}); err == nil || provider.rotatedID != 0 {
		t.Fatalf("a member rotated a signing secret: err=%v id=%d", err, provider.rotatedID)
	}

	admin := (&authPkg.Device{User: &authPkg.AuthUser{UserID: 6, BusinessID: 1, EffectiveAccess: &authorization.EffectiveAccess{IsAdmin: true}}}).Ctx(context.Background())
	got, err := api.RotateWebhookSecret(admin, &hubdeveloperv1.RotateWebhookSecretRequest{Id: 3})
	if err != nil || got.GetSecret() != "new-secret" || provider.rotatedID != 3 {
		t.Fatalf("admin rotate = %v, %v (id %d)", got, err, provider.rotatedID)
	}
}
