package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/webhooks/internal/domain"
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
		if w.BusinessID != testutil.OwnerID {
			t.Errorf("expected business_id %d, got %d", testutil.OwnerID, w.BusinessID)
		}
	}
}

func TestWebhooks_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID,
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID,
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID,
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID,
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID,
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
		id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, url, "secret", []string{"test"})
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

func TestWebhooks_Create_SecretStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID,
		"https://example.com/secret-test",
		"mysupersecret",
		[]string{"order.paid"},
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	webhooks, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, w := range webhooks {
		if w.ID == id && w.Secret != "mysupersecret" {
			t.Errorf("expected secret 'mysupersecret', got %s", w.Secret)
		}
	}
}

func TestWebhooks_List_EachHasPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID,
		"https://example.com/id-check-wh",
		"secret",
		[]string{"test"},
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	webhooks, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, w := range webhooks {
		if w.ID <= 0 {
			t.Errorf("webhook has non-positive ID: %d", w.ID)
		}
	}
}

func TestWebhooks_List_EachHasNonEmptyURL(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID,
		"https://example.com/url-check-wh",
		"secret",
		[]string{"test"},
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	webhooks, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, w := range webhooks {
		if w.URL == "" {
			t.Errorf("webhook id=%d has empty URL", w.ID)
		}
	}
}

func TestWebhooks_Delete_UnknownID_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	err := repo.Delete(context.Background(), 9999999, 9999999)
	if !errors.Is(err, domain.ErrWebhookNotFound) {
		t.Fatalf("Delete non-existent error = %v, want %v", err, domain.ErrWebhookNotFound)
	}
}

func TestWebhooks_Create_SingleEvent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID,
		"https://example.com/single-event-wh",
		"secret",
		[]string{"shipment.delivered"},
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	webhooks, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, w := range webhooks {
		if w.ID == id && len(w.Events) != 1 {
			t.Errorf("expected 1 event, got %d", len(w.Events))
		}
	}
}

func TestWebhooks_Create_UniqueIDsForDifferentURLs(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id1, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID,
		"https://example.com/unique-1",
		"secret",
		[]string{"test"},
	)
	if err != nil {
		t.Fatalf("Create 1: %v", err)
	}
	id2, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID,
		"https://example.com/unique-2",
		"secret",
		[]string{"test"},
	)
	if err != nil {
		t.Fatalf("Create 2: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Delete(context.Background(), testutil.OwnerID, id1)
		_ = repo.Delete(context.Background(), testutil.OwnerID, id2)
	})
	if id1 == id2 {
		t.Errorf("expected distinct IDs, got %d and %d", id1, id2)
	}
}

func TestWebhooks_List_AfterDeleteAll_CountDecreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID,
		"https://example.com/delete-count-wh",
		"secret",
		[]string{"test"},
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	before, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before delete: %v", err)
	}

	if err := repo.Delete(context.Background(), testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	after, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	if len(after) >= len(before) {
		t.Errorf("expected fewer webhooks after delete: before=%d after=%d", len(before), len(after))
	}
}

func TestWebhooks_List_ReturnedSliceNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	webhooks, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if webhooks == nil {
		t.Fatal("expected non-nil slice from List")
	}
}

func TestWebhooks_Create_ManyEvents(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	events := []string{"a.b", "c.d", "e.f", "g.h", "i.j"}
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID,
		"https://example.com/many-events-wh",
		"secret",
		events,
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	webhooks, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, w := range webhooks {
		if w.ID == id && len(w.Events) != len(events) {
			t.Errorf("expected %d events, got %d", len(events), len(w.Events))
		}
	}
}

func TestWebhooks_Delete_UnknownIDReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	err := repo.Delete(context.Background(), testutil.OwnerID, 9223372036854775806)
	if !errors.Is(err, domain.ErrWebhookNotFound) {
		t.Fatalf("Delete unknown id error = %v, want ErrWebhookNotFound", err)
	}
}

func TestWebhooks_Update_UnknownIDReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	err := repo.Update(context.Background(), testutil.OwnerID, 9223372036854775806,
		"https://example.com/hook", []string{"order.created"}, true)
	if !errors.Is(err, domain.ErrWebhookNotFound) {
		t.Fatalf("Update unknown id error = %v, want ErrWebhookNotFound", err)
	}
}
