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

func TestVerifyPIN_ValidResult(t *testing.T) {
	want := &domain.VerifyResult{Valid: true, StaffID: 42, Name: "Alice"}
	s := &Service{repo: &mockRepo{result: want}}
	got, err := s.VerifyPIN(context.Background(), 1, "1234")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Valid || got.StaffID != 42 || got.Name != "Alice" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestVerifyPIN_InvalidResult(t *testing.T) {
	want := &domain.VerifyResult{Valid: false}
	s := &Service{repo: &mockRepo{result: want}}
	got, err := s.VerifyPIN(context.Background(), 1, "9999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Valid {
		t.Error("expected Valid=false")
	}
}

func TestVerifyPIN_ErrInvalidPINPropagated(t *testing.T) {
	s := &Service{repo: &mockRepo{err: domain.ErrInvalidPIN}}
	_, err := s.VerifyPIN(context.Background(), 1, "0000")
	if !errors.Is(err, domain.ErrInvalidPIN) {
		t.Errorf("expected ErrInvalidPIN, got %v", err)
	}
}

func TestVerifyPIN_NilResultOnError(t *testing.T) {
	s := &Service{repo: &mockRepo{err: errors.New("any error")}}
	got, _ := s.VerifyPIN(context.Background(), 1, "1234")
	if got != nil {
		t.Errorf("expected nil result on error, got %+v", got)
	}
}

func TestVerifyPIN_EmptyPIN(t *testing.T) {
	want := &domain.VerifyResult{Valid: false}
	s := &Service{repo: &mockRepo{result: want}}
	got, err := s.VerifyPIN(context.Background(), 1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Valid {
		t.Error("empty PIN should return Valid=false")
	}
}

func TestVerifyPIN_DifferentOwners(t *testing.T) {
	want := &domain.VerifyResult{Valid: true, StaffID: 10}
	s := &Service{repo: &mockRepo{result: want}}
	got, err := s.VerifyPIN(context.Background(), 99, "5678")
	if err != nil || !got.Valid {
		t.Errorf("unexpected result for ownerID 99: %v, %v", got, err)
	}
}

func TestVerifyPIN_StaffIDInResult(t *testing.T) {
	s := &Service{repo: &mockRepo{result: &domain.VerifyResult{Valid: true, StaffID: 77}}}
	got, _ := s.VerifyPIN(context.Background(), 1, "5555")
	if got.StaffID != 77 {
		t.Errorf("expected StaffID 77, got %d", got.StaffID)
	}
}

func TestVerifyPIN_NameInResult(t *testing.T) {
	s := &Service{repo: &mockRepo{result: &domain.VerifyResult{Valid: true, Name: "Bob"}}}
	got, _ := s.VerifyPIN(context.Background(), 1, "1111")
	if got.Name != "Bob" {
		t.Errorf("expected Name Bob, got %q", got.Name)
	}
}

func TestVerifyPIN_MultipleCallsReturnSameResult(t *testing.T) {
	want := &domain.VerifyResult{Valid: true, StaffID: 5}
	s := &Service{repo: &mockRepo{result: want}}
	for i := 0; i < 3; i++ {
		got, err := s.VerifyPIN(context.Background(), 1, "1234")
		if err != nil || got != want {
			t.Errorf("call %d: unexpected result %v, %v", i, got, err)
		}
	}
}

func TestVerifyPIN_ErrInvalidPINIsDistinct(t *testing.T) {
	if domain.ErrInvalidPIN == nil {
		t.Fatal("ErrInvalidPIN should not be nil")
	}
	if domain.ErrInvalidPIN.Error() == "" {
		t.Error("ErrInvalidPIN should have a message")
	}
}
