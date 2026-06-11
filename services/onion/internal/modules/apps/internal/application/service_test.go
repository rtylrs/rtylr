package application

import (
	"context"
	"testing"
)

type mockRepo struct {
	enabled       map[string]bool
	allowed       []string
	gotEnabledBiz int64
	gotAllowed    struct{ biz, uid int64 }
}

func (m *mockRepo) GetEnabledAddons(_ context.Context, businessID int64) map[string]bool {
	m.gotEnabledBiz = businessID
	return m.enabled
}

func (m *mockRepo) GetUserAllowedApps(_ context.Context, biz, uid int64) []string {
	m.gotAllowed.biz = biz
	m.gotAllowed.uid = uid
	return m.allowed
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
