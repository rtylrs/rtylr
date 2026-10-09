package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

type AssistantProfileReader interface {
	GetAssistantProfile(context.Context) (string, error)
}

type PublicWebSearcher interface {
	Search(context.Context, string) (map[string]any, error)
}

type AssistantToolExecutor struct {
	Profile AssistantProfileReader
	Web     PublicWebSearcher
}

type webSearchBudgetKey struct{}
type webSearchBudget struct{ used atomic.Bool }

func WithWebSearchBudget(ctx context.Context) context.Context {
	return context.WithValue(ctx, webSearchBudgetKey{}, &webSearchBudget{})
}

func NewAssistantToolExecutor(profile AssistantProfileReader, web PublicWebSearcher) *AssistantToolExecutor {
	if web == nil {
		web = SearXNGWebSearch{Client: &http.Client{Timeout: 12 * time.Second}}
	}
	return &AssistantToolExecutor{Profile: profile, Web: web}
}

func (e *AssistantToolExecutor) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case "account.profile":
		if e.Profile == nil {
			return nil, true, errors.New("profile service is unavailable")
		}
		profile, err := e.Profile.GetAssistantProfile(ctx)
		return map[string]any{"profile": profile}, true, err
	case "utilities.calculate":
		expression, _ := input["expression"].(string)
		value, err := calculateArithmetic(expression)
		return map[string]any{"expression": expression, "result": value}, true, err
	case "utilities.time":
		zone, _ := input["timezone"].(string)
		if len(zone) > 128 {
			return nil, true, errors.New("timezone is too long")
		}
		location, err := time.LoadLocation(zone)
		if err != nil {
			return nil, true, errors.New("use a valid IANA timezone")
		}
		now := time.Now().In(location)
		return map[string]any{"timezone": zone, "local_time": now.Format(time.RFC3339), "date": now.Format("Monday, January 2, 2006")}, true, nil
	case "web.search":
		query, _ := input["query"].(string)
		query = strings.TrimSpace(query)
		if query == "" || len(query) > 500 || containsPersonalContact(query) {
			return nil, true, errors.New("search needs a short public query without email addresses or phone numbers")
		}
		if e.Web == nil {
			return nil, true, errors.New("web search is unavailable")
		}
		if budget, ok := ctx.Value(webSearchBudgetKey{}).(*webSearchBudget); ok && !budget.used.CompareAndSwap(false, true) {
			return nil, true, errors.New("one public web search is available per message; use its results or ask another question")
		}
		result, err := e.Web.Search(ctx, query)
		return result, true, err
	default:
		return nil, false, nil
	}
}

var (
	emailAddress    = regexp.MustCompile(`(?i)[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}`)
	phoneNumberLike = regexp.MustCompile(`\+?[0-9][0-9 .()/-]{7,}[0-9]`)
)

func containsPersonalContact(query string) bool {
	if emailAddress.MatchString(query) {
		return true
	}
	for _, candidate := range phoneNumberLike.FindAllString(query, -1) {
		digits := 0
		for _, r := range candidate {
			if r >= '0' && r <= '9' {
				digits++
			}
		}
		if digits >= 10 || (strings.HasPrefix(strings.TrimSpace(candidate), "+") && digits >= 8) {
			return true
		}
	}
	return false
}

func calculateArithmetic(expression string) (float64, error) {
	if len(expression) == 0 || len(expression) > 512 {
		return 0, errors.New("enter an arithmetic expression under 512 characters")
	}
	node, err := parser.ParseExpr(expression)
	if err != nil {
		return 0, errors.New("invalid arithmetic expression")
	}
	value, err := evalArithmetic(node)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		if err != nil {
			return 0, err
		}
		return 0, errors.New("calculation is outside the supported numeric range")
	}
	return value, nil
}

func evalArithmetic(node ast.Expr) (float64, error) {
	switch node := node.(type) {
	case *ast.BasicLit:
		if node.Kind != token.INT && node.Kind != token.FLOAT {
			return 0, errors.New("only decimal numbers are supported")
		}
		value, err := strconv.ParseFloat(node.Value, 64)
		if err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
			return 0, errors.New("invalid number")
		}
		return value, nil
	case *ast.ParenExpr:
		return evalArithmetic(node.X)
	case *ast.UnaryExpr:
		value, err := evalArithmetic(node.X)
		if err != nil {
			return 0, err
		}
		switch node.Op {
		case token.ADD:
			return value, nil
		case token.SUB:
			return -value, nil
		default:
			return 0, errors.New("unsupported arithmetic operator")
		}
	case *ast.BinaryExpr:
		left, err := evalArithmetic(node.X)
		if err != nil {
			return 0, err
		}
		right, err := evalArithmetic(node.Y)
		if err != nil {
			return 0, err
		}
		switch node.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			if right == 0 {
				return 0, errors.New("cannot divide by zero")
			}
			return left / right, nil
		default:
			return 0, errors.New("unsupported arithmetic operator")
		}
	default:
		return 0, fmt.Errorf("unsupported arithmetic expression %T", node)
	}
}

type SearXNGWebSearch struct {
	Client  *http.Client
	BaseURL string
}

func (s SearXNGWebSearch) Search(ctx context.Context, query string) (map[string]any, error) {
	baseURL := strings.TrimSpace(s.BaseURL)
	if baseURL == "" {
		baseURL = config.GetSearXNGURL()
	}
	if baseURL == "" {
		return nil, errors.New("web search is not configured")
	}
	endpoint, err := url.Parse(baseURL)
	if err != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.Host == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return nil, errors.New("web search endpoint is invalid")
	}
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/search"
	params := endpoint.Query()
	params.Set("q", query)
	params.Set("format", "json")
	// News engines keep answering when a web engine is suspended.
	params.Set("categories", "general,news")
	endpoint.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, errors.New("web search request failed")
	}
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
		return nil, fmt.Errorf("web search service returned status %d", resp.StatusCode)
	}
	payload, err := io.ReadAll(io.LimitReader(resp.Body, (1<<20)+1))
	if err != nil || len(payload) > 1<<20 {
		return nil, errors.New("web search response exceeded the size limit")
	}
	result, err := parseSearXNGResults(payload, query)
	if err != nil {
		logger.Ctx(ctx).Warn("web search returned nothing",
			logger.Error(err), logger.String("unresponsive_engines", searXNGUnresponsiveEngines(payload)))
	}
	return result, err
}

// searXNGUnresponsiveEngines lists the engines SearXNG reported as failing,
// e.g. "bing: timeout, qwant news: CAPTCHA".
func searXNGUnresponsiveEngines(payload []byte) string {
	var response struct {
		Unresponsive [][]string `json:"unresponsive_engines"`
	}
	if json.Unmarshal(payload, &response) != nil {
		return ""
	}
	engines := make([]string, 0, len(response.Unresponsive))
	for _, engine := range response.Unresponsive {
		engines = append(engines, strings.Join(engine, ": "))
	}
	return strings.Join(engines, ", ")
}

func parseSearXNGResults(payload []byte, query string) (map[string]any, error) {
	var response struct {
		Results []struct {
			Title      string `json:"title"`
			URL        string `json:"url"`
			Content    string `json:"content"`
			Suggestion string `json:"suggestion"`
		} `json:"results"`
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		return nil, errors.New("web search returned an invalid response")
	}
	sources := make([]map[string]string, 0, 8)
	seen := make(map[string]bool)
	for _, result := range response.Results {
		if len(sources) == 8 {
			break
		}
		if len(result.URL) > 2048 {
			continue
		}
		parsedURL, err := url.ParseRequestURI(result.URL)
		if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" || parsedURL.User != nil || seen[parsedURL.String()] {
			continue
		}
		title := truncateSearchText(strings.TrimSpace(result.Title), 300)
		if title == "" {
			title = parsedURL.Host
		}
		snippet := strings.TrimSpace(result.Content)
		if snippet == "" {
			snippet = strings.TrimSpace(result.Suggestion)
		}
		sources = append(sources, map[string]string{
			"title":   title,
			"url":     parsedURL.String(),
			"snippet": truncateSearchText(snippet, 1200),
		})
		seen[parsedURL.String()] = true
	}
	if len(sources) == 0 {
		return nil, errors.New("web search returned no usable results")
	}
	return map[string]any{"sources": sources, "search_query": query}, nil
}

func truncateSearchText(value string, maxRunes int) string {
	runes := []rune(value)
	if len(runes) > maxRunes {
		return string(runes[:maxRunes])
	}
	return value
}
