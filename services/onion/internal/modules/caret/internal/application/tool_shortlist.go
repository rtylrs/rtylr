package application

import (
	"sort"
	"strings"
)

// The shortlist only seeds obvious matches; tools.search reaches the rest,
// so it stays small to keep each prompt light.
const promptToolShortlistLimit = 8

var promptToolGlobalNames = map[string]struct{}{
	"business.snapshot":      {},
	"business.get_kpis":      {},
	"insights.get_dashboard": {},
}

var promptToolModuleKeywords = map[string][]string{
	"booking": {"appointment", "booking", "calendar", "service", "resource", "waitlist", "reminder"},
	"crm":     {"customer", "crm", "campaign", "segment", "lead", "marketing"},
	"documents": {
		"document", "documents", "form", "forms", "fill", "filled", "filling", "fill out", "fill in",
		"paperwork", "pdf", "scan", "scanned", "extract", "transcribe", "intake", "onboarding sheet",
	},
	"erp":     {"inventory", "stock", "supplier", "warehouse", "purchase", "transfer", "brand", "category", "product", "reorder"},
	"finance": {"invoice", "expense", "bank", "payment", "bill", "quote", "tax", "cash", "balance", "profit", "loss", "revenue", "payable", "receivable"},
	"flow":    {"task", "project", "roadmap", "workload", "flow", "milestone"},
	"hr":      {"employee", "staff", "hr", "payroll", "leave", "overtime", "attendance", "timesheet", "shift", "department", "role"},
	"insights": {
		"insight", "dashboard", "kpi", "metric", "trend", "sales", "revenue", "churn", "margin", "bestseller",
		"slow mover", "headcount", "utilization", "schedule", "scheduled", "recurring", "daily", "weekly", "report", "watch", "alert",
	},
	"loyalty": {"loyalty", "points", "reward", "member", "tier"},
	"memory":  {"remember", "memory", "preference", "prefer", "default", "always", "never", "rule", "note", "save", "keep in mind"},
	"pos":     {"pos", "order", "terminal", "register", "floor", "table", "menu", "modifier", "session", "cash", "branch", "refund", "sale"},
	"store":   {"store", "storefront", "ecommerce", "online", "website", "catalog", "slug"},
	"support": {"support", "ticket", "inbox", "canned", "article", "knowledge", "kb", "help", "unread"},
	"web":     {"web", "internet", "search", "online", "research", "news", "source", "sources"},
}

var promptToolActionKeywords = map[string][]string{
	"create": {"create", "add", "new", "record", "make", "draft"},
	"get":    {"get", "show", "open", "view", "detail", "details", "find", "lookup", "load"},
	"list":   {"list", "all", "recent", "top", "show me", "which"},
	"remember": {
		"remember", "save", "keep in mind", "note that", "always", "never", "prefer", "default",
		"from now on", "next time",
	},
	"mutate": {"approve", "receive", "send", "refund", "set", "update", "change", "move", "transfer", "delete", "remove"},
	"schedule": {
		"schedule", "scheduled", "recurring", "repeat", "daily", "weekly", "monthly", "every day", "every monday",
		"alert", "notify", "remind", "watch", "send me",
	},
	"summary": {"summary", "summarize", "overview", "status", "how are", "how is"},
}

var promptToolProductLookupKeywords = map[string][]string{
	"erp.list_categories": {"category", "categories"},
	"erp.list_suppliers":  {"supplier", "suppliers"},
	"erp.list_brands":     {"brand", "brands"},
}

type promptToolScore struct {
	tool  ToolDef
	score int
	index int
}

// PromptToolsForRequest returns a deterministic request-aware subset of the
// executable tool catalog. The full registry remains available elsewhere, but
// the model prompt should stay focused so tool choice does not get noisy.
func PromptToolsForRequest(mode CaretMode, request, systemContext string) []ToolDef {
	tools := PromptToolsForMode(mode)
	if len(tools) == 0 {
		return nil
	}

	plan := ClassifyCaretRequest(request, systemContext)
	if plan.Kind == IntentConversation {
		// The classifier found no business domain, action, or freshness signal
		// at all. Offering tools here just tempts the model into calling a
		// business capability for small talk or general-knowledge questions
		// that have nothing to do with this business's own data.
		return nil
	}

	text := normalizePromptToolText(request + " " + systemContext)
	scored := make([]promptToolScore, 0, len(tools))
	for i, tool := range tools {
		score := promptToolScoreForText(tool, text) + promptToolScoreForPlan(tool, plan)
		if score > 0 {
			scored = append(scored, promptToolScore{tool: tool, score: score, index: i})
		}
	}

	if len(scored) == 0 && plan.NeedsTool {
		for i, tool := range tools {
			if _, ok := promptToolGlobalNames[tool.Name]; ok {
				scored = append(scored, promptToolScore{tool: tool, score: 1000, index: i})
			}
		}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].index < scored[j].index
		}
		return scored[i].score > scored[j].score
	})

	if len(scored) > promptToolShortlistLimit {
		scored = scored[:promptToolShortlistLimit]
	}

	out := make([]ToolDef, 0, len(scored))
	for _, item := range scored {
		out = append(out, item.tool)
	}
	return out
}

func promptToolScoreForPlan(tool ToolDef, plan CaretRequestPlan) int {
	score := 0
	if plan.NeedsTool {
		if _, ok := promptToolGlobalNames[tool.Name]; ok {
			score += 700
		}
	}
	module := map[string]string{
		"booking": "booking", "customer": "crm", "finance": "finance", "flow": "flow", "hr": "hr", "inventory": "erp", "loyalty": "loyalty", "pos": "pos", "store": "store", "support": "support", "web": "web",
	}[plan.Domain]
	if module != "" && strings.EqualFold(module, tool.Module) {
		score += 280
	}
	if plan.Action == "create" && tool.Scope == "write" {
		score += 180
	}
	if (plan.Action == "read" || plan.Action == "compare" || plan.Action == "summarize" || plan.Action == "recommend") && tool.Scope == "read" {
		score += 80
	}
	if plan.Action == "remember" && tool.Name == "memory.remember" {
		score += 300
	}
	if tool.Name == "analytics.query" && plan.Kind == IntentAnalytics {
		// This is the generic read fallback for questions that do not map to a
		// narrower module tool. The server validates the generated spec and
		// enforces the tenant boundary before returning any rows.
		score += 700
	}
	return score
}

func promptToolScoreForText(tool ToolDef, text string) int {
	score := 0

	module := strings.ToLower(tool.Module)
	if keywords := promptToolModuleKeywords[module]; promptToolTextContainsAny(text, keywords) {
		score += 250
	}

	nameText := normalizePromptToolText(strings.ReplaceAll(tool.Name, ".", " "))
	descText := normalizePromptToolText(tool.Description)
	for _, token := range strings.Fields(text) {
		if len(token) < 3 {
			continue
		}
		if strings.Contains(nameText, token) {
			score += 28
		}
		if strings.Contains(descText, token) {
			score += 12
		}
	}

	if strings.Contains(tool.Name, ".create_") || strings.Contains(tool.Name, ".record_") {
		if promptToolTextContainsAny(text, promptToolActionKeywords["create"]) {
			score += 90
		}
	}
	if strings.Contains(tool.Name, ".get_") {
		if promptToolTextContainsAny(text, promptToolActionKeywords["get"]) {
			score += 70
		}
	}
	if strings.Contains(tool.Name, ".update_") {
		if promptToolTextContainsAny(text, promptToolActionKeywords["mutate"]) {
			score += 75
		}
	}
	if strings.Contains(tool.Name, ".delete_") {
		if promptToolTextContainsAny(text, promptToolActionKeywords["mutate"]) {
			score += 75
		}
	}
	if strings.Contains(tool.Name, ".list_") {
		if promptToolTextContainsAny(text, promptToolActionKeywords["list"]) {
			score += 55
		}
	}
	if tool.Scope == "write" && promptToolTextContainsAny(text, promptToolActionKeywords["mutate"]) {
		score += 80
	}
	if tool.Name == "memory.remember" && promptToolTextContainsAny(text, promptToolActionKeywords["remember"]) {
		score += 180
	}
	if lookupKeywords, ok := promptToolProductLookupKeywords[tool.Name]; ok && promptToolTextContainsAny(text, lookupKeywords) {
		// These reads are preflight dependencies for a product write. Keep them
		// in the shortlist even when the request also scores many write tools.
		score += 500
	}
	if strings.Contains(tool.Name, "schedule") && promptToolTextContainsAny(text, promptToolActionKeywords["schedule"]) {
		score += 140
	}
	if promptToolTextContainsAny(text, promptToolActionKeywords["summary"]) &&
		(strings.Contains(tool.Name, "summary") || strings.Contains(tool.Name, "dashboard") || strings.Contains(tool.Name, "kpis")) {
		score += 80
	}

	return score
}

func promptToolTextContainsAny(text string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(text, normalizePromptToolText(needle)) {
			return true
		}
	}
	return false
}

func normalizePromptToolText(value string) string {
	value = strings.ToLower(value)
	replacer := strings.NewReplacer(
		"_", " ",
		"-", " ",
		".", " ",
		"/", " ",
		":", " ",
		",", " ",
		"?", " ",
		"!", " ",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
	)
	return strings.Join(strings.Fields(replacer.Replace(value)), " ")
}
