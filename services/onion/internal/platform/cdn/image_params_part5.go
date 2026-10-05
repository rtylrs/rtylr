package cdn

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func appendFormatStep(chain, format string) string {
	if chainFmt, _ := parseChainParams(chain); chainFmt != "" {
		return chain
	}
	step := "f_" + strings.ToLower(format)
	if strings.TrimSpace(chain) == "" {
		return step
	}
	return chain + "/" + step
}

func canonicalImageFit(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "bounds":
		return "bounds"
	case "crop":
		return "crop"
	case "cover":
		return "cover"
	case "contain", "inside":
		return "contain"
	case "fill":
		return "fill"
	case "limit", "outside":
		return "limit"
	case "pad":
		return "pad"
	case "scale":
		return "scale"
	default:
		return "cover"
	}
}

func splitParamList(raw string) []string {
	values := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';'
	})
	out := make([]string, 0, len(values))
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func sanitizeChainValue(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "#")
	raw = strings.ReplaceAll(raw, " ", "_")
	raw = strings.ReplaceAll(raw, "/", "")
	raw = strings.ReplaceAll(raw, "\x00", "")
	raw = strings.ReplaceAll(raw, "\r", "")
	raw = strings.ReplaceAll(raw, "\n", "")
	if raw == "" {
		return raw
	}
	return raw
}

func sanitizeEffectArgs(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, ",", ":")
	raw = strings.ReplaceAll(raw, "#", "")
	raw = strings.ReplaceAll(raw, " ", "_")
	raw = strings.ReplaceAll(raw, "/", "")
	raw = strings.ReplaceAll(raw, "\x00", "")
	raw = strings.ReplaceAll(raw, "\r", "")
	raw = strings.ReplaceAll(raw, "\n", "")
	return raw
}

func firstQuery(query url.Values, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(query.Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func effectStep(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "e_") {
		return sanitizeChainValue(raw)
	}
	return "e_" + sanitizeEffectArgs(raw)
}

func textStep(query url.Values) string {
	text := firstQuery(query, "text", "txt", "caption")
	if text == "" {
		return ""
	}
	return textOverlayStep(text, firstQuery(query, "font", "tf"), query.Get("tc"), query.Get("tw"), query.Get("tg"), query.Get("tm"), firstQuery(query, "text-opacity", "to"))
}

func watermarkStep(query url.Values) (string, error) {
	text := firstQuery(query, "wm-text", "watermark", "wm")
	if text == "" {
		return "", nil
	}
	if looksLikeRemoteOverlay(text) {
		return "", fmt.Errorf("remote image watermarks are not supported")
	}
	return textOverlayStep(
		text,
		firstQuery(query, "wm-font", "font", "tf"),
		firstQuery(query, "wm-color", "tc"),
		query.Get("wm-width"),
		firstQuery(query, "wm-gravity", "tg"),
		firstQuery(query, "wm-margin", "tm"),
		query.Get("wm-opacity"),
	), nil
}

func textOverlayStep(text, font, color, maxWidth, gravity, margin, opacity string) string {
	parts := []string{"txt_" + sanitizeTextToken(text)}
	if font := sanitizeTextToken(font); font != "" {
		parts = append(parts, "font_"+font)
	}
	if color := sanitizeChainValue(color); color != "" {
		parts = append(parts, "tc_"+color)
	}
	if maxWidth := sanitizeChainValue(maxWidth); maxWidth != "" {
		parts = append(parts, "tw_"+maxWidth)
	}
	if gravity := sanitizeChainValue(gravity); gravity != "" {
		parts = append(parts, "tg_"+gravity)
	}
	if margin := sanitizeChainValue(margin); margin != "" {
		parts = append(parts, "tm_"+margin)
	}
	if opacity := sanitizeChainValue(opacity); opacity != "" {
		parts = append(parts, "to_"+opacity)
	}
	return strings.Join(parts, ",")
}

func looksLikeRemoteOverlay(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	return strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "//")
}

func sanitizeTextToken(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, ",", " ")
	raw = strings.ReplaceAll(raw, "/", " ")
	raw = strings.ReplaceAll(raw, "\x00", "")
	raw = strings.ReplaceAll(raw, "\r", " ")
	raw = strings.ReplaceAll(raw, "\n", " ")
	raw = strings.Join(strings.Fields(raw), "_")
	if len(raw) > 200 {
		raw = raw[:200]
	}
	return raw
}

func nonEmptySteps(steps []string) []string {
	out := steps[:0]
	for _, step := range steps {
		if strings.TrimSpace(step) != "" {
			out = append(out, step)
		}
	}
	return out
}

func isFalsey(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "0", "false", "off", "none", "no":
		return true
	default:
		return false
	}
}

func imageVariantSurrogateKey(p Params) string {
	return "cdn-variant-" + p.CacheKey("variant")
}

func setImageVariantHeaders(w http.ResponseWriter, p Params, varyAccept bool, tags ...string) {
	if varyAccept {
		w.Header().Set("Vary", "Accept")
	}
	if len(tags) > 0 {
		w.Header().Set("Surrogate-Key", strings.Join(tags, " "))
	}
	w.Header().Set("Voxire-CDN-Transform", p.Chain)
}

func queryInt(query url.Values, key string) int {
	v := query.Get(key)
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}
