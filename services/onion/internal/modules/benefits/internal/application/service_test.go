package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/benefits/internal/domain"
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
