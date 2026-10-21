package application

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// wikipediaHost is formatted with a language code, e.g. "en".
var (
	wikipediaHost     = "https://%s.wikipedia.org"
	wikipediaLanguage = regexp.MustCompile(`^[a-z]{2,3}$`)
)

func (e *AssistantToolExecutor) wikipedia(ctx context.Context, input map[string]any) (map[string]any, error) {
	query := inputText(input, "query", "title", "topic", "q")
	if query == "" || len([]rune(query)) > 200 {
		return nil, errors.New("give a topic to look up")
	}
	language := strings.ToLower(inputText(input, "language", "lang"))
	if !wikipediaLanguage.MatchString(language) {
		language = "en"
	}
	base := strings.Replace(wikipediaHost, "%s", language, 1)
	var search struct {
		Pages []struct {
			Key         string `json:"key"`
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"pages"`
	}
	params := url.Values{"q": {query}, "limit": {"5"}}
	if err := getPublicJSON(ctx, e.httpClient(), base+"/w/rest.php/v1/search/page?"+params.Encode(), &search); err != nil {
		return nil, err
	}
	if len(search.Pages) == 0 {
		return nil, errors.New("no Wikipedia article matches that topic")
	}
	var summary struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Extract     string `json:"extract"`
		Type        string `json:"type"`
		ContentURLs struct {
			Desktop struct {
				Page string `json:"page"`
			} `json:"desktop"`
		} `json:"content_urls"`
	}
	top := search.Pages[0]
	if err := getPublicJSON(ctx, e.httpClient(), base+"/api/rest_v1/page/summary/"+url.PathEscape(top.Key), &summary); err != nil {
		return nil, err
	}
	related := make([]map[string]string, 0, len(search.Pages)-1)
	for _, page := range search.Pages[1:] {
		related = append(related, map[string]string{"title": page.Title, "description": page.Description})
	}
	out := map[string]any{
		"title":         summary.Title,
		"description":   summary.Description,
		"summary":       truncateSearchText(summary.Extract, 3000),
		"url":           summary.ContentURLs.Desktop.Page,
		"other_matches": related,
	}
	if summary.Type == "disambiguation" {
		out["note"] = "This title is ambiguous; one of the other matches may be the intended article."
	}
	if strings.HasPrefix(summary.ContentURLs.Desktop.Page, "https://") {
		out["sources"] = []map[string]string{{"title": summary.Title + " - Wikipedia", "url": summary.ContentURLs.Desktop.Page}}
	}
	return out, nil
}
