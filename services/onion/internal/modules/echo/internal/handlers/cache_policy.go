package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"unicode"
)

var liveBusinessDataTerms = []string{
	"active",
	"appointment",
	"appointments",
	"balance",
	"billing",
	"cash",
	"customer",
	"customers",
	"discount",
	"employee",
	"employees",
	"expense",
	"expenses",
	"gift card",
	"inventory",
	"invoice",
	"invoices",
	"kds",
	"loyalty",
	"low stock",
	"order",
	"orders",
	"payment",
	"payments",
	"payroll",
	"product",
	"products",
	"receipt",
	"receipts",
	"refund",
	"revenue",
	"sale",
	"sales",
	"seat",
	"seats",
	"session",
	"sessions",
	"stock",
	"supplier",
	"suppliers",
	"table",
	"tables",
	"warehouse",
	"warehouses",
}

var liveBusinessIntentTerms = []string{
	"amount",
	"any",
	"average",
	"count",
	"current",
	"best",
	"how many",
	"how much",
	"latest",
	"least",
	"list",
	"low",
	"most",
	"open",
	"show",
	"today",
	"top",
	"total",
	"which",
}

func shouldGroundBeforeExactCache(mode CaretMode, prompt string, pageContext map[string]any) bool {
	if mode == ModeSupport {
		return false
	}

	normalized := normalizeCachePolicyText(prompt)
	if normalized == "  " {
		return false
	}

	hasBusinessTerm := containsAnyTerm(normalized, liveBusinessDataTerms)
	if !hasBusinessTerm {
		return false
	}

	if containsAnyTerm(normalized, liveBusinessIntentTerms) {
		return true
	}

	if hasDynamicPageContext(pageContext) {
		return true
	}

	return false
}

func hasDynamicPageContext(pageContext map[string]any) bool {
	if len(pageContext) == 0 {
		return false
	}
	for _, key := range []string{"dataFreshness", "data_freshness", "dataVersion", "dataVersions", "cacheContext", "cache_context", "freshness", "selected", "filters", "visibleRows"} {
		if _, ok := pageContext[key]; ok {
			return true
		}
	}
	return false
}

func containsAnyTerm(text string, terms []string) bool {
	for _, term := range terms {
		term = strings.ToLower(strings.TrimSpace(term))
		if term == "" {
			continue
		}
		if strings.Contains(term, " ") {
			if strings.Contains(text, term) {
				return true
			}
			continue
		}
		if strings.Contains(text, " "+term+" ") {
			return true
		}
	}
	return false
}

func normalizeCachePolicyText(value string) string {
	normalized := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		}
		return ' '
	}, value)
	return " " + strings.Join(strings.Fields(normalized), " ") + " "
}

func buildGroundingCacheFreshness(grounding *analystGroundingResult) string {
	if grounding == nil {
		return ""
	}

	parts := []string{
		"tool=" + grounding.toolName(),
		"rows=" + shortCacheHash(grounding.RowsJSON),
		"spec=" + shortCacheHash(grounding.SpecJSON),
		"count=" + strconv.Itoa(grounding.RowCount),
	}
	if grounding.Truncated {
		parts = append(parts, "truncated=true")
	}
	return strings.Join(parts, ";")
}

func (grounding *analystGroundingResult) toolName() string {
	if grounding == nil {
		return "analytics.query"
	}
	name := strings.TrimSpace(grounding.ToolName)
	if name == "" {
		return "analytics.query"
	}
	return name
}

func combineCacheFreshness(pageFreshness, groundingFreshness string) string {
	pageFreshness = strings.TrimSpace(pageFreshness)
	groundingFreshness = strings.TrimSpace(groundingFreshness)
	switch {
	case pageFreshness == "":
		return groundingFreshness
	case groundingFreshness == "":
		return pageFreshness
	default:
		return pageFreshness + "\n" + groundingFreshness
	}
}

func shortCacheHash(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])[:16]
}
