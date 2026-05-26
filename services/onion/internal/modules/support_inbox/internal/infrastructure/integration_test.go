package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/support_inbox/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const testInboxBusinessID int64 = 1

func TestSupportInbox_ListInbox_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, _, err := repo.ListInbox(context.Background(), testInboxBusinessID, 1, 20)
	if err != nil {
		t.Fatalf("ListInbox: %v", err)
	}
}

func TestSupportInbox_ListInbox_ReturnsTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, total, err := repo.ListInbox(context.Background(), testInboxBusinessID, 1, 20)
	if err != nil {
		t.Fatalf("ListInbox: %v", err)
	}
	if total < 0 {
		t.Errorf("expected total >= 0, got %d", total)
	}
}

func TestSupportInbox_ListInbox_DefaultPage(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, total, err := repo.ListInbox(context.Background(), testInboxBusinessID, 0, 0)
	if err != nil {
		t.Fatalf("ListInbox with zero page/size: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = items
}

func TestSupportInbox_ListInbox_MaxPageSize(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, _, err := repo.ListInbox(context.Background(), testInboxBusinessID, 1, 200)
	if err != nil {
		t.Fatalf("ListInbox with large page size: %v", err)
	}
	if len(items) > 100 {
		t.Errorf("expected at most 100 items, got %d", len(items))
	}
}

func TestSupportInbox_ListInbox_Page2(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, _, err := repo.ListInbox(context.Background(), testInboxBusinessID, 2, 10)
	if err != nil {
		t.Fatalf("ListInbox page 2: %v", err)
	}
}

func TestSupportInbox_ListInbox_UnknownBusiness_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, total, err := repo.ListInbox(context.Background(), 9999999, 1, 20)
	if err != nil {
		t.Fatalf("ListInbox unknown business: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total=0 for unknown business, got %d", total)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for unknown business, got %d", len(items))
	}
}

func TestSupportInbox_UnreadCount_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	count, err := repo.UnreadCount(context.Background(), testInboxBusinessID)
	if err != nil {
		t.Fatalf("UnreadCount: %v", err)
	}
	if count < 0 {
		t.Errorf("expected count >= 0, got %d", count)
	}
}

func TestSupportInbox_UnreadCount_UnknownBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	count, err := repo.UnreadCount(context.Background(), 9999999)
	if err != nil {
		t.Fatalf("UnreadCount unknown business: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for unknown business, got %d", count)
	}
}

func TestSupportInbox_ListInbox_TotalMatchesPage(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items1, total1, err := repo.ListInbox(context.Background(), testInboxBusinessID, 1, 5)
	if err != nil {
		t.Fatalf("ListInbox page 1: %v", err)
	}
	_, total2, err := repo.ListInbox(context.Background(), testInboxBusinessID, 2, 5)
	if err != nil {
		t.Fatalf("ListInbox page 2: %v", err)
	}
	if total1 != total2 {
		t.Errorf("total should be consistent across pages: got %d and %d", total1, total2)
	}
	_ = items1
}
