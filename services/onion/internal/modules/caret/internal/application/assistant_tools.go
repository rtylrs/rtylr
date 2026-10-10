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
	// HTTP reads public APIs and pages. It only dials public addresses.
	HTTP *http.Client
}

// generalTool runs one tool that needs no business data: utilities, public
// reference APIs and the web.
type generalTool func(e *AssistantToolExecutor, ctx context.Context, input map[string]any) (map[string]any, error)

// generalTools maps each general tool to its handler; generalToolDefs
// describes them to the model.
var generalTools = map[string]generalTool{
	"web.fetch":                (*AssistantToolExecutor).fetchPage,
	"places.geocode":           (*AssistantToolExecutor).geocodeTool,
	"places.search":            (*AssistantToolExecutor).searchPlaces,
	"weather.forecast":         (*AssistantToolExecutor).weatherForecast,
	"weather.history":          (*AssistantToolExecutor).weatherHistory,
	"weather.air_quality":      (*AssistantToolExecutor).airQuality,
	"currency.convert":         (*AssistantToolExecutor).convertCurrency,
	"markets.quote":            (*AssistantToolExecutor).stockQuote,
	"markets.crypto":           (*AssistantToolExecutor).cryptoPrice,
	"calendar.public_holidays": (*AssistantToolExecutor).publicHolidays,
	"knowledge.wikipedia":      (*AssistantToolExecutor).wikipedia,
	"knowledge.define":         (*AssistantToolExecutor).defineWord,
	"knowledge.country":        (*AssistantToolExecutor).countryFactsTool,
	"knowledge.papers":         (*AssistantToolExecutor).searchPapers,
	"utilities.date_math":      (*AssistantToolExecutor).dateMath,
	"utilities.convert_time":   (*AssistantToolExecutor).convertTime,
	"utilities.convert_units":  (*AssistantToolExecutor).convertUnits,
	"utilities.statistics":     (*AssistantToolExecutor).statistics,
	"utilities.business_math":  (*AssistantToolExecutor).businessMath,
	"text.count":               (*AssistantToolExecutor).countText,
	"utilities.random":         (*AssistantToolExecutor).randomPick,
}

// MaxWebSearchesPerMessage lets Caret refine a search (a narrower query, a
// different angle) without turning one message into a crawl.
const MaxWebSearchesPerMessage = 5

// WebSearchOptions narrow one search. They travel on the context so every
// backend in the fallback chain sees them without widening the interface.
type WebSearchOptions struct {
	// Recency is "day", "week", "month" or "year"; empty means any time.
	Recency string
	// News searches news sources instead of the general web.
	News bool
}

type webSearchOptionsKey struct{}

func withWebSearchOptions(ctx context.Context, opts WebSearchOptions) context.Context {
	return context.WithValue(ctx, webSearchOptionsKey{}, opts)
}

func webSearchOptionsFrom(ctx context.Context) WebSearchOptions {
	opts, _ := ctx.Value(webSearchOptionsKey{}).(WebSearchOptions)
	return opts
}

func parseWebSearchOptions(input map[string]any) WebSearchOptions {
	opts := WebSearchOptions{News: inputBool(input, "news") || strings.EqualFold(inputText(input, "type", "kind", "category"), "news")}
	switch recency := strings.ToLower(inputText(input, "recency", "time_range", "freshness", "since")); {
	case strings.HasPrefix(recency, "day"), recency == "24h", recency == "today":
		opts.Recency = "day"
	case strings.HasPrefix(recency, "week"), recency == "7d":
		opts.Recency = "week"
	case strings.HasPrefix(recency, "month"), recency == "30d":
		opts.Recency = "month"
	case strings.HasPrefix(recency, "year"):
		opts.Recency = "year"
	}
	return opts
}

type webSearchBudgetKey struct{}
type webSearchBudget struct{ used atomic.Int32 }

func WithWebSearchBudget(ctx context.Context) context.Context {
	return context.WithValue(ctx, webSearchBudgetKey{}, &webSearchBudget{})
}

// WithReplyBudgets starts the per-message budgets for web searches and
// sub-agent runs.
func WithReplyBudgets(ctx context.Context) context.Context {
	return WithAgentRunBudget(WithWebSearchBudget(ctx))
}

func NewAssistantToolExecutor(profile AssistantProfileReader, web PublicWebSearcher) *AssistantToolExecutor {
	if web == nil {
		client := &http.Client{Timeout: 12 * time.Second}
		web = FallbackWebSearch{
			Primary:  SearXNGWebSearch{Client: client},
			Fallback: BraveWebSearch{Client: client},
		}
	}
	return &AssistantToolExecutor{Profile: profile, Web: web, HTTP: NewPublicHTTPClient(12 * time.Second)}
}

func (e *AssistantToolExecutor) httpClient() *http.Client {
	if e.HTTP == nil {
		e.HTTP = NewPublicHTTPClient(12 * time.Second)
	}
	return e.HTTP
}

func (e *AssistantToolExecutor) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case "account.profile":
		if e.Profile == nil {
			return nil, true, errors.New("profile service is unavailable")
		}
		profile, err := e.Profile.GetAssistantProfile(ctx)
		return map[string]any{"profile": profile}, true, err
	case ToolSearchName:
		query, _ := input["query"].(string)
		if strings.TrimSpace(query) == "" {
			return nil, true, errors.New("describe the tool you need")
		}
		mode := ModeAgent
		if IsSubAgent(ctx) {
			// Sub-agents only read, so they only discover read tools.
			mode = ModeAsk
		}
		return toolSearchOutput(mode, truncateSearchText(query, 300)), true, nil
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
		if strings.TrimSpace(zone) == "" {
			// No zone named: the person's own.
			location, err = userLocation(ctx), nil
			zone = location.String()
		}
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
		if budget, ok := ctx.Value(webSearchBudgetKey{}).(*webSearchBudget); ok && budget.used.Add(1) > MaxWebSearchesPerMessage {
			return nil, true, fmt.Errorf("%d web searches are available per message; answer from the results you have", MaxWebSearchesPerMessage)
		}
		result, err := e.Web.Search(withWebSearchOptions(ctx, parseWebSearchOptions(input)), query)
		return result, true, err
	default:
		run, ok := generalTools[name]
		if !ok {
			return nil, false, nil
		}
		output, err := run(e, ctx, input)
		return output, true, err
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
	node, err := parser.ParseExpr(prepareArithmetic(expression))
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
	case *ast.Ident:
		switch strings.ToLower(node.Name) {
		case "pi":
			return math.Pi, nil
		case "e":
			return math.E, nil
		}
		return 0, fmt.Errorf("unknown name %q", node.Name)
	case *ast.CallExpr:
		return evalArithmeticCall(node)
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

var (
	percentLiteral = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*%`)
	percentOf      = regexp.MustCompile(`(?i)\)\s*of\s+`)
)

// prepareArithmetic accepts everyday notation: "15% of 200", ×, ÷ and
// thousands separators like 1,250.50 outside function calls.
func prepareArithmetic(expression string) string {
	expression = strings.NewReplacer("×", "*", "÷", "/", "−", "-").Replace(expression)
	if !strings.Contains(expression, "(") {
		expression = thousandsSeparator.ReplaceAllString(expression, "$1$2")
	}
	expression = percentLiteral.ReplaceAllString(expression, "($1/100)")
	return percentOf.ReplaceAllString(expression, ")*")
}

var thousandsSeparator = regexp.MustCompile(`(\d),(\d{3})`)

var arithmeticFunctions = map[string]func(args []float64) (float64, error){
	"sqrt":  unary(math.Sqrt),
	"abs":   unary(math.Abs),
	"floor": unary(math.Floor),
	"ceil":  unary(math.Ceil),
	"exp":   unary(math.Exp),
	"ln":    unary(math.Log),
	"log10": unary(math.Log10),
	"log2":  unary(math.Log2),
	"log":   unary(math.Log10),
	"sin":   unary(math.Sin),
	"cos":   unary(math.Cos),
	"tan":   unary(math.Tan),
	"pow": func(args []float64) (float64, error) {
		if len(args) != 2 {
			return 0, errors.New("pow takes two numbers")
		}
		return math.Pow(args[0], args[1]), nil
	},
	"round": func(args []float64) (float64, error) {
		switch len(args) {
		case 1:
			return math.Round(args[0]), nil
		case 2:
			return roundTo(args[0], int(args[1])), nil
		}
		return 0, errors.New("round takes a number and optional decimal places")
	},
	"min": func(args []float64) (float64, error) {
		if len(args) == 0 {
			return 0, errors.New("min needs numbers")
		}
		out := args[0]
		for _, v := range args[1:] {
			out = math.Min(out, v)
		}
		return out, nil
	},
	"max": func(args []float64) (float64, error) {
		if len(args) == 0 {
			return 0, errors.New("max needs numbers")
		}
		out := args[0]
		for _, v := range args[1:] {
			out = math.Max(out, v)
		}
		return out, nil
	},
}

func unary(fn func(float64) float64) func([]float64) (float64, error) {
	return func(args []float64) (float64, error) {
		if len(args) != 1 {
			return 0, errors.New("this function takes one number")
		}
		return fn(args[0]), nil
	}
}

func evalArithmeticCall(call *ast.CallExpr) (float64, error) {
	name, ok := call.Fun.(*ast.Ident)
	if !ok {
		return 0, errors.New("unsupported function call")
	}
	fn, ok := arithmeticFunctions[strings.ToLower(name.Name)]
	if !ok {
		return 0, fmt.Errorf("unknown function %q", name.Name)
	}
	if len(call.Args) > 64 {
		return 0, errors.New("too many arguments")
	}
	args := make([]float64, len(call.Args))
	for i, arg := range call.Args {
		value, err := evalArithmetic(arg)
		if err != nil {
			return 0, err
		}
		args[i] = value
	}
	return fn(args)
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
	opts := webSearchOptionsFrom(ctx)
	if opts.News {
		params.Set("categories", "news")
	}
	if opts.Recency != "" {
		params.Set("time_range", opts.Recency)
	}
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
			Published  string `json:"publishedDate"`
		} `json:"results"`
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		return nil, errors.New("web search returned an invalid response")
	}
	results := make([]webSearchResult, 0, len(response.Results))
	for _, result := range response.Results {
		snippet := strings.TrimSpace(result.Content)
		if snippet == "" {
			snippet = strings.TrimSpace(result.Suggestion)
		}
		results = append(results, webSearchResult{Title: result.Title, URL: result.URL, Snippet: snippet, Published: result.Published})
	}
	return searchSourcesOutput(results, query)
}

// webSearchResult is one hit from any search backend before validation.
type webSearchResult struct {
	Title     string
	URL       string
	Snippet   string
	Published string
}

// searchSourcesOutput keeps up to eight distinct https results and shapes them
// as the web.search tool output.
func searchSourcesOutput(results []webSearchResult, query string) (map[string]any, error) {
	sources := make([]map[string]string, 0, 8)
	seen := make(map[string]bool)
	for _, result := range results {
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
		source := map[string]string{
			"title":   title,
			"url":     parsedURL.String(),
			"snippet": truncateSearchText(strings.TrimSpace(result.Snippet), 1200),
		}
		if published := strings.TrimSpace(result.Published); published != "" && len(published) <= 40 {
			source["published"] = published
		}
		sources = append(sources, source)
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
