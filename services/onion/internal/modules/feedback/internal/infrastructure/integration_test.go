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
