package handlers

import (
	"regexp"
	"strings"
)

const answerQualityHoldBytes = 96

var answerQualityWordReplacements = []struct {
	pattern *regexp.Regexp
	value   string
}{
	{regexp.MustCompile(`(?i)\byour tenant\b`), "your business"},
	{regexp.MustCompile(`(?i)\btenant\b`), "business"},
	{regexp.MustCompile(`(?i)\brows returned\b`), "results"},
	{regexp.MustCompile(`(?i)\brow count\b`), "result count"},
	{regexp.MustCompile(`(?i)\bserver-validated\b`), "checked"},
	{regexp.MustCompile(`(?i)\bread-only analytics query\b`), "business data"},
	{regexp.MustCompile(`(?i)\banalytics query\b`), "business data"},
	{regexp.MustCompile(`(?i)\bgrounding\b`), "context"},
}

var answerQualityPhraseReplacements = []struct {
	old string
	new string
}{
	{"Based on the provided analytics query, here are the results:", ""},
	{"Based on the provided analytics query, here are", "Here are"},
	{"Based on the provided analytics query,", ""},
	{"Based on the analytics query,", ""},
	{"Based on the provided query,", ""},
	{"Based on the data provided,", ""},
	{"Based on the query results,", ""},
	{"To answer your question, I ran a server-validated, read-only analytics query scoped to your tenant.", "I checked your business data."},
	{"To answer your question, I ran a read-only analytics query scoped to your tenant.", "I checked your business data."},
	{"To answer your question, I ran a checked, business data scoped to your business.", "I checked your business data."},
	{"To answer your question, I ran a business data scoped to your business.", "I checked your business data."},
	{"To answer your question, I ran", "I checked"},
	{"The query results are:", ""},
	{"Invalid tool call marker ignored.", ""},
	{"the revenue value is null", "I do not see revenue yet"},
	{"the value is null", "I do not see a value yet"},
	{"shortlisted capability catalog", "available business capabilities"},
	{"capability catalog", "available business capabilities"},
	{"customer data and product data tools", "customer and product data"},
}

var (
	internalToolNameRe         = regexp.MustCompile(`(?i)\b(?:business|memory|analytics|insights|crm|loyalty|pos|store|support|erp|hr|flow|finance|booking)\.[a-z][a-z0-9_]*\b`)
	partialInternalToolNameRe  = regexp.MustCompile(`(?i)\b(?:business|memory|analytics|insights|crm|loyalty|pos|store|support|erp|hr|flow|finance|booking)\.[a-z0-9_]*$`)
	customerToolNameRe         = regexp.MustCompile(`(?i)\b(?:insights\.(?:get_customer_kpis|list_top_customers|get_customer_churn|get_ltv_distribution)|crm\.[a-z][a-z0-9_]*)\b`)
	productToolNameRe          = regexp.MustCompile(`(?i)\b(?:insights\.(?:get_product_kpis|list_bestsellers|list_slow_movers|get_sales_by_category|get_margin_distribution)|(?:erp|store)\.[a-z][a-z0-9_]*)\b`)
	internalProcessNarrationRe = regexp.MustCompile(`(?is)\b(?:i )?(?:obtained|got|retrieved|found) (?:this )?information using .*?(?:tools?|capabilit(?:y|ies)|catalog)\b[^.]*\.?`)
	internalToolReferenceRe    = regexp.MustCompile(`(?i)\b(?:customer|product|business) data tools?\b`)
	internalNextActionRe       = regexp.MustCompile(`(?im)^\s*(?:next move|next step)\s*:\s*(?:run|call|execute|use)\b[^\n]*(?:\n|$)`)
	internalNextActionPrefixRe = regexp.MustCompile(`(?im)^\s*(?:next move|next step)\s*:\s*(?:run|call|execute|use)\b`)
)

// These terms are held back at the streaming boundary so a chunk cannot end
// in the middle of a private capability name or catalog phrase.
var answerQualityRedactionTerms = []string{
	"insights.get_customer_kpis",
	"insights.get_product_kpis",
	"shortlisted capability catalog",
	"capability catalog",
}

func polishCaretAnswerText(text string) string {
	if text == "" {
		return ""
	}

	out := text
	for _, replacement := range answerQualityPhraseReplacements {
		out = strings.ReplaceAll(out, replacement.old, replacement.new)
		out = strings.ReplaceAll(out, strings.ToLower(replacement.old), replacement.new)
	}
	for _, replacement := range answerQualityWordReplacements {
		out = replacement.pattern.ReplaceAllString(out, replacement.value)
	}
	out = internalProcessNarrationRe.ReplaceAllString(out, "I checked your business data.")
	out = internalNextActionRe.ReplaceAllString(out, "I can check the latest business data if you want.\n")

	// Capability names are implementation details. The model needs them for
	// the structured stream, but they must never cross the public answer boundary.
	out = customerToolNameRe.ReplaceAllString(out, "customer data")
	out = productToolNameRe.ReplaceAllString(out, "product data")
	out = internalToolNameRe.ReplaceAllString(out, "business data")
	out = internalToolReferenceRe.ReplaceAllString(out, "business check")
	out = removeCaretEmDashes(out)

	out = strings.ReplaceAll(out, "  ", " ")
	out = strings.ReplaceAll(out, "\n \n", "\n\n")
	return out
}

func removeCaretEmDashes(text string) string {
	text = strings.ReplaceAll(text, " — ", ", ")
	return strings.ReplaceAll(text, "—", ", ")
}

// safePublicAnswerPrefix prevents a streamed chunk from ending in an internal
// phrase. The held tail may complete it, but the current chunk must not reach
// the client before the public-answer sanitizer can redact it.
func safePublicAnswerPrefix(text string, maxLen int) int {
	if maxLen <= 0 || maxLen > len(text) {
		return maxLen
	}

	protected := strings.ToLower(text)
	limit := maxLen
	for _, term := range answerQualityRedactionTerms {
		term = strings.ToLower(term)
		start := strings.Index(protected, term)
		if start >= 0 && start < limit && start+len(term) > limit {
			limit = start
		}

		candidate := strings.ToLower(text[:limit])
		for suffixLen := len(term) - 1; suffixLen > 0 && suffixLen <= len(candidate); suffixLen-- {
			if strings.HasSuffix(candidate, term[:suffixLen]) {
				limit -= suffixLen
				break
			}
		}
	}
	if match := partialInternalToolNameRe.FindStringIndex(text[:limit]); match != nil && match[1] == limit {
		limit = match[0]
	}
	if match := internalNextActionPrefixRe.FindStringIndex(text); match != nil && match[0] < limit {
		lineEnd := strings.IndexByte(text[match[0]:], '\n')
		if lineEnd < 0 || match[0]+lineEnd+1 > limit {
			limit = match[0]
		}
	}
	return limit
}
