package handlers

import (
	"context"
	"testing"
)

func TestCreateArticleHandler_Validate_RequiresTitle(t *testing.T) {
	h := &CreateArticleHandler{}
	req := &CreateArticleRequest{Title: "", Slug: "my-slug"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when title is empty")
	}
}

func TestCreateArticleHandler_Validate_RequiresSlug(t *testing.T) {
	h := &CreateArticleHandler{}
	req := &CreateArticleRequest{Title: "My Article", Slug: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when slug is empty")
	}
}

func TestCreateArticleHandler_Validate_HappyPath(t *testing.T) {
	h := &CreateArticleHandler{}
	req := &CreateArticleRequest{Title: "My Article", Slug: "my-article"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateArticleHandler_Validate_WhitespaceTitle(t *testing.T) {
	h := &CreateArticleHandler{}
	req := &CreateArticleRequest{Title: "   ", Slug: "my-article"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only title")
	}
}

func TestCreateArticleHandler_Validate_WhitespaceSlug(t *testing.T) {
	h := &CreateArticleHandler{}
	req := &CreateArticleRequest{Title: "My Article", Slug: "\t\n "}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only slug")
	}
}

func TestCreateArticleHandler_Validate_BothEmpty(t *testing.T) {
	h := &CreateArticleHandler{}
	req := &CreateArticleRequest{}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when both fields are empty")
	}
}

func TestCreateArticleHandler_Validate_WithOptionalFields(t *testing.T) {
	h := &CreateArticleHandler{}
	req := &CreateArticleRequest{
		Title:       "FAQ",
		Slug:        "faq",
		Body:        "Some content here.",
		Category:    "general",
		IsPublished: true,
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error with optional fields: %v", err)
	}
}

func TestUpdateArticleHandler_Validate_RequiresID(t *testing.T) {
	h := &UpdateArticleHandler{}
	req := &UpdateArticleRequest{ID: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestUpdateArticleHandler_Validate_HappyPath(t *testing.T) {
	h := &UpdateArticleHandler{}
	req := &UpdateArticleRequest{ID: 3, Title: "Updated"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateArticleHandler_Validate_LargeIDPasses(t *testing.T) {
	h := &UpdateArticleHandler{}
	req := &UpdateArticleRequest{ID: 99999}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for large id: %v", err)
	}
}

func TestDeleteArticleHandler_Validate_RequiresID(t *testing.T) {
	h := &DeleteArticleHandler{}
	req := &DeleteArticleRequest{ID: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestDeleteArticleHandler_Validate_HappyPath(t *testing.T) {
	h := &DeleteArticleHandler{}
	req := &DeleteArticleRequest{ID: 5}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteArticleHandler_Validate_LargeIDPasses(t *testing.T) {
	h := &DeleteArticleHandler{}
	req := &DeleteArticleRequest{ID: 100000}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for large id: %v", err)
	}
}
