package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/cashmovements/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/cashmovements/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const (
	fixtureBusinessID int64 = testutil.BusinessID
	fixtureBranchID   int64 = 3001
	fixtureSessionID  int64 = 3201
	fixtureEmployeeID int64 = 3601
)

func TestListCashMovementsBySession_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.ListBySession(ctx, fixtureBusinessID, fixtureSessionID)
	if err != nil {
		t.Fatalf("ListBySession error: %v", err)
	}
}

func TestListCashMovementsBySession_UnknownSessionReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	movements, err := repo.ListBySession(ctx, fixtureBusinessID, 0)
	if err != nil {
		t.Fatalf("ListBySession error: %v", err)
	}
	if len(movements) != 0 {
		t.Errorf("expected empty list for session_id=0, got %d items", len(movements))
	}
}

func TestListCashMovementsBySession_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	movements, err := repo.ListBySession(ctx, fixtureBusinessID, fixtureSessionID)
	if err != nil {
		t.Fatalf("ListBySession error: %v", err)
	}
	for i, m := range movements {
		if m.ID == 0 {
			t.Errorf("movement[%d] has zero ID", i)
		}
	}
}

func TestListCashMovementsBySession_EachSessionIDMatches(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	movements, err := repo.ListBySession(ctx, fixtureBusinessID, fixtureSessionID)
	if err != nil {
		t.Fatalf("ListBySession error: %v", err)
	}
	for i, m := range movements {
		if m.SessionID != fixtureSessionID {
			t.Errorf("movement[%d] has SessionID=%d, want %d", i, m.SessionID, fixtureSessionID)
		}
	}
}

func TestListCashMovementsBySession_EachBusinessIDMatches(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	movements, err := repo.ListBySession(ctx, fixtureBusinessID, fixtureSessionID)
	if err != nil {
		t.Fatalf("ListBySession error: %v", err)
	}
	for i, m := range movements {
		if m.BusinessID != fixtureBusinessID {
			t.Errorf("movement[%d] has BusinessID=%d, want %d", i, m.BusinessID, fixtureBusinessID)
		}
	}
}

func TestCreateCashMovement_InsertsRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.ListBySession(ctx, fixtureBusinessID, fixtureSessionID)
	if err != nil {
		t.Fatalf("ListBySession before: %v", err)
	}

	reason := "opening float"
	input := domain.CreateCashMovementInput{
		BusinessID: fixtureBusinessID,
		BranchID:   fixtureBranchID,
		SessionID:  fixtureSessionID,
		EmployeeID: fixtureEmployeeID,
		Type:       "cash_in",
		Amount:     200.00,
		Reason:     &reason,
	}
	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero inserted ID")
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_cash_movements WHERE id = ?", id)
	})

	after, err := repo.ListBySession(ctx, fixtureBusinessID, fixtureSessionID)
	if err != nil {
		t.Fatalf("ListBySession after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count to grow by 1: before=%d, after=%d", len(before), len(after))
	}
}

func TestCreateCashMovement_TypePersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantType := "cash_out"
	input := domain.CreateCashMovementInput{
		BusinessID: fixtureBusinessID,
		BranchID:   fixtureBranchID,
		SessionID:  fixtureSessionID,
		EmployeeID: fixtureEmployeeID,
		Type:       wantType,
		Amount:     50.00,
	}
	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_cash_movements WHERE id = ?", id)
	})

	movements, err := repo.ListBySession(ctx, fixtureBusinessID, fixtureSessionID)
	if err != nil {
		t.Fatalf("ListBySession: %v", err)
	}
	for _, m := range movements {
		if m.ID == id {
			if m.Type != wantType {
				t.Errorf("Type mismatch: got %q, want %q", m.Type, wantType)
			}
			return
		}
	}
	t.Errorf("inserted movement id=%d not found", id)
}

func TestCreateCashMovement_AmountPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantAmount := 123.45
	input := domain.CreateCashMovementInput{
		BusinessID: fixtureBusinessID,
		BranchID:   fixtureBranchID,
		SessionID:  fixtureSessionID,
		EmployeeID: fixtureEmployeeID,
		Type:       "cash_in",
		Amount:     wantAmount,
	}
	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_cash_movements WHERE id = ?", id)
	})

	movements, err := repo.ListBySession(ctx, fixtureBusinessID, fixtureSessionID)
	if err != nil {
		t.Fatalf("ListBySession: %v", err)
	}
	for _, m := range movements {
		if m.ID == id {
			if m.Amount != wantAmount {
				t.Errorf("Amount mismatch: got %g, want %g", m.Amount, wantAmount)
			}
			return
		}
	}
	t.Errorf("inserted movement id=%d not found", id)
}

func TestListCashMovementsBySession_TwiceReturnsSameCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.ListBySession(ctx, fixtureBusinessID, fixtureSessionID)
	if err != nil {
		t.Fatalf("first ListBySession: %v", err)
	}
	second, err := repo.ListBySession(ctx, fixtureBusinessID, fixtureSessionID)
	if err != nil {
		t.Fatalf("second ListBySession: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("count changed between calls: %d vs %d", len(first), len(second))
	}
}

func TestListCashMovementsBySession_EachHasNonZeroEmployeeID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	movements, err := repo.ListBySession(ctx, fixtureBusinessID, fixtureSessionID)
	if err != nil {
		t.Fatalf("ListBySession error: %v", err)
	}
	for i, m := range movements {
		if m.EmployeeID == 0 {
			t.Errorf("movement[%d] (id=%d) has zero EmployeeID", i, m.ID)
		}
	}
}

func TestListCashMovementsBySession_UnknownBusinessReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	movements, err := repo.ListBySession(ctx, 0, fixtureSessionID)
	if err != nil {
		t.Fatalf("ListBySession error: %v", err)
	}
	if len(movements) != 0 {
		t.Errorf("expected empty for business_id=0, got %d", len(movements))
	}
}
