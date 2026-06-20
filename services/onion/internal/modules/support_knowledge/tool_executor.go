package supportknowledge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/support_knowledge/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

const (
	listSupportArticlesToolName         = "support.list_articles"
	getSupportArticleToolName           = "support.get_article"
	defaultListSupportArticlesToolLimit = int64(20)
	maxListSupportArticlesToolLimit     = int64(50)
	supportArticleToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type getSupportArticleToolInput struct {
	ArticleID int64
	Slug      string
	Title     string
}

type listSupportArticlesToolInput struct {
	Page  int64
	Limit int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listSupportArticlesToolName:
		output, err := m.executeListSupportArticlesTool(ctx, input)
		return output, true, err
	case getSupportArticleToolName:
		output, err := m.executeGetSupportArticleTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeGetSupportArticleTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := supportArticleToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := parseGetSupportArticleToolInput(input)
	if err != nil {
		return nil, err
	}

	article, err := m.service.Get(ctx, businessID, parsed.ArticleID, parsed.Slug, parsed.Title)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"article": supportArticleToolPayload(article),
	}, nil
}

func (m *Module) executeListSupportArticlesTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	businessID, err := supportArticleToolBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	parsed, err := parseListSupportArticlesToolInput(input)
	if err != nil {
		return nil, err
	}

	articles, total, err := m.service.List(ctx, businessID, int(parsed.Page), int(parsed.Limit))
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(articles))
	for _, article := range articles {
		payloads = append(payloads, supportArticleToolPayload(article))
	}

	return map[string]any{
		"articles":  payloads,
		"total":     total,
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  parsed.Page*parsed.Limit < int64(total),
	}, nil
}

func parseGetSupportArticleToolInput(input map[string]any) (getSupportArticleToolInput, error) {
	articleID, err := supportArticleIDToolValue(firstSupportArticleToolValue(input, "article_id", "support_article_id", "knowledge_article_id", "article", "id"))
	if err != nil {
		return getSupportArticleToolInput{}, fmt.Errorf("invalid article_id: %w", err)
	}
	slug := supportArticleToolSlug(firstSupportArticleToolValue(input, "slug", "handle"))
	title := strings.TrimSpace(supportArticleToolString(firstSupportArticleToolValue(input, "title", "name")))

	if articleID <= 0 && slug == "" && title == "" {
		return getSupportArticleToolInput{}, errors.New("article_id, slug, or title is required")
	}
	return getSupportArticleToolInput{ArticleID: articleID, Slug: slug, Title: title}, nil
}

func parseListSupportArticlesToolInput(input map[string]any) (listSupportArticlesToolInput, error) {
	page, err := supportArticleToolNonNegativeInt64(firstSupportArticleToolValue(input, "page"))
	if err != nil {
		return listSupportArticlesToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page == 0 {
		page = 1
	}

	limit, err := supportArticleToolNonNegativeInt64(firstSupportArticleToolValue(input, "limit", "page_size"))
	if err != nil {
		return listSupportArticlesToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit == 0 {
		limit = defaultListSupportArticlesToolLimit
	}
	if limit > maxListSupportArticlesToolLimit {
		limit = maxListSupportArticlesToolLimit
	}

	return listSupportArticlesToolInput{Page: page, Limit: limit}, nil
}

func supportArticleToolBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User != nil && device.User.BusinessID > 0 {
		return device.User.BusinessID, nil
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if ok && authUser != nil && authUser.BusinessID > 0 {
		return authUser.BusinessID, nil
	}
	return 0, errors.New("authentication is required")
}

func supportArticleToolPayload(article domain.Article) map[string]any {
	return map[string]any{
		"id":           article.ID,
		"business_id":  article.BusinessID,
		"title":        article.Title,
		"slug":         article.Slug,
		"body":         article.Body,
		"category":     article.Category,
		"is_published": article.IsPublished,
		"views":        article.Views,
		"created_at":   formatSupportArticleToolTime(article.CreatedAt),
		"updated_at":   formatSupportArticleToolTime(article.UpdatedAt),
	}
}

func firstSupportArticleToolValue(input map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := input[key]
		if !ok || value == nil {
			continue
		}
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		return value
	}
	return nil
}

func supportArticleToolNonNegativeInt64(value any) (int64, error) {
	out, err := supportArticleToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func supportArticleIDToolValue(value any) (int64, error) {
	switch v := value.(type) {
	case string:
		raw := strings.TrimSpace(v)
		raw = strings.TrimPrefix(raw, "#")
		raw = strings.ToLower(raw)
		for _, prefix := range []string{"support-article-", "support_article_", "knowledge-article-", "knowledge_article_", "article-", "article_", "kb-", "kb_"} {
			raw = strings.TrimPrefix(raw, prefix)
		}
		return supportArticleToolInt64(raw)
	default:
		return supportArticleToolInt64(value)
	}
}

func supportArticleToolInt64(value any) (int64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		if math.Trunc(v) != v {
			return 0, errors.New("must be an integer")
		}
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		return strconv.ParseInt(raw, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func supportArticleToolString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func supportArticleToolSlug(value any) string {
	return strings.ToLower(strings.TrimSpace(supportArticleToolString(value)))
}

func formatSupportArticleToolTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UTC().Format(supportArticleToolTimestampDate)
}
