package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/integrations/internal/domain"
)

type mockRepo struct {
	listResult []domain.Integration
	listErr    error
	connectID  int64
	connectErr error
	disconnErr error
	connectGot struct {
		ownerID  int64
		provider string
		config   json.RawMessage
	}
	disconnGot struct {
		ownerID, id int64
	}
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Integration, error) {
	return m.listResult, m.listErr
}

func (m *mockRepo) Connect(_ context.Context, ownerID int64, provider string, config json.RawMessage) (int64, error) {
	m.connectGot.ownerID = ownerID
	m.connectGot.provider = provider
	m.connectGot.config = config
	return m.connectID, m.connectErr
}

func (m *mockRepo) Disconnect(_ context.Context, ownerID, id int64) error {
	m.disconnGot.ownerID = ownerID
	m.disconnGot.id = id
	return m.disconnErr
}

func TestList_Delegates(t *testing.T) {
	want := []domain.Integration{{ID: 1, Provider: "stripe"}}
	s := &Service{repo: &mockRepo{listResult: want}}
	got, err := s.List(context.Background(), 1)
	if err != nil || len(got) != 1 || got[0].Provider != "stripe" {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestConnect_RequiresProvider(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.Connect(context.Background(), 1, "", nil); err == nil {
		t.Fatalf("expected error for empty provider")
	}
}

func TestConnect_NilConfigBecomesEmptyObject(t *testing.T) {
	m := &mockRepo{connectID: 5}
	s := &Service{repo: m}
	if _, err := s.Connect(context.Background(), 7, "stripe", nil); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if string(m.connectGot.config) != "{}" {
		t.Errorf("expected nil config to become {}, got %q", m.connectGot.config)
	}
}

func TestConnect_PassesThroughConfig(t *testing.T) {
	m := &mockRepo{connectID: 9}
	s := &Service{repo: m}
	cfg := json.RawMessage(`{"key":"value"}`)
	id, _ := s.Connect(context.Background(), 7, "slack", cfg)
	if id != 9 {
		t.Errorf("expected id 9, got %d", id)
	}
	if string(m.connectGot.config) != string(cfg) {
		t.Errorf("config didn't pass through: %s", m.connectGot.config)
	}
}

func TestConnect_PropagatesRepoError(t *testing.T) {
	want := errors.New("connect failed")
	s := &Service{repo: &mockRepo{connectErr: want}}
	if _, err := s.Connect(context.Background(), 1, "x", nil); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDisconnect_RequiresIntegrationID(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Disconnect(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected error for missing integration id")
	}
}

func TestDisconnect_DelegatesIDs(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	if err := s.Disconnect(context.Background(), 7, 42); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if m.disconnGot.ownerID != 7 || m.disconnGot.id != 42 {
		t.Errorf("disconnect args: %+v", m.disconnGot)
	}
}

func TestDisconnect_PropagatesRepoError(t *testing.T) {
	want := errors.New("disconnect failed")
	s := &Service{repo: &mockRepo{disconnErr: want}}
	if err := s.Disconnect(context.Background(), 1, 10); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestList_PropagatesError(t *testing.T) {
	want := errors.New("list failed")
	s := &Service{repo: &mockRepo{listErr: want}}
	_, err := s.List(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestList_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{listResult: []domain.Integration{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d", len(got))
	}
}

func TestList_ReturnsMultiple(t *testing.T) {
	integrations := []domain.Integration{
		{ID: 1, Provider: "stripe"},
		{ID: 2, Provider: "slack"},
	}
	s := &Service{repo: &mockRepo{listResult: integrations}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2, got %d", len(got))
	}
}

func TestConnect_DelegatesOwnerID(t *testing.T) {
	m := &mockRepo{connectID: 1}
	s := &Service{repo: m}
	_, _ = s.Connect(context.Background(), 55, "stripe", nil)
	if m.connectGot.ownerID != 55 {
		t.Errorf("expected ownerID 55, got %d", m.connectGot.ownerID)
	}
}

func TestConnect_DelegatesProvider(t *testing.T) {
	m := &mockRepo{connectID: 1}
	s := &Service{repo: m}
	_, _ = s.Connect(context.Background(), 1, "shopify", nil)
	if m.connectGot.provider != "shopify" {
		t.Errorf("expected provider shopify, got %q", m.connectGot.provider)
	}
}

func TestDisconnect_ZeroIntegrationIDError(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.Disconnect(context.Background(), 1, 0)
	if err == nil {
		t.Fatal("expected error for zero integration ID")
	}
}

func TestConnect_ReturnsGeneratedID(t *testing.T) {
	m := &mockRepo{connectID: 77}
	s := &Service{repo: m}
	id, err := s.Connect(context.Background(), 1, "hubspot", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 77 {
		t.Errorf("expected 77, got %d", id)
	}
}
