package application

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// CaretIntentKind is the server-owned interpretation of a user turn. The
// model still writes the answer, but it no longer gets to decide whether a
// request is casual chat, a live-data question, a lookup, or an action without
// the backend first describing the job it is being asked to do.
type CaretIntentKind string

const (
	IntentConversation CaretIntentKind = "conversation"
	IntentSnapshot     CaretIntentKind = "business_snapshot"
	IntentAnalytics    CaretIntentKind = "analytics"
	IntentLookup       CaretIntentKind = "lookup"
	IntentAction       CaretIntentKind = "action"
	IntentMemory       CaretIntentKind = "memory"
	IntentSupport      CaretIntentKind = "support"
	IntentNavigation   CaretIntentKind = "navigation"
)

// CaretRequestPlan is deliberately serializable as plain text. It is used by
// prompt construction, cache policy, grounding, tool shortlisting, and stream
// policy, so those layers all make decisions from one interpretation instead
// of maintaining independent keyword lists.
type CaretRequestPlan struct {
	Kind               CaretIntentKind
	Domain             string
	Action             string
	Scope              string
	Freshness          string
	Confidence         int
	Terms              []string
	NeedsLiveData      bool
	NeedsFreshData     bool
	NeedsTool          bool
	NeedsApproval      bool
	IsBroad            bool
	IsSmokeCheck       bool
	IsConversational   bool
	NeedsClarification bool
	EntityType         string
	HasEntityContext   bool
}

var requestEntityIDPattern = regexp.MustCompile(`(?i)(?:^|[_\s-])(customer|product|order|invoice|employee|supplier|task|appointment|campaign|segment)[_\s-]?(?:id)?\s*[:=]\s*[a-z0-9_-]+`)

var requestDomainTerms = map[string][]string{
	"booking":   {"appointment", "booking", "calendar", "service", "resource", "waitlist", "reminder"},
	"customer":  {"customer", "customers", "crm", "campaign", "campaigns", "segment", "segments", "churn", "ltv"},
	"finance":   {"invoice", "invoices", "expense", "expenses", "bank", "payment", "payments", "bill", "bills", "quote", "tax", "cash", "balance", "profit", "loss", "revenue", "margin"},
	"flow":      {"task", "tasks", "project", "projects", "roadmap", "workload", "milestone"},
	"hr":        {"employee", "employees", "staff", "payroll", "leave", "overtime", "attendance", "timesheet", "shift", "department", "role"},
	"inventory": {"inventory", "stock", "supplier", "suppliers", "warehouse", "warehouses", "purchase", "transfer", "brand", "category", "categories", "product", "products", "reorder", "low stock"},
	"loyalty":   {"loyalty", "points", "reward", "rewards", "member", "members", "tier"},
	"pos":       {"pos", "order", "orders", "terminal", "register", "floor", "table", "menu", "modifier", "session", "cash", "refund", "sale", "sales"},
	"store":     {"storefront", "ecommerce", "online store", "website", "catalog", "slug"},
	"support":   {"support", "ticket", "tickets", "inbox", "canned", "article", "knowledge", "unread", "help"},
	"web":       {"web", "internet", "online search", "search the web", "search online", "news", "research", "sources"},
}

var requestActionTerms = map[string][]string{
	"create":    {"create", "add", "new", "record", "make", "draft"},
	"read":      {"show", "list", "get", "open", "view", "find", "lookup", "check", "see", "which"},
	"compare":   {"compare", "versus", "vs", "difference", "change", "growth", "trend"},
	"summarize": {"summary", "summarize", "overview", "brief", "snapshot", "dashboard", "health", "how are", "how is", "what stands out"},
	"recommend": {"what should", "next", "focus", "attention", "recommend", "recommendation", "suggest", "prioritize", "opportunity", "risk"},
	"remember":  {"remember", "save", "keep in mind", "note that", "always", "never", "prefer", "default", "from now on"},
	"schedule":  {"schedule", "scheduled", "recurring", "repeat", "daily", "weekly", "monthly", "alert", "notify", "remind", "watch"},
	"navigate":  {"where do i", "how do i", "how can i", "find the setting", "take me", "open the"},
}

var requestFreshnessTerms = map[string]string{
	"today": "today", "now": "now", "current": "current", "latest": "latest", "this week": "this week", "this month": "this month", "year to date": "year to date", "ytd": "year to date", "last week": "last week", "last month": "last month", "this quarter": "this quarter",
}

var smokeCheckRequests = map[string]struct{}{
	"test": {}, "testing": {}, "ping": {}, "hello": {}, "hi": {}, "hey": {},
	"are you working": {}, "is this working": {}, "does this work": {},
}

var socialGreetingPattern = regexp.MustCompile(`^(?:(?:hey|hi|hello|howdy)(?: there)?(?: how are you(?: doing)?| how is it going| what is up)?|how are you(?: doing)?|how is it going|what is up)(?: today)?(?: (?:lil man|my friend))?$`)

// ClassifyCaretRequest turns the current request and current page context into
// one deterministic execution plan. It intentionally favors live evidence for
// business language and conversation handling for social language; the model
// is not asked to infer that distinction from a blank prompt.
func ClassifyCaretRequest(request, systemContext string) CaretRequestPlan {
	text := normalizeIntentText(request)
	contextText := normalizeIntentText(systemContext)
	plan := CaretRequestPlan{
		Kind:             IntentConversation,
		Action:           "answer",
		Scope:            "current request",
		Freshness:        "not specified",
		Confidence:       42,
		IsConversational: true,
	}
	if text == "" {
		plan.Confidence = 0
		return plan
	}

	if _, ok := smokeCheckRequests[text]; ok {
		plan.IsSmokeCheck = true
		plan.Confidence = 100
		plan.Action = "acknowledge"
		return plan
	}
	if socialGreetingPattern.MatchString(text) {
		plan.Confidence = 100
		return plan
	}

	plan.Domain = strongestRequestDomain(text, contextText)
	plan.Action = strongestRequestAction(text)
	if plan.Action == "answer" && plan.Domain != "" {
		plan.Action = "read"
	}
	plan.Freshness = requestFreshness(text)
	plan.NeedsFreshData = plan.Freshness != "not specified"
	plan.Terms = requestSignalTerms(text)
	plan.HasEntityContext = hasIntentEntityContext(contextText)
	plan.EntityType = intentEntityType(text)

	if plan.Action == "navigate" {
		plan.Kind = IntentNavigation
		plan.IsConversational = false
		plan.Confidence += 22
	} else if plan.Action == "remember" {
		plan.Kind = IntentMemory
		plan.IsConversational = false
		plan.NeedsTool = true
		plan.NeedsApproval = true
		plan.Confidence += 28
	} else if plan.Action == "create" || plan.Action == "schedule" || containsAnyIntentAction(plan.Action, "mutate") {
		plan.Kind = IntentAction
		plan.IsConversational = false
		plan.NeedsTool = true
		plan.NeedsApproval = true
		plan.Confidence += 28
	} else if plan.Domain == "support" && plan.Action == "navigate" {
		plan.Kind = IntentSupport
		plan.IsConversational = false
		plan.Confidence += 18
	} else if plan.Action == "summarize" && (plan.Domain == "" || plan.Domain == "business") {
		plan.Kind = IntentSnapshot
		plan.IsBroad = true
		plan.NeedsLiveData = true
		plan.NeedsTool = true
		plan.IsConversational = false
		plan.Confidence += 32
	} else if plan.Domain != "" || plan.Action == "read" || plan.Action == "compare" || plan.Action == "recommend" {
		plan.Kind = IntentAnalytics
		plan.NeedsLiveData = true
		plan.NeedsTool = true
		plan.IsConversational = false
		plan.Confidence += 26
	}

	if isBroadIntentRequest(text) {
		plan.IsBroad = true
		plan.Kind = IntentSnapshot
		plan.NeedsLiveData = true
		plan.NeedsTool = true
		plan.IsConversational = false
		plan.Confidence += 10
	}
	if plan.NeedsFreshData {
		plan.NeedsLiveData = true
	}
	if plan.Domain == "" && plan.NeedsLiveData {
		plan.Domain = "business"
	}
	if plan.EntityType != "" && !plan.HasEntityContext && hasIntentEntityPointer(text) {
		plan.NeedsClarification = true
		plan.Confidence -= 12
	}
	if plan.Confidence > 100 {
		plan.Confidence = 100
	}
	if plan.Confidence < 0 {
		plan.Confidence = 0
	}
	plan.Scope = intentScope(plan, contextText)
	return plan
}

func (p CaretRequestPlan) PromptText() string {
	lines := []string{
		"- Intent: " + string(p.Kind),
		"- Domain: " + fallbackIntentValue(p.Domain, "general business"),
		"- Action: " + p.Action,
		"- Scope: " + p.Scope,
		"- Freshness: " + p.Freshness,
		"- Confidence: " + stringInt(p.Confidence) + "%",
	}
	if p.NeedsLiveData {
		lines = append(lines, "- Live business evidence is required before stating current facts.")
	}
	if p.NeedsTool {
		lines = append(lines, "- Use the server capability catalog when a listed read is needed.")
	}
	if p.NeedsApproval {
		lines = append(lines, "- Any mutation requires an explicit approval step.")
	}
	if p.IsBroad {
		lines = append(lines, "- This is a broad operator request; synthesize the strongest supported signals and concrete next moves.")
	}
	if p.NeedsClarification {
		lines = append(lines, "- The referenced entity is not identified; ask one short clarifying question before acting.")
	}
	return strings.Join(lines, "\n")
}

func strongestRequestDomain(text, contextText string) string {
	scores := make(map[string]int)
	for domain, terms := range requestDomainTerms {
		for _, term := range terms {
			if containsIntentTerm(text, term) {
				scores[domain] += 3
			}
			if containsIntentTerm(contextText, term) {
				scores[domain]++
			}
		}
	}
	best := ""
	bestScore := 0
	for domain, score := range scores {
		if score > bestScore || (score == bestScore && domain < best) {
			best, bestScore = domain, score
		}
	}
	return best
}

func strongestRequestAction(text string) string {
	best := "answer"
	bestScore := 0
	for action, terms := range requestActionTerms {
		score := 0
		for _, term := range terms {
			if containsIntentTerm(text, term) {
				score += len(term)
			}
		}
		if score > bestScore {
			best, bestScore = action, score
		}
	}
	if best == "answer" && (containsAnyIntentAction(text, "update", "change", "set", "send", "approve", "receive", "record") ||
		containsIntentTerm(text, "issue refund") || containsIntentTerm(text, "refund order") || containsIntentTerm(text, "process refund")) {
		return "mutate"
	}
	return best
}

func requestFreshness(text string) string {
	best := "not specified"
	for term, label := range requestFreshnessTerms {
		if containsIntentTerm(text, term) && (best == "not specified" || len(term) > len(best)) {
			best = label
		}
	}
	return best
}

func requestSignalTerms(text string) []string {
	set := make(map[string]struct{})
	for _, token := range strings.Fields(text) {
		if len([]rune(token)) >= 3 && !intentStopWords[token] {
			set[token] = struct{}{}
		}
	}
	terms := make([]string, 0, len(set))
	for term := range set {
		terms = append(terms, term)
	}
	sort.Strings(terms)
	if len(terms) > 12 {
		terms = terms[:12]
	}
	return terms
}

func isBroadIntentRequest(text string) bool {
	for _, phrase := range []string{"what should i do", "what do i do next", "what needs attention", "what stands out", "give me a brief", "give me an overview", "business health", "how are we doing", "how is my business", "what should i focus on", "help me with this page"} {
		if containsIntentTerm(text, phrase) {
			return true
		}
	}
	return false
}

func hasIntentEntityPointer(text string) bool {
	for _, pointer := range []string{"this", "that", "these", "those", "selected", "it", "the current"} {
		if containsIntentTerm(text, pointer) {
			return true
		}
	}
	return false
}

func hasIntentEntityContext(contextText string) bool {
	return requestEntityIDPattern.MatchString(contextText) || containsIntentTerm(contextText, "selected") || containsIntentTerm(contextText, "current row") || containsIntentTerm(contextText, "current item")
}

func intentEntityType(text string) string {
	for _, entity := range []string{"customer", "product", "order", "invoice", "employee", "supplier", "task", "appointment", "campaign", "segment"} {
		if containsIntentTerm(text, entity) {
			return entity
		}
	}
	return ""
}

func intentScope(plan CaretRequestPlan, contextText string) string {
	if plan.HasEntityContext {
		return "selected entity and current page scope"
	}
	if containsIntentTerm(contextText, "route") || containsIntentTerm(contextText, "module") {
		return "current page and authenticated business"
	}
	return "authenticated business"
}

func containsAnyIntentAction(text string, actions ...string) bool {
	for _, action := range actions {
		if containsIntentTerm(text, action) {
			return true
		}
	}
	return false
}

func containsIntentTerm(text, term string) bool {
	term = normalizeIntentText(term)
	if term == "" {
		return false
	}
	if strings.Contains(term, " ") {
		return strings.Contains(text, term)
	}
	for _, token := range strings.Fields(text) {
		if token == term {
			return true
		}
	}
	return false
}

func normalizeIntentText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' {
			return r
		}
		return ' '
	}, value)
	return strings.Join(strings.Fields(value), " ")
}

func fallbackIntentValue(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func stringInt(value int) string {
	return strconv.Itoa(value)
}

var intentStopWords = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "from": true, "what": true, "when": true, "where": true, "which": true, "how": true, "are": true, "is": true, "my": true, "our": true, "your": true, "this": true, "that": true,
}
