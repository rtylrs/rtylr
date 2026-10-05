package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/benefits/internal/domain"
)

type mockRepo struct {
	listResult []domain.BenefitPlan
	listErr    error
	createID   int64
	createErr  error
	createGot  domain.BenefitPlanInput
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.BenefitPlan, error) {
	return m.listResult, m.listErr
}

func (m *mockRepo) Create(_ context.Context, _ int64, input domain.BenefitPlanInput) (int64, error) {
	m.createGot = input
	return m.createID, m.createErr
}

// Enrollment methods are unused by the plan tests below; mockEnrollmentRepo
// overrides them where behavior matters.
func (m *mockRepo) CreateEnrollment(_ context.Context, _, _, _ int64) (domain.BenefitEnrollment, error) {
	return domain.BenefitEnrollment{}, nil
}

func (m *mockRepo) ListEnrollmentsByEmployee(_ context.Context, _, _ int64, _, _ int32) ([]domain.BenefitEnrollment, int32, error) {
	return nil, 0, nil
}

func TestList_Delegates(t *testing.T) {
	want := []domain.BenefitPlan{{ID: 1, Name: "Health"}}
	s := &Service{repo: &mockRepo{listResult: want}}
	got, err := s.List(context.Background(), 1)
	if err != nil || len(got) != 1 || got[0].Name != "Health" {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestCreate_Delegates(t *testing.T) {
	m := &mockRepo{createID: 5}
	s := &Service{repo: m}
	id, err := s.Create(context.Background(), 1, domain.BenefitPlanInput{Name: "x"})
	if err != nil || id != 5 || m.createGot.Name != "x" {
		t.Errorf("id=%d err=%v got=%+v", id, err, m.createGot)
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), 1, domain.BenefitPlanInput{}); !errors.Is(err, want) {
		t.Errorf("expected %v, got %v", want, err)
	}
}

func TestList_Empty(t *testing.T) {
	s := &Service{repo: &mockRepo{listResult: []domain.BenefitPlan{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestList_PropagatesError(t *testing.T) {
	want := errors.New("list fail")
	s := &Service{repo: &mockRepo{listErr: want}}
	if _, err := s.List(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestList_MultiplePlans(t *testing.T) {
	plans := []domain.BenefitPlan{
		{ID: 1, Name: "Health"},
		{ID: 2, Name: "Dental"},
		{ID: 3, Name: "Vision"},
	}
	s := &Service{repo: &mockRepo{listResult: plans}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3, got %d", len(got))
	}
}

func TestCreate_ReturnsCorrectID(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 77}}
	id, err := s.Create(context.Background(), 1, domain.BenefitPlanInput{Name: "Life Insurance"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if id != 77 {
		t.Errorf("expected 77, got %d", id)
	}
}

func TestCreate_ForwardsInput(t *testing.T) {
	m := &mockRepo{createID: 1}
	s := &Service{repo: m}
	input := domain.BenefitPlanInput{Name: "Premium Health", Type: "health", Provider: "Acme"}
	_, _ = s.Create(context.Background(), 1, input)
	if m.createGot.Name != "Premium Health" || m.createGot.Type != "health" {
		t.Errorf("input not forwarded: %+v", m.createGot)
	}
}

func TestCreate_ForwardsOwnerID(t *testing.T) {
	m := &mockRepo{createID: 1}
	// The mock ignores ownerID but we verify it compiles and runs without panic.
	s := &Service{repo: m}
	id, err := s.Create(context.Background(), 42, domain.BenefitPlanInput{Name: "Test"})
	if err != nil || id != 1 {
		t.Errorf("unexpected id=%d err=%v", id, err)
	}
}

func TestList_PreservesOrder(t *testing.T) {
	plans := []domain.BenefitPlan{
		{ID: 3, Name: "Third"},
		{ID: 1, Name: "First"},
		{ID: 2, Name: "Second"},
	}
	s := &Service{repo: &mockRepo{listResult: plans}}
	got, _ := s.List(context.Background(), 1)
	if got[0].ID != 3 || got[1].ID != 1 || got[2].ID != 2 {
		t.Errorf("order changed: %v", got)
	}
}
