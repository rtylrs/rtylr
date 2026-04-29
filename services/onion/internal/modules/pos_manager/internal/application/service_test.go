package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/domain"
)

type mockRepo struct {
	result *domain.VerifyResult
	err    error
}

func (m *mockRepo) VerifyPIN(_ context.Context, _ int64, _ string) (*domain.VerifyResult, error) {
	return m.result, m.err
}

func TestVerifyPIN_Delegates(t *testing.T) {
	want := &domain.VerifyResult{}
	s := &Service{repo: &mockRepo{result: want}}
	got, err := s.VerifyPIN(context.Background(), 1, "1234")
	if err != nil || got != want {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestVerifyPIN_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{err: want}}
	if _, err := s.VerifyPIN(context.Background(), 1, ""); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
