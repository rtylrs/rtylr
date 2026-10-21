package infrastructure_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/integrations/internal/domain"
	"github.com/voxire/onion/internal/modules/integrations/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListIntegrations_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	integrations, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List integrations: %v", err)
	}
	_ = integrations
}

func TestListIntegrations_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	integrations, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List integrations: %v", err)
	}
	for _, i := range integrations {
		if i.OwnerID != testutil.OwnerID {
			t.Errorf("integration %d has owner_id %d, want %d", i.ID, i.OwnerID, testutil.OwnerID)
		}
	}
}

func TestConnectIntegration_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg := json.RawMessage(`{"api_key":"test_key"}`)
	id, err := repo.Connect(ctx, testutil.OwnerID, "test_provider_connect", cfg)
	if err != nil {
		t.Fatalf("Connect integration: %v", err)
	}
	if id == 0 {
		t.Fatal("Connect returned id=0")
	}
}

func TestConnectIntegration_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	provider := "test_provider_list_check"
	cfg := json.RawMessage(`{}`)
	_, err := repo.Connect(ctx, testutil.OwnerID, provider, cfg)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	integrations, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after connect: %v", err)
	}
	found := false
	for _, i := range integrations {
		if i.Provider == provider {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("connected provider %q not found in list", provider)
	}
}

func TestDisconnectIntegration_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg := json.RawMessage(`{}`)
	id, err := repo.Connect(ctx, testutil.OwnerID, "test_provider_disconnect", cfg)
	if err != nil {
		t.Fatalf("Connect for disconnect test: %v", err)
	}

	if err := repo.Disconnect(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
}

func TestListIntegrations_UnknownOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	integrations, err := repo.List(ctx, 9999999)
	if err != nil {
		t.Fatalf("List integrations unknown owner: %v", err)
	}
	if len(integrations) != 0 {
		t.Errorf("expected 0 integrations for unknown owner, got %d", len(integrations))
	}
}

func TestConnectIntegration_EmptyConfig_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Connect(ctx, testutil.OwnerID, "test_provider_empty_cfg", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Connect with empty config: %v", err)
	}
	if id == 0 {
		t.Fatal("Connect returned id=0")
	}
}

func TestConnectIntegration_ProviderFieldPreserved(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	provider := "test_provider_field_check"
	cfg := json.RawMessage(`{"key":"value"}`)
	_, err := repo.Connect(ctx, testutil.OwnerID, provider, cfg)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	integrations, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, i := range integrations {
		if i.Provider == provider {
			if i.Provider == "" {
				t.Error("provider field is empty after connect")
			}
			return
		}
	}
	t.Errorf("provider %q not found after connect", provider)
}

func TestDisconnect_NonExistentID_IsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.Disconnect(ctx, testutil.OwnerID, 9999999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Disconnect non-existent = %v, want ErrNotFound", err)
	}
}

func TestConnectIntegration_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Connect(ctx, testutil.OwnerID, "test_positive_id", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestConnectIntegration_IDsAreUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Connect(ctx, testutil.OwnerID, "test_unique_a", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Connect A: %v", err)
	}
	id2, err := repo.Connect(ctx, testutil.OwnerID, "test_unique_b", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Connect B: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %d", id1)
	}
}

func TestListIntegrations_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Connect(ctx, testutil.OwnerID, "test_id_check", json.RawMessage(`{}`))

	integrations, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, intg := range integrations {
		if intg.ID == 0 {
			t.Errorf("integrations[%d].ID is zero", i)
		}
	}
}

func TestDisconnect_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Connect(ctx, testutil.OwnerID, "test_disconnect_list", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	if err := repo.Disconnect(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}

	integrations, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after disconnect: %v", err)
	}
	for _, intg := range integrations {
		if intg.ID == id {
			t.Errorf("disconnected integration id=%d still in list", id)
		}
	}
}

func TestConnectIntegration_WithComplexConfig(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg := json.RawMessage(`{"api_key":"sk-abc","base_url":"https://api.example.com","timeout":30}`)
	id, err := repo.Connect(ctx, testutil.OwnerID, "test_complex_cfg", cfg)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestListIntegrations_AfterConnect_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	provider := fmt.Sprintf("test_count_%d", time.Now().UnixNano())
	id, err := repo.Connect(ctx, testutil.OwnerID, provider, json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = repo.Disconnect(ctx, testutil.OwnerID, id) })

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count +1: before=%d after=%d", len(before), len(after))
	}
}

func TestConnect_SameProviderReusesTheRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	provider := fmt.Sprintf("test_reuse_%d", time.Now().UnixNano())
	first, err := repo.Connect(ctx, testutil.OwnerID, provider, json.RawMessage(`{"a":1}`))
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = repo.Disconnect(ctx, testutil.OwnerID, first) })
	second, err := repo.Connect(ctx, testutil.OwnerID, provider, json.RawMessage(`{"a":2}`))
	if err != nil {
		t.Fatalf("Connect again: %v", err)
	}
	if second != first {
		t.Fatalf("connecting %q twice created row %d then %d", provider, first, second)
	}
}

func TestListIntegrations_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	integrations, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if integrations == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestConnectIntegration_ZeroOwner_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Connect(ctx, 0, "test_zero_owner", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Connect with owner=0: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestListIntegrations_EachHasNonEmptyProvider(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Connect(ctx, testutil.OwnerID, "test_provider_nonempty", json.RawMessage(`{}`))

	integrations, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, intg := range integrations {
		if intg.Provider == "" {
			t.Errorf("integrations[%d].Provider is empty", i)
		}
	}
}

func TestDisconnect_WrongOwner_DoesNotRemove(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Connect(ctx, testutil.OwnerID, "test_wrong_owner_disc", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}

	_ = repo.Disconnect(ctx, 999999, id)

	integrations, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after wrong-owner disconnect: %v", err)
	}
	found := false
	for _, intg := range integrations {
		if intg.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("integration id=%d should still exist after wrong-owner disconnect", id)
	}
}
