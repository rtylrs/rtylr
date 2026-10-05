package application

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
)

const maxCacheFreshnessLen = 1600

var cacheFreshnessKeys = []string{"dataFreshness", "data_freshness", "dataVersion", "dataVersions", "cacheContext", "cache_context", "freshness"}

// BuildCacheFreshness extracts only page fields that can change the truth of
// an exact response. It intentionally excludes the full rolling conversation.
func BuildCacheFreshness(pageContext map[string]any) string {
	if len(pageContext) == 0 {
		return ""
	}
	out := make(map[string]any)
	for _, key := range cacheFreshnessKeys {
		if value, ok := pageContext[key]; ok {
			if normalized := NormalizePageContext(value, 0); normalized != nil {
				out[key] = normalized
			}
		}
	}
	for _, key := range []string{"module", "route", "selected", "filters", "visibleRows"} {
		if value, ok := pageContext[key]; ok {
			if normalized := NormalizePageContext(value, 0); normalized != nil {
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
	return SanitizeUserInput(strings.TrimSpace(string(payload)), maxCacheFreshnessLen)
}

func BuildGroundingCacheFreshness(grounding *AnalystGroundingResult) string {
	if grounding == nil {
		return ""
	}
	parts := []string{"tool=" + groundingToolName(grounding), "rows=" + shortCacheHash(grounding.RowsJSON), "spec=" + shortCacheHash(grounding.SpecJSON), "count=" + strconv.Itoa(grounding.RowCount), "source=" + shortCacheHash(grounding.EvidenceSource), "plan=" + shortCacheHash(grounding.RequestPlan)}
	if grounding.Truncated {
		parts = append(parts, "truncated=true")
	}
	return strings.Join(parts, ";")
}

func CombineCacheFreshness(pageFreshness, groundingFreshness string) string {
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

func groundingToolName(grounding *AnalystGroundingResult) string {
	if grounding == nil || strings.TrimSpace(grounding.ToolName) == "" {
		return "analytics.query"
	}
	return strings.TrimSpace(grounding.ToolName)
}

func shortCacheHash(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])[:16]
}
