package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_equipment/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_equipment/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestAssignAssetToEmployee_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	asset, err := repo.CreateAsset(ctx, testutil.OwnerID, domain.CreateAssetParams{
		Name:     "Assign Test Keyboard",
		Category: "peripherals",
	})
	if err != nil {
		t.Fatalf("CreateAsset: %v", err)
	}

	if err := repo.AssignAssetToEmployee(ctx, testutil.OwnerID, asset.ID, 4101); err != nil {
		t.Errorf("AssignAssetToEmployee: %v", err)
	}
}

func TestUnassignAsset_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	asset, err := repo.CreateAsset(ctx, testutil.OwnerID, domain.CreateAssetParams{
		Name:     "Unassign Test Mouse",
		Category: "peripherals",
	})
	if err != nil {
		t.Fatalf("CreateAsset: %v", err)
	}

	if err := repo.AssignAssetToEmployee(ctx, testutil.OwnerID, asset.ID, 4101); err != nil {
		t.Fatalf("AssignAssetToEmployee: %v", err)
	}

	if err := repo.UnassignAsset(ctx, testutil.OwnerID, asset.ID); err != nil {
		t.Errorf("UnassignAsset: %v", err)
	}
}

func TestUpdateEquipmentRequestStatus_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	req, err := repo.CreateEquipmentRequest(ctx, testutil.OwnerID, domain.CreateEquipmentRequestParams{
		EmployeeID: 4101,
		ItemName:   "Status Update Headset",
		Quantity:   1,
	})
	if err != nil {
		t.Fatalf("CreateEquipmentRequest: %v", err)
	}

	if err := repo.UpdateEquipmentRequestStatus(ctx, testutil.OwnerID, req.ID, 0, domain.EquipmentStatusApproved, 1, 4101, "approved by manager"); err != nil {
		t.Errorf("UpdateEquipmentRequestStatus: %v", err)
	}
}

func TestUpdateEquipmentRequestStatus_ReflectedOnGet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	req, err := repo.CreateEquipmentRequest(ctx, testutil.OwnerID, domain.CreateEquipmentRequestParams{
		EmployeeID: 4101,
		ItemName:   "Status Reflect Webcam",
		Quantity:   1,
	})
	if err != nil {
		t.Fatalf("CreateEquipmentRequest: %v", err)
	}

	if err := repo.UpdateEquipmentRequestStatus(ctx, testutil.OwnerID, req.ID, 0, domain.EquipmentStatusRejected, 1, 4101, "budget exceeded"); err != nil {
		t.Fatalf("UpdateEquipmentRequestStatus: %v", err)
	}

	got, err := repo.GetEquipmentRequest(ctx, testutil.OwnerID, req.ID)
	if err != nil {
		t.Fatalf("GetEquipmentRequest: %v", err)
	}
	if got.Status != domain.EquipmentStatusRejected {
		t.Errorf("expected status %q, got %q", domain.EquipmentStatusRejected, got.Status)
	}
}
