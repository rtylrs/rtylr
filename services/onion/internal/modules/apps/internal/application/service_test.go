package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/platform/authorization"
)

type mockRepo struct {
	enabled     map[string]bool
	enabledErr  error
	businessID  int64
	businessErr error
	lookedUpID  int64
}

func (m *mockRepo) GetEnabledAddons(_ context.Context, _ int64) (map[string]bool, error) {
	return m.enabled, m.enabledErr
}

func (m *mockRepo) GetUserBusinessID(_ context.Context, userID int64) (int64, error) {
	m.lookedUpID = userID
	return m.businessID, m.businessErr
}

func TestListApps_NonEmpty(t *testing.T) {
	if got := (&Service{}).ListApps(); len(got) == 0 {
		t.Fatal("expected app catalog")
	}
}

func TestGetVisibleApps_AnonymousOnlySeesUngated(t *testing.T) {
	service := &Service{repo: &mockRepo{}}
	visible := service.GetVisibleApps(context.Background(), 0, 0, nil)
	if len(visible) == 0 {
		t.Fatal("expected ungated apps for anonymous callers")
	}
	for _, app := range visible {
		if app.AddonGated != "" {
			t.Errorf("anonymous caller saw addon-gated app %q", app.ID)
		}
	}
}

func TestGetVisibleAppsForUser_RequiresResolvedLicenseAndProfileAccess(t *testing.T) {
	service := &Service{repo: &mockRepo{enabled: map[string]bool{}}}
	access := &authorization.EffectiveAccess{BusinessID: 11, Apps: []authorization.EffectiveAppAccess{
		{AppID: authorization.AppPOS, Licensed: true, HasAccess: true},
		{AppID: authorization.AppCRM, Licensed: false, HasAccess: true},
		{AppID: authorization.AppFinance, Licensed: true, HasAccess: false},
	}}

	visible, err := service.GetVisibleAppsForUser(context.Background(), 11, 42, access)
	if err != nil {
		t.Fatalf("GetVisibleAppsForUser: %v", err)
	}
	if len(visible) != 1 || visible[0].ID != "pos" {
		t.Fatalf("visible apps = %#v, want only the licensed POS app with a profile", visible)
	}
}

func TestGetVisibleAppsForUser_ResolvesBusinessFromUserWhenMissing(t *testing.T) {
	repo := &mockRepo{businessID: 7, enabled: map[string]bool{}}
	service := &Service{repo: repo}
	access := &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
		AppID: authorization.AppPOS, Licensed: true, HasAccess: true,
	}}}

	visible, err := service.GetVisibleAppsForUser(context.Background(), 0, 5, access)
	if err != nil {
		t.Fatalf("GetVisibleAppsForUser: %v", err)
	}
	if repo.lookedUpID != 5 {
		t.Fatalf("business lookup user id = %d, want 5", repo.lookedUpID)
	}
	if len(visible) != 1 || visible[0].ID != "pos" {
		t.Fatalf("visible apps = %#v, want POS", visible)
	}
}

func TestGetVisibleAppsForUser_PropagatesBusinessLookupError(t *testing.T) {
	wantErr := errors.New("business lookup failed")
	service := &Service{repo: &mockRepo{businessErr: wantErr}}
	if _, err := service.GetVisibleAppsForUser(context.Background(), 0, 5, nil); !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func TestGetVisibleAppsForUser_NoEffectiveAccessReturnsNoApps(t *testing.T) {
	service := &Service{repo: &mockRepo{enabled: map[string]bool{}}}
	visible, err := service.GetVisibleAppsForUser(context.Background(), 1, 5, nil)
	if err != nil {
		t.Fatalf("GetVisibleAppsForUser: %v", err)
	}
	if len(visible) != 0 {
		t.Fatalf("visible apps = %#v, want none", visible)
	}
}
