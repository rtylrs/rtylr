package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListPublicLocationsIsBusinessScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	otherBusinessID := int64(99201)
	testutil.GrantBusinessAccess(t, db, otherBusinessID, testutil.OwnerID)

	primary, err := repo.CreateLocation(ctx, testutil.BusinessID, domain.LocationRecord{Name: "Primary public location"})
	if err != nil {
		t.Fatalf("create primary location: %v", err)
	}
	other, err := repo.CreateLocation(ctx, otherBusinessID, domain.LocationRecord{Name: "Other business location"})
	if err != nil {
		t.Fatalf("create other location: %v", err)
	}

	locations, err := repo.ListPublicLocations(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("ListPublicLocations: %v", err)
	}
	ids := make([]int64, 0, len(locations))
	for _, location := range locations {
		ids = append(ids, location.ID)
	}
	testutil.AssertTenantIsolation(t, ids, []int64{primary.ID}, []int64{other.ID})
}

func TestCreateLocation_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	loc, err := repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationRecord{
		Name:    "Test Location Infra",
		Address: "1 Test St",
		City:    "Testville",
		Country: "TC",
	})
	if err != nil {
		t.Fatalf("CreateLocation: %v", err)
	}
	if loc.ID == 0 {
		t.Error("expected non-zero ID after create")
	}
}

func TestCreateLocation_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	loc, err := repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationRecord{
		Name:    "Name Store Location",
		Country: "TC",
	})
	if err != nil {
		t.Fatalf("CreateLocation: %v", err)
	}
	if loc.Name != "Name Store Location" {
		t.Errorf("Name = %q, want 'Name Store Location'", loc.Name)
	}
}

func TestCreateLocation_OwnerIDSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	loc, err := repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationRecord{
		Name:    "OwnerID Location",
		Country: "TC",
	})
	if err != nil {
		t.Fatalf("CreateLocation: %v", err)
	}
	if loc.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID = %d, want %d", loc.OwnerID, testutil.OwnerID)
	}
}

func TestListLocations_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.ListLocations(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListLocations before: %v", err)
	}

	_, err = repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationRecord{
		Name:    "Count Increase Location",
		Country: "TC",
	})
	if err != nil {
		t.Fatalf("CreateLocation: %v", err)
	}

	after, err := repo.ListLocations(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListLocations after: %v", err)
	}
	if len(after) <= len(before) {
		t.Errorf("expected count to increase from %d, got %d", len(before), len(after))
	}
}

func TestUpdateLocation_NameChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationRecord{
		Name:    "Before Location",
		Country: "TC",
	})
	if err != nil {
		t.Fatalf("CreateLocation: %v", err)
	}

	created.Name = "After Location"
	updated, err := repo.UpdateLocation(ctx, testutil.OwnerID, *created)
	if err != nil {
		t.Fatalf("UpdateLocation: %v", err)
	}
	if updated.Name != "After Location" {
		t.Errorf("Name = %q, want 'After Location'", updated.Name)
	}
}

func TestGetLocation_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateLocation(ctx, testutil.OwnerID, domain.LocationRecord{
		Name:    "Roundtrip Location",
		City:    "TestCity",
		Country: "TC",
	})
	if err != nil {
		t.Fatalf("CreateLocation: %v", err)
	}

	got, err := repo.GetLocation(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetLocation: %v", err)
	}
	if got.Name != "Roundtrip Location" {
		t.Errorf("Name = %q, want 'Roundtrip Location'", got.Name)
	}
	if got.City != "TestCity" {
		t.Errorf("City = %q, want 'TestCity'", got.City)
	}
}

func TestGetLocation_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetLocation(ctx, testutil.OwnerID, 999999999)
	if err == nil {
		t.Error("expected error for unknown location, got nil")
	}
}
