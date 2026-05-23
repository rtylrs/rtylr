package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
)

type mockRepo struct {
	listResult []domain.Automation
	listErr    error
	createID   int64
	createErr  error
	updateErr  error
	deleteErr  error
	createGot  domain.AutomationInput
	updateGot  domain.AutomationUpdateInput
	deleteGot  struct{ owner, id int64 }
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Automation, error) {
	return m.listResult, m.listErr
}

func (m *mockRepo) Create(_ context.Context, input domain.AutomationInput) (int64, error) {
	m.createGot = input
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, input domain.AutomationUpdateInput) error {
	m.updateGot = input
	return m.updateErr
}

func (m *mockRepo) Delete(_ context.Context, owner, id int64) error {
	m.deleteGot.owner = owner
	m.deleteGot.id = id
	return m.deleteErr
}

func TestList_Delegates(t *testing.T) {
	want := []domain.Automation{{ID: 1}}
	s := &Service{repo: &mockRepo{listResult: want}}
	got, err := s.List(context.Background(), 1)
	if err != nil || len(got) != 1 {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), domain.AutomationInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdate_Delegates(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	if err := s.Update(context.Background(), domain.AutomationUpdateInput{ID: 5}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if m.updateGot.ID != 5 {
		t.Errorf("update got %+v", m.updateGot)
	}
}

func TestDelete_Delegates(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	if err := s.Delete(context.Background(), 7, 42); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if m.deleteGot.owner != 7 || m.deleteGot.id != 42 {
		t.Errorf("delete got %+v", m.deleteGot)
	}
}
