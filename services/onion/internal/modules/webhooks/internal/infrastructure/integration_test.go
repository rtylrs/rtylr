package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/webhooks/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestWebhooks_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestWebhooks_List_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	webhooks, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, w := range webhooks {
		if w.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, w.OwnerID)
		}
	}
}

func TestWebhooks_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID,
		"https://example.com/webhook",
		"secret123",
		[]string{"order.created", "order.updated"},
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })
}

func TestWebhooks_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID,
		"https://example.com/wh-list-test",
		"secret",
		[]string{"invoice.paid"},
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	webhooks, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, w := range webhooks {
		if w.ID == id {
			found = true
			if w.URL != "https://example.com/wh-list-test" {
				t.Errorf("expected URL https://example.com/wh-list-test, got %s", w.URL)
			}
		}
	}
	if !found {
		t.Errorf("created webhook id=%d not found in list", id)
	}
}

func TestWebhooks_Create_EventsPreserved(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	events := []string{"customer.created", "customer.updated", "customer.deleted"}
	id, err := repo.Create(context.Background(), testutil.OwnerID,
		"https://example.com/events-webhook",
		"secret",
		events,
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	webhooks, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, w := range webhooks {
		if w.ID == id {
			if len(w.Events) != len(events) {
				t.Errorf("expected %d events, got %d", len(events), len(w.Events))
			}
		}
	}
}

func TestWebhooks_Delete_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID,
		"https://example.com/delete-me",
		"secret",
		[]string{"test.event"},
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	webhooks, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, w := range webhooks {
		if w.ID == id {
			t.Errorf("deleted webhook id=%d still in list", id)
		}
	}
}

func TestWebhooks_List_UnknownOwner_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	webhooks, err := repo.List(context.Background(), 9999999)
	if err != nil {
		t.Fatalf("List unknown owner: %v", err)
	}
	if len(webhooks) != 0 {
		t.Errorf("expected 0 for unknown owner, got %d", len(webhooks))
	}
}

func TestWebhooks_Create_EmptyEvents(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID,
		"https://example.com/no-events",
		"secret",
		[]string{},
	)
	if err != nil {
		t.Fatalf("Create with empty events: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestWebhooks_Create_Multiple(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	urls := []string{
		"https://example.com/webhook-a",
		"https://example.com/webhook-b",
		"https://example.com/webhook-c",
	}
	ids := make([]int64, 0, len(urls))
	for _, url := range urls {
		id, err := repo.Create(context.Background(), testutil.OwnerID, url, "secret", []string{"test"})
		if err != nil {
			t.Fatalf("Create %s: %v", url, err)
		}
		ids = append(ids, id)
	}
	t.Cleanup(func() {
		for _, id := range ids {
			_ = repo.Delete(context.Background(), testutil.OwnerID, id)
		}
	})
	if len(ids) != len(urls) {
		t.Errorf("expected %d webhooks, got %d", len(urls), len(ids))
	}
}
