package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/apps/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetEnabledAddons_ReturnsNonNilMap(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	addons, err := repo.GetEnabledAddons(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("GetEnabledAddons: %v", err)
	}
	if addons == nil {
		t.Fatal("expected a non-nil addon map")
	}
}

func TestGetEnabledAddons_UnknownBusinessReturnsEmptyMap(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	addons, err := repo.GetEnabledAddons(context.Background(), 999999)
	if err != nil {
		t.Fatalf("GetEnabledAddons: %v", err)
	}
	if len(addons) != 0 {
		t.Fatalf("unknown business has %d enabled addons, want none", len(addons))
	}
}
