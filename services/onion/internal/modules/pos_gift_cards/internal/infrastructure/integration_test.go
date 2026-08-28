package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_gift_cards/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListGiftCards_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
}

func TestListGiftCards_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cards, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(cards) != 0 {
		t.Errorf("expected empty list for owner_id=0, got %d cards", len(cards))
	}
}

func TestListGiftCards_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cards, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, c := range cards {
		if c.ID == 0 {
			t.Errorf("card[%d] has zero ID", i)
		}
	}
}

func TestListGiftCards_EachOwnerIDMatches(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cards, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, c := range cards {
		if c.OwnerID != testutil.OwnerID {
			t.Errorf("card[%d] has OwnerID=%d, want %d", i, c.OwnerID, testutil.OwnerID)
		}
	}
}

func TestListGiftCards_EachHasNonEmptyCode(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cards, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, c := range cards {
		if c.Code == "" {
			t.Errorf("card[%d] (id=%d) has empty Code", i, c.ID)
		}
	}
}

func TestCreateGiftCard_InsertsRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	input := domain.GiftCardInput{
		PaidOrderID:   3301,
		Code:          "INTTEST-001",
		InitialAmount: 50.00,
		CustomerName:  "Test Customer",
		CustomerEmail: "test@example.com",
	}
	id, err := repo.CreateForOwner(ctx, testutil.BusinessID, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero ID")
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_gift_card_transactions WHERE gift_card_id = ?", id)
		db.ExecContext(ctx, "DELETE FROM pos_gift_cards WHERE id = ?", id)
	})

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count to grow by 1: before=%d, after=%d", len(before), len(after))
	}
}

func TestGetByCode_ReturnsCard(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.GiftCardInput{
		PaidOrderID:   3301,
		Code:          "INTTEST-002",
		InitialAmount: 100.00,
	}
	id, err := repo.CreateForOwner(ctx, testutil.BusinessID, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_gift_card_transactions WHERE gift_card_id = ?", id)
		db.ExecContext(ctx, "DELETE FROM pos_gift_cards WHERE id = ?", id)
	})

	card, err := repo.GetByCode(ctx, testutil.OwnerID, "INTTEST-002")
	if err != nil {
		t.Fatalf("GetByCode error: %v", err)
	}
	if card.Code != "INTTEST-002" {
		t.Errorf("Code mismatch: got %q", card.Code)
	}
	if card.InitialAmount != 100.00 {
		t.Errorf("InitialAmount mismatch: got %g, want 100.00", card.InitialAmount)
	}
}

func TestGetByCode_UnknownCodeReturnsErrGiftCardNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByCode(ctx, testutil.OwnerID, "NO-SUCH-CODE-XYZ")
	if !errors.Is(err, domain.ErrGiftCardNotFound) {
		t.Errorf("expected ErrGiftCardNotFound, got %v", err)
	}
}

func TestRedeem_ReducesBalance(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.GiftCardInput{
		PaidOrderID:   3301,
		Code:          "INTTEST-003",
		InitialAmount: 75.00,
	}
	id, err := repo.CreateForOwner(ctx, testutil.BusinessID, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_gift_card_transactions WHERE gift_card_id = ?", id)
		db.ExecContext(ctx, "DELETE FROM pos_gift_cards WHERE id = ?", id)
	})

	updated, err := repo.RedeemForOrder(ctx, testutil.BusinessID, testutil.OwnerID, domain.RedeemInput{
		OrderID: 3301,
		Code:   "INTTEST-003",
		Amount: 25.00,
	})
	if err != nil {
		t.Fatalf("Redeem error: %v", err)
	}
	const wantBalance = 50.00
	if updated.CurrentBalance != wantBalance {
		t.Errorf("CurrentBalance after redeem: got %g, want %g", updated.CurrentBalance, wantBalance)
	}
}

func TestRedeem_InsufficientBalanceErrors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.GiftCardInput{
		PaidOrderID:   3301,
		Code:          "INTTEST-004",
		InitialAmount: 10.00,
	}
	id, err := repo.CreateForOwner(ctx, testutil.BusinessID, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_gift_card_transactions WHERE gift_card_id = ?", id)
		db.ExecContext(ctx, "DELETE FROM pos_gift_cards WHERE id = ?", id)
	})

	_, err = repo.RedeemForOrder(ctx, testutil.BusinessID, testutil.OwnerID, domain.RedeemInput{
		OrderID: 3301,
		Code:   "INTTEST-004",
		Amount: 100.00,
	})
	if !errors.Is(err, domain.ErrInsufficientBalance) {
		t.Errorf("expected ErrInsufficientBalance, got %v", err)
	}
}

func TestListGiftCards_TwiceReturnsSameCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first List: %v", err)
	}
	second, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second List: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("count changed: %d vs %d", len(first), len(second))
	}
}

func TestCreateGiftCard_WithExpiry_Persisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.GiftCardInput{
		PaidOrderID:   3301,
		Code:          "INTTEST-005",
		InitialAmount: 20.00,
		ExpiresAt:     "2030-12-31",
	}
	id, err := repo.CreateForOwner(ctx, testutil.BusinessID, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_gift_card_transactions WHERE gift_card_id = ?", id)
		db.ExecContext(ctx, "DELETE FROM pos_gift_cards WHERE id = ?", id)
	})

	card, err := repo.GetByCode(ctx, testutil.OwnerID, "INTTEST-005")
	if err != nil {
		t.Fatalf("GetByCode error: %v", err)
	}
	if card.ExpiresAt == nil {
		t.Error("expected ExpiresAt to be set, got nil")
	}
}

func TestListGiftCards_InitialAmountEqualsCurrentBalance_ForNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.GiftCardInput{
		PaidOrderID:   3301,
		Code:          "INTTEST-006",
		InitialAmount: 60.00,
	}
	id, err := repo.CreateForOwner(ctx, testutil.BusinessID, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_gift_card_transactions WHERE gift_card_id = ?", id)
		db.ExecContext(ctx, "DELETE FROM pos_gift_cards WHERE id = ?", id)
	})

	cards, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for _, c := range cards {
		if c.ID == id {
			if c.InitialAmount != c.CurrentBalance {
				t.Errorf("new card: InitialAmount=%g != CurrentBalance=%g", c.InitialAmount, c.CurrentBalance)
			}
			return
		}
	}
	t.Errorf("inserted card id=%d not found", id)
}
