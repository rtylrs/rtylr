package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/application"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
	hrequipmentv1 "github.com/voxire/voxie/gen/go/hr/equipment/v1"
)

// TestEquipmentAPIListAssetsAppliesEmployeeFilterAndReportsTotal covers the
// GET /hr/equipment/v1/assets handler: employee_id used to be dropped (every
// employee's assets were returned) and total/limit were fabricated from the
// returned page length instead of the full matching set.
func TestEquipmentAPIListAssetsAppliesEmployeeFilterAndReportsTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)
	service := application.NewService(ctx, db, nil)

	employeeID := int64(4112)
	asset, err := repo.CreateAsset(ctx, testutil.OwnerID, domain.CreateAssetParams{
		Name:     "Employee Filtered Monitor",
		Category: "electronics",
	})
	if err != nil {
		t.Fatalf("CreateAsset: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), "DELETE FROM hr_assets WHERE id = ?", asset.ID); err != nil {
			t.Errorf("cleanup asset: %v", err)
		}
	})
	if err := repo.AssignAssetToEmployee(ctx, testutil.OwnerID, asset.ID, employeeID); err != nil {
		t.Fatalf("AssignAssetToEmployee: %v", err)
	}

	apiCtx := authPkg.WithAuthUser(ctx, &authPkg.AuthUser{UserID: testutil.OwnerID, BusinessID: testutil.BusinessID, Role: "owner"})
	resp, err := NewEquipmentAPI(service).ListAssets(apiCtx, &hrequipmentv1.ListAssetsRequest{EmployeeId: &employeeID})
	if err != nil {
		t.Fatalf("ListAssets: %v", err)
	}

	found := false
	for _, payload := range resp.GetAssets() {
		if payload.GetEmployeeId() != employeeID {
			t.Errorf("asset %d assigned to %d leaked into the employee %d filter", payload.GetId(), payload.GetEmployeeId(), employeeID)
		}
		if payload.GetId() == asset.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("asset %d assigned to employee %d is missing from the response", asset.ID, employeeID)
	}
	if resp.GetTotal() < 1 {
		t.Errorf("total = %d, want at least the seeded asset", resp.GetTotal())
	}
	if resp.GetLimit() != 20 {
		t.Errorf("limit = %d, want the effective page size 20", resp.GetLimit())
	}
}
