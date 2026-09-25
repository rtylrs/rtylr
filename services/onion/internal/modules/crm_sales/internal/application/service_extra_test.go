package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/domain"
)

func TestUpdateContact_HappyPath(t *testing.T) {
	m := &mockRepo{contact: domain.Contact{ID: 3, Name: "Updated"}}
	s := &Service{repo: m}
	got, err := s.UpdateContact(context.Background(), 3, 1, domain.UpdateContactRequest{Name: "Updated"})
	if err != nil || got.Name != "Updated" {
		t.Errorf("got %+v, err %v", got, err)
	}
}

func TestUpdateContact_PropagatesError(t *testing.T) {
	want := errors.New("update contact failed")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.UpdateContact(context.Background(), 3, 1, domain.UpdateContactRequest{})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCreateContact_PropagatesError(t *testing.T) {
	want := errors.New("create contact failed")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.CreateContact(context.Background(), 1, domain.CreateContactRequest{Name: "Alice"})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDeleteContact_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.DeleteContact(context.Background(), 1, 10); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestListTouchpoints_PropagatesError(t *testing.T) {
	want := errors.New("list touchpoints error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.ListTouchpoints(context.Background(), 1, 5)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCreateTouchpoint_PropagatesError(t *testing.T) {
	want := errors.New("create touchpoint error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.CreateTouchpoint(context.Background(), 1, 2, domain.CreateTouchpointRequest{})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDeleteTouchpoint_PropagatesError(t *testing.T) {
	want := errors.New("delete touchpoint error")
	s := &Service{repo: &mockRepo{err: want}}
	if err := s.DeleteTouchpoint(context.Background(), 1, 10); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetPipeline_PropagatesError(t *testing.T) {
	want := errors.New("pipeline error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.GetPipeline(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestListContracts_PropagatesError(t *testing.T) {
	want := errors.New("list contracts error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.ListContracts(context.Background(), 1, 0, 0)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetContract_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{contract: domain.Contract{ID: 7, Plan: "enterprise"}}}
	got, err := s.GetContract(context.Background(), 7, 1)
	if err != nil || got.ID != 7 {
		t.Errorf("got %+v, err %v", got, err)
	}
}

func TestUpdateContract_HappyPath(t *testing.T) {
	m := &mockRepo{contract: domain.Contract{ID: 1, Plan: "starter", Seats: 5}}
	s := &Service{repo: m}
	got, err := s.UpdateContract(context.Background(), 1, 10, domain.UpdateContractRequest{Plan: "starter", Seats: 5})
	if err != nil || got.Plan != "starter" {
		t.Errorf("got %+v, err %v", got, err)
	}
}

func TestCreateContract_SetsDefaultSeats(t *testing.T) {
	m := &mockRepo{contract: domain.Contract{ID: 1, Plan: "pro", Seats: 1}}
	s := &Service{repo: m}
	got, err := s.CreateContract(context.Background(), 1, domain.CreateContractRequest{Seats: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Seats != 1 {
		t.Errorf("expected default seats=1, got %d", got.Seats)
	}
}
