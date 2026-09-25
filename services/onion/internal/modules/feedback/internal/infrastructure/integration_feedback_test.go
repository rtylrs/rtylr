package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/feedback/internal/domain"
	"github.com/voxire/onion/internal/modules/feedback/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListFeedback_TypeAndSourceCombined(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FeedbackInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5301,
		Type:       "nps",
		Source:     "sms",
		Score:      7,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
	})

	items, _, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{Type: "nps", Source: "sms"}, 1, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, f := range items {
		if f.Type != "nps" || f.Source != "sms" {
			t.Errorf("unexpected item type=%q source=%q for id=%d", f.Type, f.Source, f.ID)
		}
	}
}

func TestCreateFeedback_WithZeroCustomerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FeedbackInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Type:       "review",
		Score:      3,
		Comment:    "anonymous feedback",
	})
	if err != nil {
		t.Fatalf("Create with zero CustomerID: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
	})
}

func TestListFeedback_PaginationPageSize5(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, total, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{}, 1, 5)
	if err != nil {
		t.Fatalf("List pageSize=5: %v", err)
	}
	if len(items) > 5 {
		t.Errorf("expected at most 5 items, got %d", len(items))
	}
	if total < 0 {
		t.Errorf("negative total: %d", total)
	}
}

func TestCreateFeedback_ScoreZeroAllowed(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FeedbackInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5301,
		Type:       "review",
		Score:      0,
	})
	if err != nil {
		t.Fatalf("Create score=0: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
	})
}

func TestListFeedback_EmptySourceFilterReturnsAll(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FeedbackInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5301,
		Type:       "review",
		Score:      4,
		Source:     "web",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
	})

	// Source="" means no filter applied, so the created item must appear.
	items, _, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{}, 1, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, f := range items {
		if f.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected feedback id=%d to appear without source filter", id)
	}
}

func TestListFeedback_UnknownSourceReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, total, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{Source: "no_such_source_xyz"}, 1, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for unknown source, got %d", len(items))
	}
	if total != 0 {
		t.Errorf("expected total=0 for unknown source, got %d", total)
	}
}

func TestListFeedback_UnknownTypeReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, total, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{Type: "no_such_type_xyz"}, 1, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for unknown type, got %d", len(items))
	}
	if total != 0 {
		t.Errorf("expected total=0 for unknown type, got %d", total)
	}
}
