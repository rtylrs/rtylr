package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

// errWebSearchNotConfigured marks a backend that has no credentials, so the
// fallback chain can skip it without hiding the primary's real error.
var errWebSearchNotConfigured = errors.New("web search is not configured")

// FallbackWebSearch asks the self-hosted SearXNG first and, when it fails or
// finds nothing, a paid search API. Free engines often block a server IP, so
// the fallback keeps Caret from answering "no results".
type FallbackWebSearch struct {
	Primary  PublicWebSearcher
	Fallback PublicWebSearcher
}

func (s FallbackWebSearch) Search(ctx context.Context, query string) (map[string]any, error) {
	result, err := s.Primary.Search(ctx, query)
	if err == nil || s.Fallback == nil {
		return result, err
	}
	fallback, fallbackErr := s.Fallback.Search(ctx, query)
	if errors.Is(fallbackErr, errWebSearchNotConfigured) {
		return nil, err
	}
	if fallbackErr != nil {
		logger.Ctx(ctx).Warn("fallback web search failed", logger.Error(fallbackErr), logger.String("primary_error", err.Error()))
		return nil, err
	}
	logger.Ctx(ctx).Info("web search answered by fallback", logger.String("primary_error", err.Error()))
	return fallback, nil
}

const braveSearchEndpoint = "https://api.search.brave.com/res/v1/web/search"

// BraveWebSearch queries the Brave Search API. It is inert until
// BRAVE_SEARCH_API_KEY is set.
type BraveWebSearch struct {
	Client   *http.Client
	APIKey   string
	Endpoint string
}

func (s BraveWebSearch) Search(ctx context.Context, query string) (map[string]any, error) {
	apiKey := strings.TrimSpace(s.APIKey)
	if apiKey == "" {
		apiKey = config.GetBraveSearchAPIKey()
	}
	if apiKey == "" {
		return nil, errWebSearchNotConfigured
	}
	endpoint := s.Endpoint
	if endpoint == "" {
		endpoint = braveSearchEndpoint
	}
	params := url.Values{}
	params.Set("q", query)
	params.Set("count", "10")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, errors.New("web search request failed")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", apiKey)
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("web search request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("brave search returned status %d", resp.StatusCode)
	}
	payload, err := io.ReadAll(io.LimitReader(resp.Body, (1<<20)+1))
	if err != nil || len(payload) > 1<<20 {
		return nil, errors.New("web search response exceeded the size limit")
	}
	return parseBraveResults(payload, query)
}

var htmlTag = regexp.MustCompile(`<[^>]*>`)

func parseBraveResults(payload []byte, query string) (map[string]any, error) {
	var response struct {
		Web struct {
			Results []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
			} `json:"results"`
		} `json:"web"`
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		return nil, errors.New("web search returned an invalid response")
	}
	results := make([]webSearchResult, 0, len(response.Web.Results))
	for _, result := range response.Web.Results {
		results = append(results, webSearchResult{
			Title:   plainSearchText(result.Title),
			URL:     result.URL,
			Snippet: plainSearchText(result.Description),
		})
	}
	return searchSourcesOutput(results, query)
}

// plainSearchText drops the <strong> highlighting Brave adds to matches.
func plainSearchText(value string) string {
	return html.UnescapeString(htmlTag.ReplaceAllString(value, ""))
}
