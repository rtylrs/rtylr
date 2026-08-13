package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/support_inbox/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const testInboxBusinessID int64 = 1

func TestSupportInbox_ListInbox_MessageSummaryHandlesTiesEmptyHistoriesAndPagination(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	businessID := time.Now().UnixNano()
	baseTime := time.Now().UTC().Truncate(time.Second)

	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, "DELETE FROM support_tickets WHERE business_id = ?", businessID); err != nil {
			t.Errorf("clean up support tickets: %v", err)
		}
	})

	insertTicket := func(subject string, updatedAt time.Time) int64 {
		t.Helper()
		result, err := db.ExecContext(ctx, `
			INSERT INTO support_tickets
				(business_id, customer_name, customer_email, subject, status, priority, channel, updated_at)
			VALUES (?, ?, ?, ?, 'open', 'medium', 'email', ?)`,
			businessID, "Inbox test", fmt.Sprintf("%s@example.test", subject), subject, updatedAt)
		if err != nil {
			t.Fatalf("insert ticket %q: %v", subject, err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("ticket ID for %q: %v", subject, err)
		}
		return id
	}

	insertMessage := func(ticketID int64, body string, createdAt time.Time) {
		t.Helper()
		if _, err := db.ExecContext(ctx, `
			INSERT INTO support_messages (ticket_id, sender_type, sender_name, body, created_at)
			VALUES (?, 'customer', 'Inbox test', ?, ?)`, ticketID, body, createdAt); err != nil {
			t.Fatalf("insert message for ticket %d: %v", ticketID, err)
		}
	}

	newestTicketID := insertTicket("newest", baseTime.Add(3*time.Minute))
	tiedTicketID := insertTicket("tied", baseTime.Add(2*time.Minute))
	emptyTicketID := insertTicket("empty", baseTime.Add(time.Minute))
	insertMessage(newestTicketID, "older", baseTime)
	insertMessage(newestTicketID, "newest", baseTime.Add(time.Minute))
	insertMessage(tiedTicketID, "first at tie", baseTime)
	insertMessage(tiedTicketID, "second at tie", baseTime)

	repo := infrastructure.NewRepository(db)
	pageOne, total, err := repo.ListInbox(ctx, businessID, 1, 2)
	if err != nil {
		t.Fatalf("ListInbox page 1: %v", err)
	}
	if total != 3 {
		t.Fatalf("total = %d, want 3", total)
	}
	if len(pageOne) != 2 {
		t.Fatalf("page 1 items = %d, want 2", len(pageOne))
	}
	if pageOne[0].TicketID != newestTicketID || pageOne[0].LastMessage != "newest" || pageOne[0].MessageCount != 2 {
		t.Fatalf("newest ticket summary = %#v", pageOne[0])
	}
	if pageOne[1].TicketID != tiedTicketID || pageOne[1].LastMessage != "second at tie" || pageOne[1].MessageCount != 2 {
		t.Fatalf("tied ticket summary = %#v", pageOne[1])
	}

	pageTwo, total, err := repo.ListInbox(ctx, businessID, 2, 2)
	if err != nil {
		t.Fatalf("ListInbox page 2: %v", err)
	}
	if total != 3 || len(pageTwo) != 1 {
		t.Fatalf("page 2 = %#v, total = %d", pageTwo, total)
	}
	if pageTwo[0].TicketID != emptyTicketID || pageTwo[0].LastMessage != "" || !pageTwo[0].LastMessageAt.IsZero() || pageTwo[0].MessageCount != 0 {
		t.Fatalf("empty ticket summary = %#v", pageTwo[0])
	}
}

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

func TestSupportInbox_ListInbox_PageSize1_ReturnsAtMostOne(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, _, err := repo.ListInbox(context.Background(), testInboxBusinessID, 1, 1)
	if err != nil {
		t.Fatalf("ListInbox pageSize=1: %v", err)
	}
	if len(items) > 1 {
		t.Errorf("expected at most 1 item, got %d", len(items))
	}
}

func TestSupportInbox_ListInbox_ReturnsSliceNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, _, err := repo.ListInbox(context.Background(), testInboxBusinessID, 1, 20)
	if err != nil {
		t.Fatalf("ListInbox: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestSupportInbox_ListInbox_Page3_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, total, err := repo.ListInbox(context.Background(), testInboxBusinessID, 3, 5)
	if err != nil {
		t.Fatalf("ListInbox page 3: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
}

func TestSupportInbox_UnreadCount_IsNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	count, err := repo.UnreadCount(context.Background(), testInboxBusinessID)
	if err != nil {
		t.Fatalf("UnreadCount: %v", err)
	}
	if count < 0 {
		t.Errorf("UnreadCount should be >= 0, got %d", count)
	}
}

func TestSupportInbox_ListInbox_TotalConsistentWithPageSize10(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, total10, err := repo.ListInbox(context.Background(), testInboxBusinessID, 1, 10)
	if err != nil {
		t.Fatalf("ListInbox pageSize=10: %v", err)
	}
	_, total20, err := repo.ListInbox(context.Background(), testInboxBusinessID, 1, 20)
	if err != nil {
		t.Fatalf("ListInbox pageSize=20: %v", err)
	}
	if total10 != total20 {
		t.Errorf("total should be the same regardless of page size: got %d and %d", total10, total20)
	}
}

func TestSupportInbox_ListInbox_NegativePage_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, _, err := repo.ListInbox(context.Background(), testInboxBusinessID, -1, 10)
	if err != nil {
		t.Fatalf("ListInbox negative page: %v", err)
	}
}

func TestSupportInbox_ListInbox_ZeroPageSize_UsesDefault(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, total, err := repo.ListInbox(context.Background(), testInboxBusinessID, 1, 0)
	if err != nil {
		t.Fatalf("ListInbox pageSize=0: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = items
}

func TestSupportInbox_UnreadCount_ConsistentOnRepeatCall(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	c1, err := repo.UnreadCount(context.Background(), testInboxBusinessID)
	if err != nil {
		t.Fatalf("UnreadCount first call: %v", err)
	}
	c2, err := repo.UnreadCount(context.Background(), testInboxBusinessID)
	if err != nil {
		t.Fatalf("UnreadCount second call: %v", err)
	}
	if c1 != c2 {
		t.Errorf("UnreadCount inconsistent across calls: %d vs %d", c1, c2)
	}
}

func TestSupportInbox_ListInbox_PageBeyondTotal_ReturnsEmptyItems(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, total, err := repo.ListInbox(context.Background(), testInboxBusinessID, 1, 20)
	if err != nil {
		t.Fatalf("ListInbox: %v", err)
	}
	// Request a page well beyond the total.
	beyondPage := (total/20 + 100)
	items, _, err := repo.ListInbox(context.Background(), testInboxBusinessID, beyondPage, 20)
	if err != nil {
		t.Fatalf("ListInbox beyond total: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items on page far beyond total, got %d", len(items))
	}
}

func TestSupportInbox_ListInbox_ItemCountBoundedByPageSize(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const pageSize = 3
	items, _, err := repo.ListInbox(context.Background(), testInboxBusinessID, 1, pageSize)
	if err != nil {
		t.Fatalf("ListInbox pageSize=%d: %v", pageSize, err)
	}
	if len(items) > pageSize {
		t.Errorf("expected at most %d items, got %d", pageSize, len(items))
	}
}
