package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/expenses/internal/domain"
	"github.com/voxire/onion/internal/modules/expenses/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListExpenses_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	expenses, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List expenses error: %v", err)
	}
	// May be empty for fixture data; assert no error and valid return.
	_ = expenses
}

func TestListExpenseCategories_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	cats, err := repo.ListCategories(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListCategories error: %v", err)
	}
	// May be empty for fixture data; assert no error and valid return.
	_ = cats
}

func TestListExpenses_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	expenses, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List(0): %v", err)
	}
	if len(expenses) != 0 {
		t.Fatalf("expected 0 expenses for owner 0, got %d", len(expenses))
	}
}

func TestListExpenses_WrongOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	expenses, err := repo.List(ctx, 999999)
	if err != nil {
		t.Fatalf("List wrong owner: %v", err)
	}
	if len(expenses) != 0 {
		t.Fatalf("expected 0 expenses for unknown owner, got %d", len(expenses))
	}
}

func TestCreateExpense_Persists(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ExpenseInput{
		OwnerID:     testutil.OwnerID,
		Amount:      99.99,
		Currency:    "USD",
		Description: "Test expense",
		Vendor:      "Test Vendor",
		ExpenseDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateExpense_AmountStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ExpenseInput{
		OwnerID:     testutil.OwnerID,
		Amount:      123.45,
		Currency:    "USD",
		Description: "Amount check",
		ExpenseDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	expense, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if expense.Amount != 123.45 {
		t.Errorf("Amount = %f, want 123.45", expense.Amount)
	}
}

func TestCreateExpense_VendorStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	vendor := "Acme Corp"
	id, err := repo.Create(ctx, domain.ExpenseInput{
		OwnerID:     testutil.OwnerID,
		Amount:      50.00,
		Currency:    "USD",
		Description: "Vendor check",
		Vendor:      vendor,
		ExpenseDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	expense, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if expense.Vendor != vendor {
		t.Errorf("Vendor = %q, want %q", expense.Vendor, vendor)
	}
}

func TestGetExpenseByID_ReturnsCorrectRecord(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ExpenseInput{
		OwnerID:     testutil.OwnerID,
		Amount:      77.77,
		Currency:    "USD",
		Description: "GetByID check",
		ExpenseDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	expense, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if expense.ID != id {
		t.Errorf("ID = %d, want %d", expense.ID, id)
	}
}

func TestGetExpenseByID_NotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, testutil.OwnerID, 999999999)
	if !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("expected ErrExpenseNotFound, got %v", err)
	}
}

func TestGetExpenseByID_WrongOwner_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ExpenseInput{
		OwnerID:     testutil.OwnerID,
		Amount:      10.00,
		Currency:    "USD",
		Description: "Wrong owner check",
		ExpenseDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = repo.GetByID(ctx, 999999, id)
	if !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("expected ErrExpenseNotFound for wrong owner, got %v", err)
	}
}

func TestDeleteExpense_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ExpenseInput{
		OwnerID:     testutil.OwnerID,
		Amount:      15.00,
		Currency:    "USD",
		Description: "To be deleted",
		ExpenseDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = repo.GetByID(ctx, testutil.OwnerID, id)
	if !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("expected ErrExpenseNotFound after delete, got %v", err)
	}
}

func TestDeleteExpense_WrongOwner_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ExpenseInput{
		OwnerID:     testutil.OwnerID,
		Amount:      25.00,
		Currency:    "USD",
		Description: "Wrong owner delete",
		ExpenseDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Delete(ctx, 999999, id)
	if !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("expected ErrExpenseNotFound for wrong-owner delete, got %v", err)
	}
}

func TestUpdateExpenseStatus_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ExpenseInput{
		OwnerID:     testutil.OwnerID,
		Amount:      30.00,
		Currency:    "USD",
		Description: "Status update",
		ExpenseDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.UpdateStatus(ctx, testutil.OwnerID, id, "approved"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	expense, err := repo.GetByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetByID after UpdateStatus: %v", err)
	}
	if expense.Status != "approved" {
		t.Errorf("Status = %q, want %q", expense.Status, "approved")
	}
}

func TestUpdateExpenseStatus_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpdateStatus(ctx, testutil.OwnerID, 999999999, "approved")
	if !errors.Is(err, domain.ErrExpenseNotFound) {
		t.Errorf("expected ErrExpenseNotFound, got %v", err)
	}
}

func TestListExpenses_OwnerIDScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, domain.ExpenseInput{
		OwnerID:     testutil.OwnerID,
		Amount:      5.00,
		Currency:    "USD",
		Description: "Owner scope check",
		ExpenseDate: time.Now(),
	})

	expenses, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, e := range expenses {
		if e.OwnerID != testutil.OwnerID {
			t.Errorf("expenses[%d].OwnerID = %d, want %d", i, e.OwnerID, testutil.OwnerID)
		}
	}
}

func TestListCategories_ReturnsRows(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cats, err := repo.ListCategories(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	// Validate structure of any returned categories.
	for i, c := range cats {
		if c.ID == 0 {
			t.Errorf("cats[%d].ID is 0", i)
		}
		if c.Name == "" {
			t.Errorf("cats[%d].Name is empty", i)
		}
	}
}

func TestListExpenseCategories_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cats, err := repo.ListCategories(ctx, 0)
	if err != nil {
		t.Fatalf("ListCategories(0): %v", err)
	}
	if len(cats) != 0 {
		t.Fatalf("expected 0 categories for owner 0, got %d", len(cats))
	}
}

func TestCreateExpense_ThenAppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ExpenseInput{
		OwnerID:     testutil.OwnerID,
		Amount:      42.00,
		Currency:    "USD",
		Description: "Appears in list",
		ExpenseDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	expenses, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	found := false
	for _, e := range expenses {
		if e.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expense id=%d not found in list", id)
	}
}

func TestSumByOwner_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	total, err := repo.SumByOwner(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("SumByOwner: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %f", total)
	}
}

func TestSumByOwner_ZeroOwnerReturnsZero(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	total, err := repo.SumByOwner(ctx, 0)
	if err != nil {
		t.Fatalf("SumByOwner(ownerID=0): %v", err)
	}
	if total != 0 {
		t.Errorf("expected 0 for ownerID=0, got %f", total)
	}
}

func TestSumByOwner_ReflectsCreatedExpense(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ExpenseInput{
		OwnerID:     testutil.OwnerID,
		Amount:      250.00,
		Currency:    "USD",
		Description: "SumByOwner test expense",
		ExpenseDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	total, err := repo.SumByOwner(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("SumByOwner: %v", err)
	}
	if total < 250.00 {
		t.Errorf("expected total >= 250.00, got %f", total)
	}
}

func TestCountByCategory_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	counts, err := repo.CountByCategory(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountByCategory: %v", err)
	}
	_ = counts
}

func TestCountByCategory_ZeroOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	counts, err := repo.CountByCategory(ctx, 0)
	if err != nil {
		t.Fatalf("CountByCategory(ownerID=0): %v", err)
	}
	if len(counts) != 0 {
		t.Errorf("expected empty map for ownerID=0, got %v", counts)
	}
}

func TestCountByCategory_UncategorisedBucket(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ExpenseInput{
		OwnerID:     testutil.OwnerID,
		Amount:      15.00,
		Currency:    "USD",
		Description: "Uncategorised expense",
		ExpenseDate: time.Now(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	counts, err := repo.CountByCategory(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("CountByCategory: %v", err)
	}
	if counts["uncategorised"] <= 0 {
		t.Errorf("expected at least 1 uncategorised expense, got %d", counts["uncategorised"])
	}
}
