package infrastructure_test

import (
	"context"
	"encoding/json"
	"testing"

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

func TestDisconnect_NonExistentID_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Disconnecting a non-existent ID should not error (0 rows affected is fine).
	err := repo.Disconnect(ctx, testutil.OwnerID, 9999999)
	if err != nil {
		t.Fatalf("Disconnect non-existent: %v", err)
	}
}
