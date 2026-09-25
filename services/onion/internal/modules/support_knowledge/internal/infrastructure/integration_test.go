package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/support_knowledge/internal/domain"
	"github.com/voxire/onion/internal/modules/support_knowledge/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const testKBBusinessID int64 = 1

func TestSupportKnowledge_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, _, err := repo.List(context.Background(), testKBBusinessID, 1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestSupportKnowledge_List_ReturnsTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, total, err := repo.List(context.Background(), testKBBusinessID, 1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < 0 {
		t.Errorf("expected total >= 0, got %d", total)
	}
}

func TestSupportKnowledge_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ArticleInput{
		BusinessID:  testKBBusinessID,
		Title:       "Integration Test Article",
		Slug:        "integration-test-article",
		Body:        "This is the body of the article.",
		Category:    "general",
		IsPublished: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testKBBusinessID, id) })
}

func TestSupportKnowledge_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ArticleInput{
		BusinessID: testKBBusinessID,
		Title:      "ListTestArticle",
		Slug:       "list-test-article",
		Body:       "body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testKBBusinessID, id) })

	items, _, err := repo.List(context.Background(), testKBBusinessID, 1, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, a := range items {
		if a.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created article id=%d not found in list", id)
	}
}

func TestSupportKnowledge_GetByTitleIsCaseInsensitive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ArticleInput{
		BusinessID: testKBBusinessID,
		Title:      "Case Sensitive Title",
		Slug:       "case-sensitive-title",
		Body:       "body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testKBBusinessID, id) })

	article, err := repo.Get(context.Background(), testKBBusinessID, 0, "", "cAsE sEnSiTiVe TiTlE")
	if err != nil {
		t.Fatalf("Get by mixed-case title: %v", err)
	}
	if article.ID != id {
		t.Fatalf("article id = %d, want %d", article.ID, id)
	}
}

func TestSupportKnowledge_Update_ChangesTitle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ArticleInput{
		BusinessID: testKBBusinessID, Title: "OldTitle", Slug: "old-title", Body: "body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testKBBusinessID, id) })

	if err := repo.Update(context.Background(), testKBBusinessID, id, domain.ArticleInput{
		BusinessID: testKBBusinessID, Title: "NewTitle", Slug: "new-title", Body: "body",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _, _ := repo.List(context.Background(), testKBBusinessID, 1, 100)
	for _, a := range items {
		if a.ID == id && a.Title != "NewTitle" {
			t.Errorf("expected title NewTitle, got %s", a.Title)
		}
	}
}

func TestSupportKnowledge_Delete_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ArticleInput{
		BusinessID: testKBBusinessID, Title: "DeleteMe", Slug: "delete-me", Body: "body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), testKBBusinessID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, _, _ := repo.List(context.Background(), testKBBusinessID, 1, 100)
	for _, a := range items {
		if a.ID == id {
			t.Errorf("deleted article id=%d still in list", id)
		}
	}
}

func TestSupportKnowledge_List_Pagination(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, total1, err := repo.List(context.Background(), testKBBusinessID, 1, 5)
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}
	_, total2, err := repo.List(context.Background(), testKBBusinessID, 2, 5)
	if err != nil {
		t.Fatalf("List page 2: %v", err)
	}
	if total1 != total2 {
		t.Errorf("total inconsistent across pages: %d vs %d", total1, total2)
	}
}

func TestSupportKnowledge_List_UnknownBusiness_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, total, err := repo.List(context.Background(), 9999999, 1, 20)
	if err != nil {
		t.Fatalf("List unknown business: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total=0 for unknown business, got %d", total)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestSupportKnowledge_Update_IsPublished(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ArticleInput{
		BusinessID: testKBBusinessID, Title: "DraftArticle", Slug: "draft-article", Body: "body", IsPublished: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testKBBusinessID, id) })

	if err := repo.Update(context.Background(), testKBBusinessID, id, domain.ArticleInput{
		BusinessID: testKBBusinessID, Title: "DraftArticle", Slug: "draft-article", Body: "body", IsPublished: true,
	}); err != nil {
		t.Fatalf("Update publish: %v", err)
	}
}

func TestSupportKnowledge_Create_WithCategory(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ArticleInput{
		BusinessID: testKBBusinessID,
		Title:      "Billing FAQ",
		Slug:       "billing-faq",
		Body:       "How to manage billing.",
		Category:   "billing",
	})
	if err != nil {
		t.Fatalf("Create with category: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testKBBusinessID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestSupportKnowledge_Create_FieldsPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ArticleInput{
		BusinessID:  testKBBusinessID,
		Title:       "Field Persistence Test",
		Slug:        "field-persistence-test",
		Body:        "Detailed body content",
		Category:    "technical",
		IsPublished: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testKBBusinessID, id) })

	items, _, err := repo.List(context.Background(), testKBBusinessID, 1, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID == id {
			if a.Title != "Field Persistence Test" {
				t.Errorf("expected title 'Field Persistence Test', got %s", a.Title)
			}
			if a.Slug != "field-persistence-test" {
				t.Errorf("expected slug 'field-persistence-test', got %s", a.Slug)
			}
			if a.Body != "Detailed body content" {
				t.Errorf("expected body 'Detailed body content', got %s", a.Body)
			}
			if a.Category != "technical" {
				t.Errorf("expected category 'technical', got %s", a.Category)
			}
			return
		}
	}
	t.Errorf("created article id=%d not found in list", id)
}

func TestSupportKnowledge_List_PageSize1_ReturnsAtMostOne(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, _ := repo.Create(context.Background(), domain.ArticleInput{
		BusinessID: testKBBusinessID, Title: "Page1Test", Slug: "page1-test", Body: "body",
	})
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testKBBusinessID, id) })

	items, _, err := repo.List(context.Background(), testKBBusinessID, 1, 1)
	if err != nil {
		t.Fatalf("List pageSize=1: %v", err)
	}
	if len(items) > 1 {
		t.Errorf("expected at most 1 item, got %d", len(items))
	}
}

func TestSupportKnowledge_Update_Body(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ArticleInput{
		BusinessID: testKBBusinessID, Title: "BodyUpdate", Slug: "body-update", Body: "old body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testKBBusinessID, id) })

	if err := repo.Update(context.Background(), testKBBusinessID, id, domain.ArticleInput{
		BusinessID: testKBBusinessID, Title: "BodyUpdate", Slug: "body-update", Body: "new body",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _, _ := repo.List(context.Background(), testKBBusinessID, 1, 100)
	for _, a := range items {
		if a.ID == id && a.Body != "new body" {
			t.Errorf("expected body 'new body', got %s", a.Body)
		}
	}
}

func TestSupportKnowledge_Update_Category(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ArticleInput{
		BusinessID: testKBBusinessID, Title: "CatUpdate", Slug: "cat-update", Body: "body", Category: "general",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testKBBusinessID, id) })

	if err := repo.Update(context.Background(), testKBBusinessID, id, domain.ArticleInput{
		BusinessID: testKBBusinessID, Title: "CatUpdate", Slug: "cat-update", Body: "body", Category: "advanced",
	}); err != nil {
		t.Fatalf("Update category: %v", err)
	}

	items, _, _ := repo.List(context.Background(), testKBBusinessID, 1, 100)
	for _, a := range items {
		if a.ID == id && a.Category != "advanced" {
			t.Errorf("expected category 'advanced', got %s", a.Category)
		}
	}
}

func TestSupportKnowledge_List_TotalNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, total, err := repo.List(context.Background(), testKBBusinessID, 1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < 0 {
		t.Errorf("expected total >= 0, got %d", total)
	}
}

func TestSupportKnowledge_Delete_WrongBusiness_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	err := repo.Delete(context.Background(), 9999999, 9999999)
	if err != nil {
		t.Fatalf("Delete non-existent: %v", err)
	}
}

func TestSupportKnowledge_Create_Multiple_AllAppearInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	slugs := []string{"multi-kb-a", "multi-kb-b", "multi-kb-c"}
	ids := make([]int64, 0, len(slugs))
	for i, slug := range slugs {
		id, err := repo.Create(context.Background(), domain.ArticleInput{
			BusinessID: testKBBusinessID,
			Title:      "MultiKB " + slug,
			Slug:       slug,
			Body:       "body",
		})
		if err != nil {
			t.Fatalf("Create [%d]: %v", i, err)
		}
		ids = append(ids, id)
	}
	t.Cleanup(func() {
		for _, id := range ids {
			_ = repo.Delete(context.Background(), testKBBusinessID, id)
		}
	})

	items, _, err := repo.List(context.Background(), testKBBusinessID, 1, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := make(map[int64]bool)
	for _, a := range items {
		found[a.ID] = true
	}
	for _, id := range ids {
		if !found[id] {
			t.Errorf("article id=%d not found in list", id)
		}
	}
}

func TestSupportKnowledge_List_EachHasPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ArticleInput{
		BusinessID: testKBBusinessID, Title: "IDCheck", Slug: "id-check-kb", Body: "body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testKBBusinessID, id) })

	items, _, err := repo.List(context.Background(), testKBBusinessID, 1, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.ID <= 0 {
			t.Errorf("article has non-positive ID: %d", a.ID)
		}
	}
}

func TestSupportKnowledge_List_EachHasNonEmptyTitle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ArticleInput{
		BusinessID: testKBBusinessID, Title: "TitleCheck", Slug: "title-check-kb", Body: "body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testKBBusinessID, id) })

	items, _, err := repo.List(context.Background(), testKBBusinessID, 1, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.Title == "" {
			t.Errorf("article id=%d has empty title", a.ID)
		}
	}
}

func TestSupportKnowledge_List_BusinessScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ArticleInput{
		BusinessID: testKBBusinessID, Title: "ScopeCheck", Slug: "scope-check-kb", Body: "body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testKBBusinessID, id) })

	items, _, err := repo.List(context.Background(), testKBBusinessID, 1, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, a := range items {
		if a.BusinessID != testKBBusinessID {
			t.Errorf("expected business_id %d, got %d", testKBBusinessID, a.BusinessID)
		}
	}
}

func TestSupportKnowledge_Update_WrongBusiness_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	err := repo.Update(context.Background(), 9999999, 9999999, domain.ArticleInput{
		BusinessID: 9999999, Title: "ghost", Slug: "ghost", Body: "ghost",
	})
	if err != nil {
		t.Fatalf("Update on missing row: %v", err)
	}
}
