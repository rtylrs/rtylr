package application

import (
	"strings"
	"unicode"
)

const (
	DefaultToolCallBudget     = 6
	DefaultContinuationBudget = 2
)

var (
	heavyDataActionTerms     = []string{"download", "dump", "every", "export", "raw"}
	heavyDataCollectionTerms = []string{"appointment", "appointments", "attendance", "bill", "bills", "customer", "customers", "employee", "employees", "expense", "expenses", "invoice", "invoices", "order", "orders", "payment", "payments", "product", "products", "receipt", "receipts", "refund", "refunds", "sales", "session", "sessions", "shift", "shifts", "stock movement", "stock movements", "transaction", "transactions"}
	heavyDataRangeTerms      = []string{"all time", "all-time", "entire history", "full history", "last year", "past year", "since we opened"}
)

func HeavyDataRequestReason(question string) (string, bool) {
	normalized := normalizeRequestPolicyText(question)
	if strings.TrimSpace(normalized) == "" {
		return "", false
	}
	if containsRequestTerm(normalized, []string{"csv", "spreadsheet"}) && containsRequestTerm(normalized, heavyDataCollectionTerms) {
		return "export requested for a large business collection", true
	}
	if containsRequestTerm(normalized, heavyDataActionTerms) && containsRequestTerm(normalized, heavyDataCollectionTerms) {
		return "raw or complete business records requested", true
	}
	if containsRequestTerm(normalized, []string{"list all", "show all", "give me all"}) && containsRequestTerm(normalized, heavyDataCollectionTerms) {
		return "complete row listing requested", true
	}
	if containsRequestTerm(normalized, heavyDataRangeTerms) && containsRequestTerm(normalized, heavyDataCollectionTerms) && containsRequestTerm(normalized, []string{"list", "show", "export", "download"}) {
		return "wide historical row request", true
	}
	return "", false
}

func HasDynamicPageContext(pageContext map[string]any) bool {
	for _, key := range []string{"dataFreshness", "data_freshness", "dataVersion", "dataVersions", "cacheContext", "cache_context", "freshness", "selected", "filters", "visibleRows"} {
		if _, ok := pageContext[key]; ok {
			return true
		}
	}
	return false
}

func containsRequestTerm(text string, terms []string) bool {
	for _, term := range terms {
		term = strings.ToLower(strings.TrimSpace(term))
		if strings.Contains(term, " ") {
			if strings.Contains(text, term) {
				return true
			}
		} else if strings.Contains(text, " "+term+" ") {
			return true
		}
	}
	return false
}

func normalizeRequestPolicyText(value string) string {
	normalized := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		}
		return ' '
	}, value)
	return " " + strings.Join(strings.Fields(normalized), " ") + " "
}
