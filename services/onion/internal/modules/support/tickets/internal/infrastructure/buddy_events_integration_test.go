package infrastructure

import (
	"context"
	"testing"

	supportcontract "github.com/voxire/onion/internal/modules/support/contract"
	"github.com/voxire/onion/internal/modules/support/tickets/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestTicketAndCustomerMessagePublishFacts(t *testing.T) {
	db := testutil.OpenDB(t)
	const business = int64(920598)
	clean := func() {
		for _, q := range []string{"DELETE FROM support_messages WHERE ticket_id IN (SELECT id FROM support_tickets WHERE business_id = ?)", "DELETE FROM support_tickets WHERE business_id = ?", "DELETE FROM support_customers WHERE business_id = ?", "DELETE FROM support_outbox WHERE business_id = ?"} {
			if _, err := db.Exec(q, business); err != nil {
				t.Fatal(err)
			}
		}
	}
	clean()
	t.Cleanup(clean)
	ctx := context.Background()
	repo := NewRepository(db)
	id, err := repo.Create(ctx, domain.TicketInput{BusinessID: business, CustomerName: "Example", CustomerEmail: "example@example.test", Subject: "Help"})
	if err != nil {
		t.Fatal(err)
	}
	for _, sender := range []string{"customer", "agent"} {
		if _, err := repo.CreateMessage(ctx, domain.MessageInput{TicketID: id, SenderType: sender, SenderName: "Example", Body: "Hello"}); err != nil {
			t.Fatal(err)
		}
	}
	var ticketCount, messageCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM support_outbox WHERE business_id = ? AND event_type = ? AND subject_id = ?", business, supportcontract.TicketCreated, id).Scan(&ticketCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM support_outbox WHERE business_id = ? AND event_type = ?", business, supportcontract.CustomerMessageCreated).Scan(&messageCount); err != nil {
		t.Fatal(err)
	}
	if ticketCount != 1 || messageCount != 1 {
		t.Fatalf("facts: ticket=%d customer messages=%d", ticketCount, messageCount)
	}
}
