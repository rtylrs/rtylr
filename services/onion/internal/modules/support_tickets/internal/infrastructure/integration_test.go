package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/support_tickets/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// supportTicketsBusinessID matches the fixture businesses row with owner_id=1001.
const supportTicketsBusinessID int64 = 1

func TestList_SupportTickets_NoStatus_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tickets, total, err := repo.List(ctx, supportTicketsBusinessID, "", 1, 20)
	if err != nil {
		t.Fatalf("support_tickets.List: unexpected error: %v", err)
	}

	if tickets == nil {
		t.Fatal("support_tickets.List: expected non-nil slice, got nil")
	}

	if total < 0 {
		t.Fatalf("support_tickets.List: expected non-negative total, got %d", total)
	}
}

func TestList_SupportTickets_OpenStatus_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tickets, _, err := repo.List(ctx, supportTicketsBusinessID, "open", 1, 20)
	if err != nil {
		t.Fatalf("support_tickets.List(status=open): unexpected error: %v", err)
	}

	// All returned tickets must have the requested status.
	for _, tk := range tickets {
		if tk.Status != "open" {
			t.Errorf("support_tickets.List(status=open): got ticket with status=%q, want 'open'", tk.Status)
		}
	}
}

func TestGetByID_SupportTickets_NotFoundReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, supportTicketsBusinessID, 0)
	if err == nil {
		t.Fatal("support_tickets.GetByID(0): expected error for non-existent ticket, got nil")
	}
}
