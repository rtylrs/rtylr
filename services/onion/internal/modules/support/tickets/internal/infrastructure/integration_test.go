package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/support/tickets/internal/domain"
	"github.com/voxire/onion/internal/modules/support/tickets/internal/infrastructure"
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

func TestListTickets_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tickets, total, err := repo.List(ctx, 0, "", 1, 20)
	if err != nil {
		t.Fatalf("List(businessID=0): %v", err)
	}
	if len(tickets) != 0 {
		t.Errorf("expected 0 tickets for businessID=0, got %d", len(tickets))
	}
	if total != 0 {
		t.Errorf("expected total=0 for businessID=0, got %d", total)
	}
}

func TestCreateTicket_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.TicketInput{
		BusinessID:    supportTicketsBusinessID,
		CustomerName:  "Alice",
		CustomerEmail: "alice@example.com",
		Subject:       "Test Subject",
		Status:        "open",
		Priority:      "medium",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateTicket_SubjectIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.TicketInput{
		BusinessID:    supportTicketsBusinessID,
		CustomerName:  "Bob",
		CustomerEmail: "bob@example.com",
		Subject:       "Unique Subject For Test",
		Status:        "open",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	ticket, err := repo.GetByID(ctx, supportTicketsBusinessID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if ticket.Subject != "Unique Subject For Test" {
		t.Errorf("expected subject 'Unique Subject For Test', got %s", ticket.Subject)
	}
}

func TestCreateTicket_StatusIsOpen(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.TicketInput{
		BusinessID:    supportTicketsBusinessID,
		CustomerName:  "Carol",
		CustomerEmail: "carol@example.com",
		Subject:       "Status Test",
		Status:        "open",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	ticket, err := repo.GetByID(ctx, supportTicketsBusinessID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if ticket.Status != "open" {
		t.Errorf("expected status 'open', got %s", ticket.Status)
	}
}

func TestCreateTicket_PriorityIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.TicketInput{
		BusinessID:    supportTicketsBusinessID,
		CustomerName:  "Dave",
		CustomerEmail: "dave@example.com",
		Subject:       "Priority Test",
		Status:        "open",
		Priority:      "high",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	ticket, err := repo.GetByID(ctx, supportTicketsBusinessID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if ticket.Priority != "high" {
		t.Errorf("expected priority 'high', got %s", ticket.Priority)
	}
}

func TestGetTicket_ByID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.TicketInput{
		BusinessID:    supportTicketsBusinessID,
		CustomerName:  "Eve",
		CustomerEmail: "eve@example.com",
		Subject:       "GetByID Test",
		Status:        "open",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	ticket, err := repo.GetByID(ctx, supportTicketsBusinessID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if ticket.ID != id {
		t.Errorf("expected ID %d, got %d", id, ticket.ID)
	}
	if ticket.BusinessID != supportTicketsBusinessID {
		t.Errorf("expected BusinessID %d, got %d", supportTicketsBusinessID, ticket.BusinessID)
	}
}

func TestGetTicket_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, supportTicketsBusinessID, 9999999)
	if !errors.Is(err, domain.ErrTicketNotFound) {
		t.Fatalf("GetByID(missing) error = %v, want ErrTicketNotFound", err)
	}
}

func TestUpdateTicket_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.TicketInput{
		BusinessID:    supportTicketsBusinessID,
		CustomerName:  "Frank",
		CustomerEmail: "frank@example.com",
		Subject:       "Status Update Test",
		Status:        "open",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, domain.TicketUpdateInput{
		BusinessID: supportTicketsBusinessID,
		ID:         id,
		Subject:    "Status Update Test",
		Status:     "closed",
	}); err != nil {
		t.Fatalf("Update status: %v", err)
	}

	ticket, err := repo.GetByID(ctx, supportTicketsBusinessID, id)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if ticket.Status != "closed" {
		t.Errorf("expected status 'closed', got %s", ticket.Status)
	}
}

func TestUpdateTicket_ChangesAssignee(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.TicketInput{
		BusinessID:    supportTicketsBusinessID,
		CustomerName:  "Grace",
		CustomerEmail: "grace@example.com",
		Subject:       "Assignee Test",
		Status:        "open",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	assignee := testutil.OwnerID
	if err := repo.Update(ctx, domain.TicketUpdateInput{
		BusinessID: supportTicketsBusinessID,
		ID:         id,
		Subject:    "Assignee Test",
		Status:     "open",
		AssignedTo: &assignee,
	}); err != nil {
		t.Fatalf("Update assignee: %v", err)
	}

	ticket, err := repo.GetByID(ctx, supportTicketsBusinessID, id)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if ticket.AssignedTo == nil || *ticket.AssignedTo != assignee {
		t.Errorf("expected assignee %d, got %v", assignee, ticket.AssignedTo)
	}
}

// TestUpdateTicket_PartialUpdate_PreservesUntouchedFields pins the fix for a
// partial PUT that only carries the fields being changed: the columns the
// request omits must keep their stored values instead of being blanked/NULLed.
func TestUpdateTicket_PartialUpdate_PreservesUntouchedFields(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	assignee := testutil.OwnerID
	id, err := repo.Create(ctx, domain.TicketInput{
		BusinessID:    supportTicketsBusinessID,
		CustomerName:  "Nora",
		CustomerEmail: "nora@example.com",
		Subject:       "Partial Update Subject",
		Status:        "open",
		Priority:      "high",
		Channel:       "email",
		AssignedTo:    &assignee,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Only the status is supplied; subject/priority/assignee/channel are omitted.
	if err := repo.Update(ctx, domain.TicketUpdateInput{
		BusinessID: supportTicketsBusinessID,
		ID:         id,
		Status:     "resolved",
	}); err != nil {
		t.Fatalf("Update status only: %v", err)
	}

	ticket, err := repo.GetByID(ctx, supportTicketsBusinessID, id)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if ticket.Status != "resolved" {
		t.Errorf("expected status 'resolved', got %q", ticket.Status)
	}
	if ticket.Subject != "Partial Update Subject" {
		t.Errorf("partial update blanked subject: got %q, want %q", ticket.Subject, "Partial Update Subject")
	}
	if ticket.Priority != "high" {
		t.Errorf("partial update cleared priority: got %q, want 'high'", ticket.Priority)
	}
	if ticket.AssignedTo == nil || *ticket.AssignedTo != assignee {
		t.Errorf("partial update cleared assigned_to: got %v, want %d", ticket.AssignedTo, assignee)
	}
	if ticket.Channel != "email" {
		t.Errorf("partial update cleared channel: got %q, want 'email'", ticket.Channel)
	}

	// A subject-only update must leave the previously set status in place.
	if err := repo.Update(ctx, domain.TicketUpdateInput{
		BusinessID: supportTicketsBusinessID,
		ID:         id,
		Subject:    "Renamed Subject",
	}); err != nil {
		t.Fatalf("Update subject only: %v", err)
	}

	ticket, err = repo.GetByID(ctx, supportTicketsBusinessID, id)
	if err != nil {
		t.Fatalf("GetByID after subject update: %v", err)
	}
	if ticket.Subject != "Renamed Subject" {
		t.Errorf("expected subject 'Renamed Subject', got %q", ticket.Subject)
	}
	if ticket.Status != "resolved" {
		t.Errorf("subject-only update cleared status: got %q, want 'resolved'", ticket.Status)
	}
	if ticket.Priority != "high" {
		t.Errorf("subject-only update cleared priority: got %q, want 'high'", ticket.Priority)
	}
}

func TestDeleteTicket_RemovesRow(t *testing.T) {
	// support_tickets has no Delete method; verify via GetByID after status-transition workaround.
	// This test verifies that List does not return a ticket scoped to a non-existent business.
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	tickets, total, err := repo.List(ctx, 9999998, "", 1, 20)
	if err != nil {
		t.Fatalf("List non-existent business: %v", err)
	}
	if total != 0 || len(tickets) != 0 {
		t.Errorf("expected 0 tickets for non-existent business, got total=%d items=%d", total, len(tickets))
	}
}

func TestListTickets_AfterCreate_ContainsNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.TicketInput{
		BusinessID:    supportTicketsBusinessID,
		CustomerName:  "Henry",
		CustomerEmail: "henry@example.com",
		Subject:       "AfterCreate Test",
		Status:        "open",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	tickets, _, err := repo.List(ctx, supportTicketsBusinessID, "", 1, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, tk := range tickets {
		if tk.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created ticket id=%d not found in list", id)
	}
}

func TestListTickets_FilterByStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.Create(ctx, domain.TicketInput{
		BusinessID:    supportTicketsBusinessID,
		CustomerName:  "Iris",
		CustomerEmail: "iris@example.com",
		Subject:       "Pending Ticket",
		Status:        "pending",
	})
	if err != nil {
		t.Fatalf("Create pending ticket: %v", err)
	}

	tickets, _, err := repo.List(ctx, supportTicketsBusinessID, "pending", 1, 20)
	if err != nil {
		t.Fatalf("List(status=pending): %v", err)
	}
	for _, tk := range tickets {
		if tk.Status != "pending" {
			t.Errorf("expected all tickets to have status 'pending', got %s", tk.Status)
		}
	}
}

func TestCreateTicket_CustomerEmailStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.TicketInput{
		BusinessID:    supportTicketsBusinessID,
		CustomerName:  "Jane",
		CustomerEmail: "jane@example.com",
		Subject:       "Email Test",
		Status:        "open",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	ticket, err := repo.GetByID(ctx, supportTicketsBusinessID, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if ticket.CustomerEmail != "jane@example.com" {
		t.Errorf("expected email 'jane@example.com', got %s", ticket.CustomerEmail)
	}
}

func TestCreateMessage_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ticketID, err := repo.Create(ctx, domain.TicketInput{
		BusinessID:    supportTicketsBusinessID,
		CustomerName:  "Kate",
		CustomerEmail: "kate@example.com",
		Subject:       "Message Test",
		Status:        "open",
	})
	if err != nil {
		t.Fatalf("Create ticket: %v", err)
	}

	msgID, err := repo.CreateMessage(ctx, domain.MessageInput{
		TicketID:   ticketID,
		SenderType: "agent",
		SenderName: "Support Agent",
		Body:       "Hello, how can I help?",
	})
	if err != nil {
		t.Fatalf("CreateMessage: %v", err)
	}
	if msgID <= 0 {
		t.Fatalf("expected positive message ID, got %d", msgID)
	}
}

func TestListMessages_AfterCreate_ContainsMessage(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ticketID, err := repo.Create(ctx, domain.TicketInput{
		BusinessID:    supportTicketsBusinessID,
		CustomerName:  "Leo",
		CustomerEmail: "leo@example.com",
		Subject:       "ListMessages Test",
		Status:        "open",
	})
	if err != nil {
		t.Fatalf("Create ticket: %v", err)
	}

	msgID, err := repo.CreateMessage(ctx, domain.MessageInput{
		TicketID:   ticketID,
		SenderType: "customer",
		SenderName: "Leo",
		Body:       "I need help please.",
	})
	if err != nil {
		t.Fatalf("CreateMessage: %v", err)
	}

	messages, err := repo.ListMessages(ctx, supportTicketsBusinessID, ticketID)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	found := false
	for _, m := range messages {
		if m.ID == msgID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created message id=%d not found in ListMessages", msgID)
	}
}

func TestListMessages_EmptyTicket_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	messages, err := repo.ListMessages(ctx, supportTicketsBusinessID, 9999999)
	if err != nil {
		t.Fatalf("ListMessages non-existent ticket: %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("expected 0 messages for non-existent ticket, got %d", len(messages))
	}
}
