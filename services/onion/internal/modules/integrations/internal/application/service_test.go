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
