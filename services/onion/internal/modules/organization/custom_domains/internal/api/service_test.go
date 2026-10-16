package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/custom_domains/internal/application"
	"github.com/voxire/onion/internal/modules/organization/custom_domains/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/tenant"
	hubconfigv1 "github.com/voxire/voxie/gen/go/hub/config/v1"
)

func TestPayloadVerificationTokenVisibility(t *testing.T) {
	item := domain.CustomDomain{ID: 7, BusinessID: 42, Domain: "menu.example.com", VerificationToken: "challenge-token"}

	if got := payload(item, true).GetVerificationToken(); got != item.VerificationToken {
		t.Fatalf("registration payload token = %q, want %q", got, item.VerificationToken)
	}
	if got := payload(item, false).GetVerificationToken(); got != "" {
		t.Fatalf("non-registration payload token = %q, want it redacted", got)
	}
}

type resolvingScope struct{}

func (resolvingScope) ActiveBusinessOwner(context.Context, int64) (int64, error) { return 1, nil }
func (resolvingScope) UserBusiness(_ context.Context, businessID, _ int64) (int64, error) {
	return businessID, nil
}

type removeOnlyRepo struct{ removed bool }

func (r *removeOnlyRepo) Get(context.Context, int64) (domain.CustomDomain, error) {
	return domain.CustomDomain{}, domain.ErrNotFound
}
func (r *removeOnlyRepo) GetByDomain(context.Context, string) (domain.CustomDomain, error) {
	return domain.CustomDomain{}, domain.ErrNotFound
}
func (r *removeOnlyRepo) Register(context.Context, domain.CustomDomain) (domain.CustomDomain, error) {
	return domain.CustomDomain{}, nil
}
func (r *removeOnlyRepo) Remove(context.Context, int64) error         { r.removed = true; return nil }
func (r *removeOnlyRepo) Verify(context.Context, int64, string) error { return nil }

func TestDomainMutationsRequireBusinessAdmin(t *testing.T) {
	t.Cleanup(tenant.SetScopeResolver(resolvingScope{}))
	repo := &removeOnlyRepo{}
	api := New(application.NewService(context.Background(), repo))

	memberCtx := (&authPkg.Device{User: &authPkg.AuthUser{UserID: 5, BusinessID: 1, EffectiveAccess: &authorization.EffectiveAccess{}}}).Ctx(context.Background())
	if _, err := api.RegisterCustomDomain(memberCtx, &hubconfigv1.RegisterCustomDomainRequest{Domain: "menu.example.com"}); err == nil {
		t.Fatal("a non-admin member registered a custom domain")
	}
	if _, err := api.VerifyCustomDomain(memberCtx, &hubconfigv1.VerifyCustomDomainRequest{VerificationToken: "t"}); err == nil {
		t.Fatal("a non-admin member verified a custom domain")
	}
	if _, err := api.RemoveCustomDomain(memberCtx, &hubconfigv1.RemoveCustomDomainRequest{}); err == nil || repo.removed {
		t.Fatal("a non-admin member removed a custom domain")
	}

	adminCtx := (&authPkg.Device{User: &authPkg.AuthUser{UserID: 6, BusinessID: 1, EffectiveAccess: &authorization.EffectiveAccess{IsAdmin: true}}}).Ctx(context.Background())
	if _, err := api.RemoveCustomDomain(adminCtx, &hubconfigv1.RemoveCustomDomainRequest{}); err != nil || !repo.removed {
		t.Fatalf("admin remove failed: err=%v removed=%v", err, repo.removed)
	}
}
