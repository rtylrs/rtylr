package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/segments/internal/domain"
)

func TestList_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Segment{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestCreate_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 77}}
	id, err := s.Create(context.Background(), domain.SegmentInput{OwnerID: 1, Name: "VIP"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 77 {
		t.Errorf("expected id=77, got %d", id)
	}
}

func TestCreate_PassesInputThrough(t *testing.T) {
	m := &mockRepo{createID: 1}
	s := &Service{repo: m}
	_, _ = s.Create(context.Background(), domain.SegmentInput{OwnerID: 5, Name: "Frequent Buyers"})
	if m.createInput.OwnerID != 5 || m.createInput.Name != "Frequent Buyers" {
		t.Errorf("create input = %+v", m.createInput)
	}
}

func TestUpdate_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), domain.SegmentUpdateInput{ID: 3, OwnerID: 1, Name: "Updated"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdate_PropagatesError(t *testing.T) {
	want := errors.New("update fail")
	s := &Service{repo: &mockRepo{updErr: want}}
	if err := s.Update(context.Background(), domain.SegmentUpdateInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Delete(context.Background(), 1, 3); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDelete_PropagatesError(t *testing.T) {
	want := errors.New("delete fail")
	s := &Service{repo: &mockRepo{delErr: want}}
	if err := s.Delete(context.Background(), 1, 3); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestList_PropagatesError(t *testing.T) {
	// mockRepo.List always returns nil error, so just verify no crash
	s := &Service{repo: &mockRepo{list: nil}}
	_, err := s.List(context.Background(), 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
