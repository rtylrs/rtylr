package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/support_knowledge/internal/application"
	"github.com/voxire/onion/internal/modules/support_knowledge/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type ArticlePayload struct {
	ID          int64  `json:"id"`
	BusinessID  int64  `json:"business_id"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Body        string `json:"body"`
	Category    string `json:"category"`
	IsPublished bool   `json:"is_published"`
	Views       int    `json:"views"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toArticlePayload(a domain.Article) ArticlePayload {
	return ArticlePayload{
		ID: a.ID, BusinessID: a.BusinessID, Title: a.Title, Slug: a.Slug,
		Body: a.Body, Category: a.Category, IsPublished: a.IsPublished, Views: a.Views,
		CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: a.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// -- LIST --------------------------------------------------------------------

func ListArticlesRoutes() []string  { return []string{"/support/knowledge"} }
func CreateArticleRoutes() []string { return []string{"/support/knowledge"} }
func UpdateArticleRoutes() []string { return []string{"/support/knowledge"} }
func DeleteArticleRoutes() []string { return []string{"/support/knowledge"} }

type ListArticlesHandler struct{ Service *application.Service }

type ListArticlesRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type ListArticlesResponse struct {
	Articles []ArticlePayload `json:"articles"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

func (h *ListArticlesHandler) Serve(ctx context.Context, req *ListArticlesRequest) (ListArticlesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListArticlesResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	items, total, err := h.Service.List(ctx, device.User.UserID, req.Page, req.PageSize)
	if err != nil {
		return ListArticlesResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list articles"}
	}
	payloads := make([]ArticlePayload, len(items))
	for i, a := range items {
		payloads[i] = toArticlePayload(a)
	}
	return ListArticlesResponse{Articles: payloads, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// -- CREATE ------------------------------------------------------------------

type CreateArticleHandler struct{ Service *application.Service }

type CreateArticleRequest struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Body        string `json:"body"`
	Category    string `json:"category"`
	IsPublished bool   `json:"is_published"`
}

type CreateArticleResponse struct {
	ArticleID int64  `json:"article_id"`
	Message   string `json:"message"`
}

func (h *CreateArticleHandler) Validate(_ context.Context, req *CreateArticleRequest) *errorModels.ServerError {
	if req.Title == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Title is required"}
	}
	if req.Slug == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Slug is required"}
	}
	return nil
}

func (h *CreateArticleHandler) Serve(ctx context.Context, req *CreateArticleRequest) (CreateArticleResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateArticleResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, domain.ArticleInput{
		BusinessID:  device.User.UserID,
		Title:       req.Title,
		Slug:        req.Slug,
		Body:        req.Body,
		Category:    req.Category,
		IsPublished: req.IsPublished,
	})
	if err != nil {
		return CreateArticleResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create article"}
	}
	return CreateArticleResponse{ArticleID: id, Message: "Article created"}, nil
}

// -- UPDATE ------------------------------------------------------------------

type UpdateArticleHandler struct{ Service *application.Service }

type UpdateArticleRequest struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Body        string `json:"body"`
	Category    string `json:"category"`
	IsPublished bool   `json:"is_published"`
}

type UpdateArticleResponse struct {
	Message string `json:"message"`
}

func (h *UpdateArticleHandler) Validate(_ context.Context, req *UpdateArticleRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "ID is required"}
	}
	return nil
}

func (h *UpdateArticleHandler) Serve(ctx context.Context, req *UpdateArticleRequest) (UpdateArticleResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateArticleResponse{}, errorModels.PermissionDeniedServerError()
	}
	err := h.Service.Update(ctx, device.User.UserID, req.ID, domain.ArticleInput{
		Title:       req.Title,
		Slug:        req.Slug,
		Body:        req.Body,
		Category:    req.Category,
		IsPublished: req.IsPublished,
	})
	if err != nil {
		return UpdateArticleResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update article"}
	}
	return UpdateArticleResponse{Message: "Article updated"}, nil
}

// -- DELETE ------------------------------------------------------------------

type DeleteArticleHandler struct{ Service *application.Service }

type DeleteArticleRequest struct {
	ID int64 `json:"id"`
}

type DeleteArticleResponse struct {
	Message string `json:"message"`
}

func (h *DeleteArticleHandler) Validate(_ context.Context, req *DeleteArticleRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "ID is required"}
	}
	return nil
}

func (h *DeleteArticleHandler) Serve(ctx context.Context, req *DeleteArticleRequest) (DeleteArticleResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteArticleResponse{}, errorModels.PermissionDeniedServerError()
	}
	err := h.Service.Delete(ctx, device.User.UserID, req.ID)
	if err != nil {
		return DeleteArticleResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete article"}
	}
	return DeleteArticleResponse{Message: "Article deleted"}, nil
}
