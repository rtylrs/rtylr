package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/support_knowledge/internal/application"
	"github.com/voxire/onion/internal/modules/support_knowledge/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	supportknowledgev1 "github.com/voxire/voxie/gen/go/support/knowledge/v1"
)

type KnowledgeAPI struct {
	service *application.Service
}

func NewKnowledgeAPI(service *application.Service) *KnowledgeAPI {
	return &KnowledgeAPI{service: service}
}

func (a *KnowledgeAPI) ListArticles(ctx context.Context, req *supportknowledgev1.ListArticlesRequest) (*supportknowledgev1.ListArticlesResponse, error) {
	businessID, errModel := businessIDFromContext(ctx)
	if errModel != nil {
		return nil, errModel
	}

	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	items, total, err := a.service.List(ctx, businessID, page, pageSize)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list articles")
	}

	return &supportknowledgev1.ListArticlesResponse{
		Articles: newArticles(items),
		Total:    int32(total), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

func (a *KnowledgeAPI) CreateArticle(ctx context.Context, req *supportknowledgev1.CreateArticleRequest) (*supportknowledgev1.CreateArticleResponse, error) {
	if strings.TrimSpace(req.GetTitle()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Title is required")
	}
	if strings.TrimSpace(req.GetSlug()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Slug is required")
	}

	businessID, errModel := businessIDFromContext(ctx)
	if errModel != nil {
		return nil, errModel
	}

	id, err := a.service.Create(ctx, domain.ArticleInput{
		BusinessID:  businessID,
		Title:       req.GetTitle(),
		Slug:        req.GetSlug(),
		Body:        req.GetBody(),
		Category:    req.GetCategory(),
		IsPublished: req.GetIsPublished(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create article")
	}

	return &supportknowledgev1.CreateArticleResponse{ArticleId: id, Message: "Article created"}, nil
}

func (a *KnowledgeAPI) UpdateArticle(ctx context.Context, req *supportknowledgev1.UpdateArticleRequest) (*supportknowledgev1.UpdateArticleResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}

	businessID, errModel := businessIDFromContext(ctx)
	if errModel != nil {
		return nil, errModel
	}

	if err := a.service.Update(ctx, businessID, req.GetId(), domain.ArticleInput{
		Title:       req.GetTitle(),
		Slug:        req.GetSlug(),
		Body:        req.GetBody(),
		Category:    req.GetCategory(),
		IsPublished: req.GetIsPublished(),
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update article")
	}

	return &supportknowledgev1.UpdateArticleResponse{Message: "Article updated"}, nil
}

func (a *KnowledgeAPI) DeleteArticle(ctx context.Context, req *supportknowledgev1.DeleteArticleRequest) (*supportknowledgev1.DeleteArticleResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}

	businessID, errModel := businessIDFromContext(ctx)
	if errModel != nil {
		return nil, errModel
	}

	if err := a.service.Delete(ctx, businessID, req.GetId()); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete article")
	}

	return &supportknowledgev1.DeleteArticleResponse{Message: "Article deleted"}, nil
}

func businessIDFromContext(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return device.User.BusinessID, nil
}

func newArticles(items []domain.Article) []*supportknowledgev1.Article {
	out := make([]*supportknowledgev1.Article, 0, len(items))
	for _, item := range items {
		out = append(out, &supportknowledgev1.Article{
			Id:          item.ID,
			BusinessId:  item.BusinessID,
			Title:       item.Title,
			Slug:        item.Slug,
			Body:        item.Body,
			Category:    item.Category,
			IsPublished: item.IsPublished,
			Views:       int32(item.Views), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			CreatedAt:   formatTime(item.CreatedAt),
			UpdatedAt:   formatTime(item.UpdatedAt),
		})
	}
	return out
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
