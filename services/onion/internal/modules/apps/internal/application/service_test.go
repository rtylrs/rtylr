package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

type mockRepo struct {
	enabled        map[string]bool
	enabledErr     error
	allowed        []string
	allowedErr     error
	businessID     int64
	businessErr    error
	gotEnabledBiz  int64
	gotBusinessUID int64
	gotAllowed     struct{ biz, uid int64 }
}

func (m *mockRepo) GetEnabledAddons(_ context.Context, businessID int64) (map[string]bool, error) {
	m.gotEnabledBiz = businessID
	return m.enabled, m.enabledErr
}

func (m *mockRepo) GetUserAllowedApps(_ context.Context, biz, uid int64) ([]string, error) {
	m.gotAllowed.biz = biz
	m.gotAllowed.uid = uid
	return m.allowed, m.allowedErr
}

func (m *mockRepo) GetUserBusinessID(_ context.Context, uid int64) (int64, error) {
	m.gotBusinessUID = uid
	if m.businessErr != nil {
		return 0, m.businessErr
	}
	return m.businessID, nil
}

func TestListApps_NonEmpty(t *testing.T) {
	if got := (&Service{}).ListApps(); len(got) == 0 {
		t.Error("expected ListApps to return at least one app")
	}
}

func TestGetEnabledAddons_Delegates(t *testing.T) {
	m := &mockRepo{enabled: map[string]bool{"x": true}}
	s := &Service{repo: m}
	if got := s.GetEnabledAddons(context.Background(), 7); !got["x"] {
		t.Errorf("expected x=true, got %v", got)
	}
	if m.gotEnabledBiz != 7 {
		t.Errorf("biz id didn't pass through, got %d", m.gotEnabledBiz)
	}
}

func TestGetUserAllowedApps_Delegates(t *testing.T) {
	m := &mockRepo{allowed: []string{"hub", "pos"}}
	s := &Service{repo: m}
	got := s.GetUserAllowedApps(context.Background(), 7, 42)
	if len(got) != 2 || got[0] != "hub" {
		t.Errorf("got %v", got)
	}
	if m.gotAllowed.biz != 7 || m.gotAllowed.uid != 42 {
		t.Errorf("ids didn't pass through: %+v", m.gotAllowed)
	}
}

func TestGetVisibleApps_AnonymousOnlySeesUngated(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	// userID=0 means anonymous
	got := s.GetVisibleApps(context.Background(), 0, 0, "")
	for _, a := range got {
		if a.AddonGated != "" {
			t.Errorf("anonymous should not see addon-gated app %s", a.ID)
		}
	}
	if len(got) == 0 {
		t.Error("expected at least one ungated app for anonymous")
	}
}

func TestGetVisibleApps_OwnerSeesAll(t *testing.T) {
	s := &Service{repo: &mockRepo{enabled: map[string]bool{}}}
	all := s.ListApps()
	visible := s.GetVisibleApps(context.Background(), 1, 1, "owner")
	if len(visible) != len(all) {
		t.Errorf("owner should see all %d apps, got %d", len(all), len(visible))
	}
}

func TestGetVisibleApps_AdminSeesAll(t *testing.T) {
	s := &Service{repo: &mockRepo{enabled: map[string]bool{}}}
	all := s.ListApps()
	visible := s.GetVisibleApps(context.Background(), 1, 2, "admin")
	if len(visible) != len(all) {
		t.Errorf("admin should see all %d apps, got %d", len(all), len(visible))
	}
}

func TestGetVisibleApps_ConsoleAdminSeesAll(t *testing.T) {
	s := &Service{repo: &mockRepo{enabled: map[string]bool{}}}
	all := s.ListApps()
	visible := s.GetVisibleApps(context.Background(), 1, 2, "console_admin")
	if len(visible) != len(all) {
		t.Errorf("console admin should see all %d apps, got %d", len(all), len(visible))
	}
}

func TestGetVisibleApps_RegularUserSeesOnlyAllowed(t *testing.T) {
	m := &mockRepo{
		enabled: map[string]bool{},
		allowed: []string{"pos", "finance"},
	}
	s := &Service{repo: m}
	visible := s.GetVisibleApps(context.Background(), 1, 5, "staff")
	if len(visible) != 2 {
		t.Errorf("expected 2 allowed apps, got %d", len(visible))
	}
	for _, a := range visible {
		if a.ID != "pos" && a.ID != "finance" {
			t.Errorf("unexpected app %s for staff user", a.ID)
		}
	}
}

func TestGetVisibleApps_RegularUserSeesNoAppsIfNotAllowed(t *testing.T) {
	m := &mockRepo{
		enabled: map[string]bool{},
		allowed: []string{},
	}
	s := &Service{repo: m}
	visible := s.GetVisibleApps(context.Background(), 1, 5, "staff")
	if len(visible) != 0 {
		t.Errorf("expected no apps for staff with no permissions, got %d", len(visible))
	}
}

func TestGetVisibleAppsForUser_ResolvesMissingBusinessID(t *testing.T) {
	m := &mockRepo{
		enabled:    map[string]bool{},
		allowed:    []string{"pos", "finance"},
		businessID: 7,
	}
	s := &Service{repo: m}

	visible, err := s.GetVisibleAppsForUser(context.Background(), 0, 5, "staff")
	if err != nil {
		t.Fatalf("GetVisibleAppsForUser returned error: %v", err)
	}
	if m.gotBusinessUID != 5 {
		t.Fatalf("expected business lookup for user 5, got %d", m.gotBusinessUID)
	}
	if m.gotAllowed.biz != 7 || m.gotAllowed.uid != 5 {
		t.Fatalf("expected allowed apps lookup for business 7/user 5, got %+v", m.gotAllowed)
	}
	if len(visible) != 2 {
		t.Fatalf("expected 2 visible apps, got %d", len(visible))
	}
}

func TestGetVisibleAppsForUser_AnonymousDoesNotResolveBusinessID(t *testing.T) {
	m := &mockRepo{enabled: map[string]bool{}}
	s := &Service{repo: m}

	visible, err := s.GetVisibleAppsForUser(context.Background(), 0, 0, "")
	if err != nil {
		t.Fatalf("GetVisibleAppsForUser returned error: %v", err)
	}
	if m.gotBusinessUID != 0 {
		t.Fatalf("anonymous calls should not look up user business, got %d", m.gotBusinessUID)
	}
	if len(visible) == 0 {
		t.Fatal("expected anonymous visible apps")
	}
}

func TestGetVisibleAppsForUser_PropagatesBusinessLookupError(t *testing.T) {
	wantErr := errors.New("business lookup failed")
	m := &mockRepo{businessErr: wantErr}
	s := &Service{repo: m}

	_, err := s.GetVisibleAppsForUser(context.Background(), 0, 5, "staff")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestGetEnabledAddons_EmptyMap(t *testing.T) {
	m := &mockRepo{enabled: map[string]bool{}}
	s := &Service{repo: m}
	got := s.GetEnabledAddons(context.Background(), 1)
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestGetUserAllowedApps_Empty(t *testing.T) {
	m := &mockRepo{allowed: []string{}}
	s := &Service{repo: m}
	got := s.GetUserAllowedApps(context.Background(), 1, 5)
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

// Regression: a transient repository failure used to be handed to the cacher
// with a nil error, so the fail-closed empty result was cached for the full
// 5-minute TTL and outlived the incident. The fetch closures now propagate the
// error, which stops the cacher from storing the empty value.
//
// A real in-memory cache is required to observe the difference: without one
// every call re-fetches and the bug is invisible.
func TestGetEnabledAddons_FailureIsNotCached(t *testing.T) {
	if err := cacher.InitializeLocalCache(context.Background()); err != nil {
		t.Fatalf("InitializeLocalCache: %v", err)
	}

	const businessID int64 = 987001
	m := &mockRepo{enabledErr: errors.New("transient db failure")}
	s := &Service{repo: m}

	if got := s.GetEnabledAddons(context.Background(), businessID); len(got) != 0 {
		t.Fatalf("failed lookup should fail closed to an empty map, got %v", got)
	}

	m.enabled, m.enabledErr = map[string]bool{"crm": true}, nil
	if got := s.GetEnabledAddons(context.Background(), businessID); !got["crm"] {
		t.Fatalf("recovered lookup still saw the cached empty map: %v", got)
	}
}

func TestGetUserAllowedApps_FailureIsNotCached(t *testing.T) {
	if err := cacher.InitializeLocalCache(context.Background()); err != nil {
		t.Fatalf("InitializeLocalCache: %v", err)
	}

	const businessID int64 = 987002
	const userID int64 = 987003
	m := &mockRepo{allowedErr: errors.New("transient db failure")}
	s := &Service{repo: m}

	if got := s.GetUserAllowedApps(context.Background(), businessID, userID); len(got) != 0 {
		t.Fatalf("failed lookup should fail closed to an empty list, got %v", got)
	}

	m.allowed, m.allowedErr = []string{"hub"}, nil
	if got := s.GetUserAllowedApps(context.Background(), businessID, userID); len(got) != 1 || got[0] != "hub" {
		t.Fatalf("recovered lookup still saw the cached empty list: %v", got)
	}
}
