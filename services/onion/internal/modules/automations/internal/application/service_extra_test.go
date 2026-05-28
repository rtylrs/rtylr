package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
)

func TestList_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{listResult: []domain.Automation{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestCreate_ZeroOwnerID(t *testing.T) {
	m := &mockRepo{createID: 10}
	s := &Service{repo: m}
	id, err := s.Create(context.Background(), domain.AutomationInput{OwnerID: 0, Name: "Zero Owner"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 10 {
		t.Errorf("expected id=10, got %d", id)
	}
}

func TestCreate_PassesNameThrough(t *testing.T) {
	m := &mockRepo{createID: 1}
	s := &Service{repo: m}
	_, _ = s.Create(context.Background(), domain.AutomationInput{OwnerID: 1, Name: "Birthday Automation"})
	if m.createGot.Name != "Birthday Automation" {
		t.Errorf("Name = %q, want 'Birthday Automation'", m.createGot.Name)
	}
}

func TestCreate_PassesTriggerTypeThrough(t *testing.T) {
	m := &mockRepo{createID: 1}
	s := &Service{repo: m}
	_, _ = s.Create(context.Background(), domain.AutomationInput{
		OwnerID:     1,
		Name:        "Trigger Test",
		TriggerType: "customer_signup",
	})
	if m.createGot.TriggerType != "customer_signup" {
		t.Errorf("TriggerType = %q, want 'customer_signup'", m.createGot.TriggerType)
	}
}

func TestUpdate_PassesIDThrough(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	_ = s.Update(context.Background(), domain.AutomationUpdateInput{ID: 99, OwnerID: 1, Name: "Updated"})
	if m.updateGot.ID != 99 {
		t.Errorf("update ID = %d, want 99", m.updateGot.ID)
	}
}

func TestDelete_PassesOwnerAndIDThrough(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	_ = s.Delete(context.Background(), 5, 55)
	if m.deleteGot.owner != 5 || m.deleteGot.id != 55 {
		t.Errorf("delete got owner=%d id=%d, want 5/55", m.deleteGot.owner, m.deleteGot.id)
	}
}

func TestList_MultipleItems(t *testing.T) {
	items := []domain.Automation{
		{ID: 1, Name: "A"},
		{ID: 2, Name: "B"},
		{ID: 3, Name: "C"},
	}
	s := &Service{repo: &mockRepo{listResult: items}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 items, got %d", len(got))
	}
}

func TestCreate_IsActiveIsPassedThrough(t *testing.T) {
	m := &mockRepo{createID: 1}
	s := &Service{repo: m}
	_, _ = s.Create(context.Background(), domain.AutomationInput{
		OwnerID:  1,
		Name:     "Active Test",
		IsActive: true,
	})
	if !m.createGot.IsActive {
		t.Error("expected IsActive=true to pass through to repo")
	}
}

func TestUpdate_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.Update(context.Background(), domain.AutomationUpdateInput{
		ID:      1,
		OwnerID: 10,
		Name:    "Updated Name",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.Delete(context.Background(), 1, 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
