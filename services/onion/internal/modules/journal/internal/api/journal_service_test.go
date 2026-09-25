package api

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/journal/internal/domain"
)

func TestNewJournalEntryPayloadIncludesAccountName(t *testing.T) {
	payload := newJournalEntryPayload(domain.JournalEntry{
		ID:          1,
		Date:        time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC),
		Description: "Balanced entry",
		Lines: []domain.JournalLine{{
			AccountID:   10,
			AccountName: "Cash",
			Debit:       100,
		}},
	})

	if got := payload.GetLines()[0].GetAccountName(); got != "Cash" {
		t.Fatalf("account name = %q, want Cash", got)
	}
}
