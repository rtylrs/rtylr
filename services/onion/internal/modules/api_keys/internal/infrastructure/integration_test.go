package infrastructure_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/api_keys/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestList_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if keys == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestList_WrongOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	keys, err := repo.List(ctx, 999999)
	if err != nil {
		t.Fatalf("List wrong owner: %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("expected 0 keys for unknown owner, got %d", len(keys))
	}
}

func TestCreate_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "test-key", "hash-abc123", "sk_test", []string{"read"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreate_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	name := fmt.Sprintf("integration-test-key-%d", testutil.OwnerID)
	id, err := repo.Create(ctx, testutil.OwnerID, name, "hash-xyz789", "sk_int", []string{"read", "write"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after create: %v", err)
	}

	found := false
	for _, k := range keys {
		if k.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created key id=%d not found in list", id)
	}
}

func TestCreate_ScopesRoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	scopes := []string{"orders:read", "products:write"}
	id, err := repo.Create(ctx, testutil.OwnerID, "scope-test-key", "hash-scope123", "sk_sc", scopes)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	for _, k := range keys {
		if k.ID != id {
			continue
		}
		if len(k.Scopes) != len(scopes) {
			t.Errorf("scopes length: got %d, want %d", len(k.Scopes), len(scopes))
		}
		break
	}
}

func TestRevoke_CreatedKey_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "revoke-test-key", "hash-rev123", "sk_rv", []string{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Revoke(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
}

func TestRevoke_WrongOwner_NotRevoked(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "cross-owner-key", "hash-cross123", "sk_cr", []string{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Revoking with wrong owner should not error (0 rows affected silently ignored by repo)
	_ = repo.Revoke(ctx, 999999, id)

	// Key should still appear in list for real owner
	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, k := range keys {
		if k.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("key should still exist after wrong-owner revoke attempt")
	}
}

func TestList_KeysHaveOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, k := range keys {
		if k.OwnerID != testutil.OwnerID {
			t.Errorf("keys[%d].OwnerID = %d, want %d", i, k.OwnerID, testutil.OwnerID)
		}
	}
}

func TestList_KeysHaveNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, "named-key", "hash-name1", "sk_nm", []string{})
	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, k := range keys {
		if k.Name == "" {
			t.Errorf("keys[%d].Name is empty", i)
		}
	}
}

func TestList_KeysHavePrefix(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, "prefix-test-key", "hash-pfx1", "sk_px", []string{})
	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, k := range keys {
		if k.KeyPrefix == "" {
			t.Errorf("keys[%d].KeyPrefix is empty", i)
		}
	}
}

func TestCreate_EmptyScopes_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "empty-scope-key", "hash-empty1", "sk_es", []string{})
	if err != nil {
		t.Fatalf("Create with empty scopes: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestRevoke_ThenNotInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, "delete-then-check", "hash-del1", "sk_dl", []string{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Revoke(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Revoke: %v", err)
	}

	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after revoke: %v", err)
	}
	for _, k := range keys {
		if k.ID == id {
			t.Errorf("revoked key id=%d still appears in list", id)
		}
	}
}
