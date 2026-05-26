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
