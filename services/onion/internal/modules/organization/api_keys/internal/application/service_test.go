package application

import (
	"context"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/api_keys/internal/domain"
)

func TestGenerateAPIKey_ProducesUniquePrefixedKeys(t *testing.T) {
	a, err := generateAPIKey()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	b, err := generateAPIKey()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if a == b {
		t.Fatal("expected two generated keys to differ")
	}
	if !strings.HasPrefix(a, apiKeyPrefix) {
		t.Errorf("expected API key prefix %q, got %q", apiKeyPrefix, a)
	}
	for _, c := range strings.TrimPrefix(a, apiKeyPrefix) {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("non-hex character %q in key %q", c, a)
			break
		}
	}
}

func TestHashKey_DeterministicAndDifferentForDifferentInput(t *testing.T) {
	first := hashKey("abc")
	second := hashKey("abc")
	if first != second {
		t.Fatal("hashKey should be deterministic for the same input")
	}
	if hashKey("abc") == hashKey("abd") {
		t.Fatal("hashKey should differ for different input")
	}
	// SHA-256 hex is exactly 64 chars.
	if len(hashKey("abc")) != 64 {
		t.Errorf("expected 64-char hex hash, got %d", len(hashKey("abc")))
	}
}

func TestCreateAPIKeyInputFromRequest(t *testing.T) {
	input := CreateAPIKeyInputFromRequest("my key", []string{"read", "write"})
	if input.Name != "my key" || strings.Join(input.Scopes, ",") != "read,write" {
		t.Errorf("unexpected input: %+v", input)
	}
}

type mockRepo struct {
	listResult []domain.APIKey
	listErr    error

	createID  int64
	createErr error
	createGot struct {
		ownerID    int64
		businessID int64
		name       string
		hash       string
		prefix     string
		scopes     []string
	}

	revokeErr error
	revokeGot struct {
		businessID int64
		keyID      int64
	}
}

func (m *mockRepo) List(_ context.Context, businessID int64) ([]domain.APIKey, error) {
	_ = businessID
	return m.listResult, m.listErr
}

func (m *mockRepo) Create(_ context.Context, ownerID, businessID int64, name, hash, prefix string, scopes []string) (int64, error) {
	m.createGot.ownerID = ownerID
	m.createGot.businessID = businessID
	m.createGot.name = name
	m.createGot.hash = hash
	m.createGot.prefix = prefix
	m.createGot.scopes = scopes
	return m.createID, m.createErr
}

func (m *mockRepo) Revoke(_ context.Context, businessID, keyID int64) error {
	m.revokeGot.businessID = businessID
	m.revokeGot.keyID = keyID
	return m.revokeErr
}

func TestCreate_RequiresName(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, _, err := s.Create(context.Background(), 1, 1, domain.CreateAPIKeyInput{}); err == nil {
		t.Fatalf("expected error for empty name")
	}
}

func TestRevoke_RequiresKeyID(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Revoke(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected error for missing key id")
	}
}

func TestList_DelegatesToRepo(t *testing.T) {
	want := []domain.APIKey{{ID: 1, OwnerID: 7, Name: "ci"}}
	m := &mockRepo{listResult: want}
	s := &Service{repo: m}
	got, err := s.List(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 1 || got[0].ID != 1 {
		t.Errorf("expected one key, got %+v", got)
	}
}

func TestCreate_HappyPathHashesAndStores(t *testing.T) {
	m := &mockRepo{createID: 99}
	s := &Service{repo: m}

	id, raw, err := s.Create(context.Background(), 7, 7, domain.CreateAPIKeyInput{
		Name:   "test key",
		Scopes: []string{"read"},
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if id != 99 {
		t.Errorf("expected id 99, got %d", id)
	}
	if !strings.HasPrefix(raw, apiKeyPrefix) {
		t.Errorf("raw key should include API prefix %q, got %q", apiKeyPrefix, raw)
	}
	if m.createGot.ownerID != 7 || m.createGot.name != "test key" {
		t.Errorf("repo args unexpected: %+v", m.createGot)
	}
	if m.createGot.prefix != raw[:16] {
		t.Errorf("prefix should be first 16 of raw key")
	}
	if m.createGot.hash == raw {
		t.Error("hash should not equal raw key (was supposed to be sha256-hex)")
	}
	if len(m.createGot.hash) != 64 {
		t.Errorf("hash should be 64-char hex, got %d", len(m.createGot.hash))
	}
}

func TestCreate_NilScopesNormalizedToEmpty(t *testing.T) {
	m := &mockRepo{createID: 1}
	s := &Service{repo: m}
	if _, _, err := s.Create(context.Background(), 1, 1, domain.CreateAPIKeyInput{Name: "k"}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if m.createGot.scopes == nil || len(m.createGot.scopes) != 0 {
		t.Errorf("expected empty (non-nil) scope slice, got %v", m.createGot.scopes)
	}
}

func TestCreate_PropagatesRepoError(t *testing.T) {
	m := &mockRepo{createErr: errSentinel}
	s := &Service{repo: m}
	if _, _, err := s.Create(context.Background(), 1, 1, domain.CreateAPIKeyInput{Name: "k"}); err == nil {
		t.Fatal("expected error to bubble up")
	}
}

func TestRevoke_DelegatesWithIDs(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	if err := s.Revoke(context.Background(), 7, 42); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if m.revokeGot.businessID != 7 || m.revokeGot.keyID != 42 {
		t.Errorf("revoke args mismatch: %+v", m.revokeGot)
	}
}

var errSentinel = sentinelErr("repo failed")

type sentinelErr string

func (s sentinelErr) Error() string { return string(s) }
