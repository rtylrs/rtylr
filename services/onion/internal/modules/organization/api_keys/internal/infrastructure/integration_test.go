package infrastructure_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/api_keys/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/api_keys/internal/infrastructure"
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

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "test-key", "hash-abc123", "sk_test", []string{"read"})
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
	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, name, "hash-xyz789", "sk_int", []string{"read", "write"})
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
	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "scope-test-key", "hash-scope123", "sk_sc", scopes)
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

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "revoke-test-key", "hash-rev123", "sk_rv", []string{})
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

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "cross-owner-key", "hash-cross123", "sk_cr", []string{})
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

	_, _ = repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "named-key", "hash-name1", "sk_nm", []string{})
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

	_, _ = repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "prefix-test-key", "hash-pfx1", "sk_px", []string{})
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

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "empty-scope-key", "hash-empty1", "sk_es", []string{})
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

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "delete-then-check", "hash-del1", "sk_dl", []string{})
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

func TestCreate_IDsAreUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "unique-key-a", "hash-ua1", "sk_ua", []string{})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}
	id2, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "unique-key-b", "hash-ub1", "sk_ub", []string{})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %d", id1)
	}
}

func TestCreate_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	_, err = repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "count-inc-key", "hash-ci1", "sk_ci", []string{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count +1: before=%d after=%d", len(before), len(after))
	}
}

func TestCreate_MultipleScopesStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	scopes := []string{"orders:read", "products:read", "customers:write"}
	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "multi-scope-key", "hash-ms1", "sk_ms", scopes)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, k := range keys {
		if k.ID == id {
			if len(k.Scopes) != len(scopes) {
				t.Errorf("Scopes len = %d, want %d", len(k.Scopes), len(scopes))
			}
			return
		}
	}
	t.Errorf("key id=%d not found", id)
}

func TestList_KeysHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "id-check-key", "hash-ic1", "sk_ic", []string{})

	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, k := range keys {
		if k.ID == 0 {
			t.Errorf("keys[%d].ID is zero", i)
		}
	}
}

func TestRevoke_NonExistent_IsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.Revoke(ctx, testutil.OwnerID, 9999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Revoke non-existent = %v, want ErrNotFound", err)
	}
}

func TestCreate_SingleScopeStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	scopes := []string{"reports:read"}
	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "single-scope-key", "hash-ss1", "sk_ss", scopes)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, k := range keys {
		if k.ID == id {
			if len(k.Scopes) != 1 {
				t.Errorf("expected 1 scope, got %d", len(k.Scopes))
			}
			return
		}
	}
	t.Errorf("key id=%d not found", id)
}

func TestCreate_NameStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const want = "my-production-key"
	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, want, "hash-name-stored", "sk_ns", []string{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, k := range keys {
		if k.ID == id {
			if k.Name != want {
				t.Errorf("Name = %q, want %q", k.Name, want)
			}
			return
		}
	}
	t.Errorf("key id=%d not found", id)
}

func TestCreate_PrefixStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const want = "sk_prod"
	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "prefix-stored-key", "hash-pfx-stored", want, []string{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, k := range keys {
		if k.ID == id {
			if k.KeyPrefix != want {
				t.Errorf("KeyPrefix = %q, want %q", k.KeyPrefix, want)
			}
			return
		}
	}
	t.Errorf("key id=%d not found", id)
}

func TestList_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	keys, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected 0 keys for owner_id=0, got %d", len(keys))
	}
}

func TestRevoke_CreatedKey_NoLongerInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "revoke-list-check", "hash-rlc1", "sk_rlc", []string{"read"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Revoke(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Revoke: %v", err)
	}

	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, k := range keys {
		if k.ID == id {
			t.Errorf("revoked key id=%d still in list", id)
		}
	}
}

func TestCreate_TwoRevoke_One_OtherStillPresent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "keep-key", "hash-kk1", "sk_kk", []string{})
	if err != nil {
		t.Fatalf("Create keep: %v", err)
	}
	id2, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, "revoke-key", "hash-rk1", "sk_rk", []string{})
	if err != nil {
		t.Fatalf("Create revoke: %v", err)
	}

	if err := repo.Revoke(ctx, testutil.OwnerID, id2); err != nil {
		t.Fatalf("Revoke: %v", err)
	}

	keys, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	foundKeep := false
	for _, k := range keys {
		if k.ID == id1 {
			foundKeep = true
		}
		if k.ID == id2 {
			t.Errorf("revoked key id=%d still present", id2)
		}
	}
	if !foundKeep {
		t.Errorf("kept key id=%d not found after revoking sibling", id1)
	}
}
