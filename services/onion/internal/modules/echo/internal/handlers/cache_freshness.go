package handlers

import (
	"encoding/json"
	"strings"
)

const maxCacheFreshnessLen = 1600

var cacheFreshnessKeys = []string{
	"dataFreshness",
	"data_freshness",
	"dataVersion",
	"dataVersions",
	"cacheContext",
	"cache_context",
	"freshness",
}

func buildCacheFreshness(pageContext map[string]any) string {
	if len(pageContext) == 0 {
		return ""
	}

	out := make(map[string]any)
	for _, key := range cacheFreshnessKeys {
		if value, ok := pageContext[key]; ok {
			if normalized := normalizePageContext(value, 0); normalized != nil {
				out[key] = normalized
			}
		}
	}

	for _, key := range []string{"module", "route"} {
		if value, ok := pageContext[key]; ok {
			if normalized := normalizePageContext(value, 0); normalized != nil {
				out[key] = normalized
			}
		}
	}

	for _, key := range []string{"selected", "filters", "visibleRows"} {
		if value, ok := pageContext[key]; ok {
			if normalized := normalizePageContext(value, 0); normalized != nil {
				out[key] = normalized
			}
		}
	}

	if len(out) == 0 {
		return ""
	}

	payload, err := json.Marshal(out)
	if err != nil {
		return ""
	}
	if len(payload) > maxCacheFreshnessLen {
		payload = payload[:maxCacheFreshnessLen]
	}
	return sanitizeUserInput(strings.TrimSpace(string(payload)), maxCacheFreshnessLen)
}
