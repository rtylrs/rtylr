package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/feedback/internal/domain"
	"github.com/voxire/onion/internal/modules/feedback/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListFeedback_ReturnsSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, total, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{}, 1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = items
}

func TestListFeedback_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{}, 1, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, f := range items {
		if f.OwnerID != testutil.OwnerID {
			t.Errorf("expected OwnerID=%d, got %d for feedback id=%d", testutil.OwnerID, f.OwnerID, f.ID)
		}
	}
}

func TestCreateFeedback_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FeedbackInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5301,
		Type:       "review",
		Score:      5,
		Comment:    "Integration test feedback",
		Source:     "test",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
	defer func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
	}()
}

func TestCreateFeedback_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FeedbackInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5301,
		Type:       "review",
		Score:      4,
		Comment:    "Appear in list",
		Source:     "direct",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
	}()

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
		t.Errorf("expected feedback id=%d to appear in List", id)
	}
}

func TestListFeedback_FilterByType(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FeedbackInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5301,
		Type:       "nps",
		Score:      8,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
	}()

	items, _, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{Type: "nps"}, 1, 100)
	if err != nil {
		t.Fatalf("List with type filter: %v", err)
	}
	for _, f := range items {
		if f.Type != "nps" {
			t.Errorf("expected Type='nps', got %q for feedback id=%d", f.Type, f.ID)
		}
	}
}

func TestListFeedback_FilterBySource(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FeedbackInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5302,
		Type:       "review",
		Source:     "email",
		Score:      3,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
	}()

	items, _, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{Source: "email"}, 1, 100)
	if err != nil {
		t.Fatalf("List with source filter: %v", err)
	}
	for _, f := range items {
		if f.Source != "email" {
			t.Errorf("expected Source='email', got %q", f.Source)
		}
	}
}

func TestListFeedback_MinScoreFilter(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FeedbackInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5301,
		Type:       "review",
		Score:      9,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
	}()

	items, _, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{MinScore: 8, MinScoreSet: true}, 1, 100)
	if err != nil {
		t.Fatalf("List with min score: %v", err)
	}
	for _, f := range items {
		if f.Score < 8 {
			t.Errorf("expected Score >= 8, got %d for feedback id=%d", f.Score, f.ID)
		}
	}
}

func TestListFeedback_Pagination(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Page size of 1 should limit the result set.
	items, total, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{}, 1, 1)
	if err != nil {
		t.Fatalf("List paginated: %v", err)
	}
	if total < 0 {
		t.Errorf("negative total: %d", total)
	}
	if len(items) > 1 {
		t.Errorf("expected at most 1 item per page, got %d", len(items))
	}
}

func TestListFeedback_IsolatedByOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, _, err := repo.List(ctx, 9999999, domain.ListFilters{}, 1, 100)
	if err != nil {
		t.Fatalf("List other owner: %v", err)
	}
	for _, f := range items {
		if f.OwnerID == testutil.OwnerID {
			t.Errorf("isolation violated: feedback id=%d belongs to OwnerID=%d", f.ID, f.OwnerID)
		}
	}
}
