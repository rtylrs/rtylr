package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
	"github.com/voxire/onion/internal/modules/suppliers/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/audit"
	"github.com/voxire/onion/internal/testutil"
)

func TestListSuppliers_ReturnsFixtureRows(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	suppliers, _, err := repo.ListSuppliers(context.Background(), testutil.OwnerID, 50, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}
	if len(suppliers) == 0 {
		t.Fatal("expected at least one supplier, got none")
	}

	found := false
	for _, s := range suppliers {
		if s.Name == "Test Supplier" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find 'Test Supplier' in results, got: %v", suppliers)
	}
}

func TestGetSupplier_ReturnsCorrectNameAndActiveStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	// List all and find supplier 5101
	suppliers, _, err := repo.ListSuppliers(context.Background(), testutil.OwnerID, 100, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}

	var found bool
	for _, s := range suppliers {
		if s.ID == 5101 {
			found = true
			if s.Name != "Test Supplier" {
				t.Errorf("expected name='Test Supplier', got %q", s.Name)
			}
			if !s.IsActive {
				t.Errorf("expected is_active=true for supplier 5101, got false")
			}
			break
		}
	}
	if !found {
		t.Fatal("supplier id=5101 not found in listing for owner 1001")
	}
}

func TestListSuppliers_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	suppliers, _, err := repo.ListSuppliers(ctx, 0, 50, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}
	if len(suppliers) != 0 {
		t.Errorf("expected empty slice for owner_id=0, got %d rows", len(suppliers))
	}
}

func TestListSuppliers_ContainsTestSupplier(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	suppliers, _, err := repo.ListSuppliers(ctx, testutil.OwnerID, 50, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}

	var found bool
	for _, s := range suppliers {
		if s.ID == 5101 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected supplier id=5101 in results")
	}
}

func TestListSuppliers_NameIsTestSupplier(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	suppliers, _, err := repo.ListSuppliers(ctx, testutil.OwnerID, 50, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}

	for _, s := range suppliers {
		if s.ID == 5101 {
			if s.Name != "Test Supplier" {
				t.Errorf("expected name='Test Supplier', got %q", s.Name)
			}
			return
		}
	}
	t.Fatal("supplier id=5101 not found")
}

func TestListSuppliers_EmailIsSupplierTestLocal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	suppliers, _, err := repo.ListSuppliers(ctx, testutil.OwnerID, 50, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}

	for _, s := range suppliers {
		if s.ID == 5101 {
			if s.Email != "supplier@test.local" {
				t.Errorf("expected email='supplier@test.local', got %q", s.Email)
			}
			return
		}
	}
	t.Fatal("supplier id=5101 not found")
}

func TestListSuppliers_PaginationFirstPage(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	suppliers, _, err := repo.ListSuppliers(ctx, testutil.OwnerID, 1, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}
	if len(suppliers) > 1 {
		t.Errorf("expected at most 1 row with limit=1, got %d", len(suppliers))
	}
}

func TestListSuppliers_HasMoreFalseWhenFewRows(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, hasMore, err := repo.ListSuppliers(ctx, testutil.OwnerID, 1000, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}
	if hasMore {
		t.Error("expected hasMore=false with large limit, got true")
	}
}

func TestCreateSupplier_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SupplierInput{
		Name:     "New Test Supplier",
		Email:    "new@supplier.test",
		IsActive: true,
	}

	id, err := repo.CreateSupplier(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreateSupplier error: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCreateSupplier_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SupplierInput{
		Name:     "StoredNameSupplier",
		IsActive: true,
	}

	id, err := repo.CreateSupplier(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreateSupplier error: %v", err)
	}

	suppliers, _, err := repo.ListSuppliers(ctx, testutil.OwnerID, 1000, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}

	for _, s := range suppliers {
		if s.ID == id {
			if s.Name != "StoredNameSupplier" {
				t.Errorf("expected name='StoredNameSupplier', got %q", s.Name)
			}
			return
		}
	}
	t.Fatalf("created supplier id=%d not found in list", id)
}

func TestCreateSupplier_EmailIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SupplierInput{
		Name:     "EmailSupplier",
		Email:    "stored@email.test",
		IsActive: true,
	}

	id, err := repo.CreateSupplier(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreateSupplier error: %v", err)
	}

	suppliers, _, err := repo.ListSuppliers(ctx, testutil.OwnerID, 1000, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}

	for _, s := range suppliers {
		if s.ID == id {
			if s.Email != "stored@email.test" {
				t.Errorf("expected email='stored@email.test', got %q", s.Email)
			}
			return
		}
	}
	t.Fatalf("created supplier id=%d not found in list", id)
}

func TestCreateSupplier_IsActiveTrue(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.SupplierInput{
		Name:     "ActiveSupplier",
		IsActive: true,
	}

	id, err := repo.CreateSupplier(ctx, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("CreateSupplier error: %v", err)
	}

	suppliers, _, err := repo.ListSuppliers(ctx, testutil.OwnerID, 1000, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}

	for _, s := range suppliers {
		if s.ID == id {
			if !s.IsActive {
				t.Errorf("expected is_active=true, got false")
			}
			return
		}
	}
	t.Fatalf("created supplier id=%d not found in list", id)
}

func TestUpdateSupplier_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	createInput := domain.SupplierInput{
		Name:     "BeforeUpdate",
		IsActive: true,
	}
	id, err := repo.CreateSupplier(ctx, testutil.OwnerID, createInput)
	if err != nil {
		t.Fatalf("CreateSupplier error: %v", err)
	}

	updateInput := domain.SupplierInput{
		Name:     "AfterUpdate",
		IsActive: true,
	}
	if err := repo.UpdateSupplier(ctx, testutil.OwnerID, id, updateInput); err != nil {
		t.Fatalf("UpdateSupplier error: %v", err)
	}

	suppliers, _, err := repo.ListSuppliers(ctx, testutil.OwnerID, 1000, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}

	for _, s := range suppliers {
		if s.ID == id {
			if s.Name != "AfterUpdate" {
				t.Errorf("expected name='AfterUpdate', got %q", s.Name)
			}
			return
		}
	}
	t.Fatalf("updated supplier id=%d not found in list", id)
}

func TestDeleteSupplier_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateSupplier(ctx, testutil.OwnerID, domain.SupplierInput{Name: "ToDelete", IsActive: true})
	if err != nil {
		t.Fatalf("CreateSupplier error: %v", err)
	}

	if err := repo.DeleteSupplier(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteSupplier error: %v", err)
	}

	suppliers, _, err := repo.ListSuppliers(ctx, testutil.OwnerID, 1000, 0)
	if err != nil {
		t.Fatalf("ListSuppliers error: %v", err)
	}

	for _, s := range suppliers {
		if s.ID == id {
			t.Errorf("expected supplier id=%d to be deleted, but still found in list", id)
		}
	}
}

func TestDeleteSupplier_NonExistentIsNoop(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.DeleteSupplier(ctx, testutil.OwnerID, 999999999); err != nil {
		t.Errorf("expected no error deleting non-existent supplier, got: %v", err)
	}
}

func TestCountSuppliers_OwnerHasSuppliers(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	count, err := repo.CountSuppliers(ctx, testutil.OwnerID, false)
	if err != nil {
		t.Fatalf("CountSuppliers: %v", err)
	}
	if count <= 0 {
		t.Errorf("expected at least 1 supplier, got %d", count)
	}
}

func TestCountSuppliers_ZeroOwner_ReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	count, err := repo.CountSuppliers(ctx, 0, false)
	if err != nil {
		t.Fatalf("CountSuppliers(ownerID=0): %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for ownerID=0, got %d", count)
	}
}

func TestCountSuppliers_AfterCreate_Increases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.CountSuppliers(ctx, testutil.OwnerID, false)
	if err != nil {
		t.Fatalf("CountSuppliers before: %v", err)
	}

	id, err := repo.CreateSupplier(ctx, testutil.OwnerID, domain.SupplierInput{Name: "Count Test Supplier", IsActive: true})
	if err != nil {
		t.Fatalf("CreateSupplier: %v", err)
	}
	_ = id

	after, err := repo.CountSuppliers(ctx, testutil.OwnerID, false)
	if err != nil {
		t.Fatalf("CountSuppliers after: %v", err)
	}
	if after != before+1 {
		t.Errorf("expected count to increase by 1, before=%d after=%d", before, after)
	}
}

func TestCountSuppliers_ActiveOnly(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	count, err := repo.CountSuppliers(ctx, testutil.OwnerID, true)
	if err != nil {
		t.Fatalf("CountSuppliers(activeOnly): %v", err)
	}
	if count < 0 {
		t.Errorf("expected non-negative count, got %d", count)
	}
}

func TestGetSupplierByID_FixtureSupplier(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	s, err := repo.GetSupplierByID(ctx, testutil.OwnerID, 5101)
	if err != nil {
		t.Fatalf("GetSupplierByID(5101): %v", err)
	}
	if s.ID != 5101 {
		t.Errorf("expected ID=5101, got %d", s.ID)
	}
	if s.Name != "Test Supplier" {
		t.Errorf("expected Name='Test Supplier', got %q", s.Name)
	}
}

func TestGetSupplierByID_NotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetSupplierByID(ctx, testutil.OwnerID, 9999999)
	if !errors.Is(err, domain.ErrSupplierNotFound) {
		t.Fatalf("expected ErrSupplierNotFound, got %v", err)
	}
}

func TestGetSupplierByID_WrongOwner_NotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Supplier 5101 belongs to owner 1001; a different owner should not see it.
	_, err := repo.GetSupplierByID(ctx, 9999999, 5101)
	if !errors.Is(err, domain.ErrSupplierNotFound) {
		t.Fatalf("expected ErrSupplierNotFound for wrong owner, got %v", err)
	}
}

func TestCreateSupplier_AuditFailureRollsBackSupplier(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const name = "Audit rollback supplier"

	id, err := repo.CreateSupplier(ctx, testutil.OwnerID, domain.SupplierInput{Name: name, IsActive: true}, audit.Event{
		BusinessID: testutil.OwnerID,
		EntityType: "supplier",
		Action:     "created",
		NewValues:  make(chan struct{}),
	})
	if err == nil {
		t.Fatalf("CreateSupplier succeeded with id=%d despite audit failure", id)
	}

	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM erp_suppliers
		WHERE business_id = ? AND name = ?`, testutil.OwnerID, name).Scan(&count); err != nil {
		t.Fatalf("count rolled-back supplier: %v", err)
	}
	if count != 0 {
		t.Fatalf("supplier count after audit rollback = %d, want 0", count)
	}
}

func TestUpdateSupplier_AuditFailureRollsBackUpdate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	id, err := repo.CreateSupplier(ctx, testutil.OwnerID, domain.SupplierInput{Name: "Before audit", IsActive: true})
	if err != nil {
		t.Fatalf("CreateSupplier: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteSupplier(context.Background(), testutil.OwnerID, id) })

	err = repo.UpdateSupplier(ctx, testutil.OwnerID, id, domain.SupplierInput{Name: "After audit", IsActive: true}, audit.Event{
		BusinessID: testutil.OwnerID, EntityType: "supplier", EntityID: id,
		Action: "updated", NewValues: make(chan struct{}),
	})
	if err == nil {
		t.Fatal("UpdateSupplier succeeded despite audit failure")
	}
	stored, err := repo.GetSupplierByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetSupplierByID: %v", err)
	}
	if stored.Name != "Before audit" {
		t.Fatalf("name after audit rollback = %q, want Before audit", stored.Name)
	}
}

func TestGetSupplierByID_CreatedThenFetched(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateSupplier(ctx, testutil.OwnerID, domain.SupplierInput{
		Name:     "FetchByID Supplier",
		Email:    "fetchbyid@test.local",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateSupplier: %v", err)
	}

	s, err := repo.GetSupplierByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetSupplierByID: %v", err)
	}
	if s.Name != "FetchByID Supplier" {
		t.Errorf("expected Name='FetchByID Supplier', got %q", s.Name)
	}
	if s.Email != "fetchbyid@test.local" {
		t.Errorf("expected Email='fetchbyid@test.local', got %q", s.Email)
	}
}
