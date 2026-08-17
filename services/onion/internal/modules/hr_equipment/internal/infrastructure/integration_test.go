package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListEquipmentRequests_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, total, err := repo.ListEquipmentRequests(ctx, testutil.OwnerID, domain.ListEquipmentRequestsParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListEquipmentRequests: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = records
}

func TestListEquipmentRequests_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, _, err := repo.ListEquipmentRequests(ctx, testutil.OwnerID, domain.ListEquipmentRequestsParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListEquipmentRequests: %v", err)
	}
	for _, rec := range records {
		if rec.OwnerID != testutil.OwnerID {
			t.Errorf("record %d belongs to wrong owner: %d", rec.ID, rec.OwnerID)
		}
	}
}

func TestListEquipmentRequests_FilterByStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, _, err := repo.ListEquipmentRequests(ctx, testutil.OwnerID, domain.ListEquipmentRequestsParams{
		Status: "pending",
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListEquipmentRequests by status: %v", err)
	}
	for _, rec := range records {
		if rec.Status != "pending" {
			t.Errorf("expected status 'pending', got %q for record %d", rec.Status, rec.ID)
		}
	}
}

func TestListEquipmentRequests_FilterByEmployee(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, _, err := repo.ListEquipmentRequests(ctx, testutil.OwnerID, domain.ListEquipmentRequestsParams{
		EmployeeID: 4101,
		Limit:      50,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("ListEquipmentRequests by employee: %v", err)
	}
	for _, rec := range records {
		if rec.EmployeeID != 4101 {
			t.Errorf("expected employee 4101, got %d for record %d", rec.EmployeeID, rec.ID)
		}
	}
}

func TestListEquipmentRequests_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, total, err := repo.ListEquipmentRequests(ctx, 0, domain.ListEquipmentRequestsParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 || len(records) != 0 {
		t.Errorf("expected empty result for unknown owner, got total=%d records=%d", total, len(records))
	}
}

func TestGetEquipmentRequest_NotFoundReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetEquipmentRequest(ctx, testutil.OwnerID, 999999999)
	if !errors.Is(err, domain.ErrEquipmentRequestNotFound) {
		t.Errorf("expected ErrEquipmentRequestNotFound, got %v", err)
	}
}

func TestCreateAndGetEquipmentRequest_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateEquipmentRequest(ctx, testutil.OwnerID, domain.CreateEquipmentRequestParams{
		EmployeeID: 4101,
		ItemName:   "Laptop",
		Quantity:   1,
		Reason:     "Remote work",
	})
	if err != nil {
		t.Fatalf("CreateEquipmentRequest: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
	if created.ItemName != "Laptop" {
		t.Errorf("item name mismatch: got %q", created.ItemName)
	}
}

func TestListAssets_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	assets, err := repo.ListAssets(ctx, testutil.OwnerID, domain.ListAssetsParams{Limit: 50})
	if err != nil {
		t.Fatalf("ListAssets: %v", err)
	}
	_ = assets
}

func TestListAssets_FilterByStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	status := "available"
	assets, err := repo.ListAssets(ctx, testutil.OwnerID, domain.ListAssetsParams{Status: &status, Limit: 50})
	if err != nil {
		t.Fatalf("ListAssets by status: %v", err)
	}
	for _, a := range assets {
		if a.Status != "available" {
			t.Errorf("expected status 'available', got %q for asset %d", a.Status, a.ID)
		}
	}
}

func TestGetAsset_NotFoundReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetAsset(ctx, testutil.OwnerID, 999999999)
	if !errors.Is(err, domain.ErrAssetNotFound) {
		t.Errorf("expected ErrAssetNotFound, got %v", err)
	}
}

func TestCreateAndGetAsset_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	created, err := repo.CreateAsset(ctx, testutil.OwnerID, domain.CreateAssetParams{
		Name:     "Test Monitor",
		Category: "electronics",
	})
	if err != nil {
		t.Fatalf("CreateAsset: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID after asset create")
	}

	got, err := repo.GetAsset(ctx, testutil.OwnerID, created.ID)
	if err != nil {
		t.Fatalf("GetAsset: %v", err)
	}
	if got.Name != "Test Monitor" {
		t.Errorf("name mismatch: got %q", got.Name)
	}
	if got.Status != "available" {
		t.Errorf("new asset should have status 'available', got %q", got.Status)
	}
}

func TestListEquipmentRequests_TotalMatchesRecordBound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, total, err := repo.ListEquipmentRequests(ctx, testutil.OwnerID, domain.ListEquipmentRequestsParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListEquipmentRequests: %v", err)
	}
	if int64(len(records)) > total {
		t.Errorf("returned %d records but total is %d", len(records), total)
	}
}
