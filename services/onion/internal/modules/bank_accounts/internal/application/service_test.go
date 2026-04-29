package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
)

type mockRepo struct {
	listResult []domain.BankAccount
	listErr    error
	createID   int64
	createErr  error
	createGot  struct {
		ownerID int64
		input   domain.BankAccountInput
	}
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.BankAccount, error) {
	return m.listResult, m.listErr
}
func (m *mockRepo) Create(_ context.Context, ownerID int64, input domain.BankAccountInput) (int64, error) {
	m.createGot.ownerID = ownerID
	m.createGot.input = input
	return m.createID, m.createErr
}

func TestList_Delegates(t *testing.T) {
	want := []domain.BankAccount{{ID: 1, Name: "Operating"}}
	s := &Service{repo: &mockRepo{listResult: want}}
	got, err := s.List(context.Background(), 1)
	if err != nil || len(got) != 1 || got[0].Name != "Operating" {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestCreate_Delegates(t *testing.T) {
	m := &mockRepo{createID: 7}
	s := &Service{repo: m}
	id, err := s.Create(context.Background(), 42, domain.BankAccountInput{Name: "x"})
	if err != nil || id != 7 {
		t.Errorf("id=%d err=%v", id, err)
	}
	if m.createGot.ownerID != 42 || m.createGot.input.Name != "x" {
		t.Errorf("repo got %+v", m.createGot)
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("db down")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), 1, domain.BankAccountInput{}); !errors.Is(err, want) {
		t.Errorf("expected %v, got %v", want, err)
	}
}
