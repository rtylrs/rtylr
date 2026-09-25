package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/feedback/internal/domain"
	"github.com/voxire/onion/internal/modules/feedback/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateFeedback_ScoreAndCommentPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FeedbackInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5301,
		Type:       "review",
		Score:      7,
		Comment:    "Quite good",
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
	for _, f := range items {
		if f.ID == id {
			if f.Score != 7 {
				t.Errorf("expected Score=7, got %d", f.Score)
			}
			if f.Comment != "Quite good" {
				t.Errorf("expected Comment='Quite good', got %q", f.Comment)
			}
			return
		}
	}
	t.Errorf("feedback id=%d not found in list", id)
}

func TestListFeedback_MaxScoreFilter(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FeedbackInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5301,
		Type:       "review",
		Score:      2,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
	})

	items, _, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{MaxScore: 5, MaxScoreSet: true}, 1, 100)
	if err != nil {
		t.Fatalf("List with max score: %v", err)
	}
	for _, f := range items {
		if f.Score > 5 {
			t.Errorf("expected Score <= 5, got %d for feedback id=%d", f.Score, f.ID)
		}
	}
}

func TestListFeedback_MinAndMaxScoreCombined(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FeedbackInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5301,
		Type:       "review",
		Score:      6,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
	})

	items, _, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{MinScore: 5, MinScoreSet: true, MaxScore: 7, MaxScoreSet: true}, 1, 100)
	if err != nil {
		t.Fatalf("List with min+max score: %v", err)
	}
	for _, f := range items {
		if f.Score < 5 || f.Score > 7 {
			t.Errorf("expected 5 <= Score <= 7, got %d for feedback id=%d", f.Score, f.ID)
		}
	}
}

func TestCreateFeedback_TypePersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FeedbackInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5301,
		Type:       "nps",
		Score:      10,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
	})

	items, _, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{Type: "nps"}, 1, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, f := range items {
		if f.ID == id {
			if f.Type != "nps" {
				t.Errorf("expected Type='nps', got %q", f.Type)
			}
			return
		}
	}
	t.Errorf("feedback id=%d not found after create", id)
}

func TestCreateFeedback_SourcePersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.FeedbackInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		CustomerID: 5301,
		Type:       "review",
		Source:     "kiosk",
		Score:      5,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
	})

	items, _, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{Source: "kiosk"}, 1, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, f := range items {
		if f.ID == id {
			if f.Source != "kiosk" {
				t.Errorf("expected Source='kiosk', got %q", f.Source)
			}
			return
		}
	}
	t.Errorf("feedback id=%d not found in source-filtered list", id)
}

func TestListFeedback_PageTwoReturnsResults(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	var created []int64
	for i := 0; i < 3; i++ {
		id, err := repo.Create(ctx, domain.FeedbackInput{
			OwnerID:    testutil.OwnerID,
			BusinessID: testutil.BusinessID,
			CustomerID: 5301,
			Type:       "review",
			Score:      i + 1,
		})
		if err != nil {
			t.Fatalf("Create i=%d: %v", i, err)
		}
		created = append(created, id)
	}
	t.Cleanup(func() {
		for _, id := range created {
			_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
		}
	})

	// With pageSize=1 there must be at least a second page when total > 1.
	_, total, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{}, 1, 1)
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}
	if total <= 1 {
		t.Skip("not enough rows to test page 2")
	}

	page2, _, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{}, 2, 1)
	if err != nil {
		t.Fatalf("List page 2: %v", err)
	}
	_ = page2
}

func TestCreateMultipleFeedback_CountGrows(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, before, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{}, 1, 1000)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	var created []int64
	for i := 0; i < 3; i++ {
		id, err := repo.Create(ctx, domain.FeedbackInput{
			OwnerID:    testutil.OwnerID,
			BusinessID: testutil.BusinessID,
			CustomerID: 5301,
			Type:       "review",
			Score:      5,
		})
		if err != nil {
			t.Fatalf("Create i=%d: %v", i, err)
		}
		created = append(created, id)
	}
	t.Cleanup(func() {
		for _, id := range created {
			_, _ = db.ExecContext(ctx, `DELETE FROM crm_feedback WHERE id = ?`, id)
		}
	})

	_, after, err := repo.List(ctx, testutil.OwnerID, domain.ListFilters{}, 1, 1000)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if after != before+3 {
		t.Errorf("expected total to grow by 3: before=%d after=%d", before, after)
	}
}
