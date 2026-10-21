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
	id, _ := s.Connect(context.Background(), 7, "stripe", cfg)
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
	if _, err := s.Connect(context.Background(), 1, "stripe", nil); !errors.Is(err, want) {
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
	_, _ = s.Connect(context.Background(), 1, " Stripe ", nil)
	if m.connectGot.provider != "stripe" {
		t.Errorf("expected provider stripe, got %q", m.connectGot.provider)
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
	id, err := s.Connect(context.Background(), 1, "stripe", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 77 {
		t.Errorf("expected 77, got %d", id)
	}
}

func TestConnect_RejectsUnknownProviders(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	for _, provider := range []string{"", "shopify", "stripe/../x", "<script>"} {
		if _, err := s.Connect(context.Background(), 1, provider, nil); !errors.Is(err, domain.ErrInvalidProvider) {
			t.Errorf("provider %q: err = %v, want ErrInvalidProvider", provider, err)
		}
	}
	if m.connectGot.provider != "" {
		t.Fatal("an unsupported provider reached the repository")
	}
}

func TestConnect_RejectsNonObjectOrOversizedConfig(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	for name, config := range map[string]json.RawMessage{
		"array":   json.RawMessage(`[1]`),
		"string":  json.RawMessage(`"x"`),
		"null":    json.RawMessage(`null`),
		"garbage": json.RawMessage(`{`),
		"too big": json.RawMessage(`{"a":"` + string(make([]byte, maxConfigBytes)) + `"}`),
	} {
		if _, err := s.Connect(context.Background(), 1, "stripe", config); !errors.Is(err, domain.ErrInvalidConfig) {
			t.Errorf("%s: err = %v, want ErrInvalidConfig", name, err)
		}
	}
}

func TestRedactConfigMasksSecretsAtAnyDepth(t *testing.T) {
	raw := json.RawMessage(`{"account":"acct_1","secret_key":"sk_live_x","nested":{"webhookSecret":"whsec_y","mode":"live"},"list":[{"apiToken":"t"}]}`)
	var got map[string]any
	if err := json.Unmarshal(RedactConfig(raw), &got); err != nil {
		t.Fatal(err)
	}
	if got["account"] != "acct_1" || got["secret_key"] != "[redacted]" {
		t.Fatalf("top level = %v", got)
	}
	nested := got["nested"].(map[string]any)
	if nested["webhookSecret"] != "[redacted]" || nested["mode"] != "live" {
		t.Fatalf("nested = %v", nested)
	}
	if got["list"].([]any)[0].(map[string]any)["apiToken"] != "[redacted]" {
		t.Fatalf("list = %v", got["list"])
	}
}

func TestRedactConfigNeverReturnsUnparseableInput(t *testing.T) {
	if got := string(RedactConfig(json.RawMessage(`{sk_live_x`))); got != "{}" {
		t.Fatalf("got %q, want {}", got)
	}
}
