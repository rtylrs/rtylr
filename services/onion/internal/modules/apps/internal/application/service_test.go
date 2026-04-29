package application

import (
	"context"
	"testing"
)

type mockRepo struct {
	enabled map[string]bool
	allowed []string
	gotEnabledBiz int64
	gotAllowed   struct{ biz, uid int64 }
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
