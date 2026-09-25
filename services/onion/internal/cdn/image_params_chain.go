package cdn

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func classChain(query url.Values) (string, error) {
	raw := strings.TrimSpace(query.Get("class"))
	if raw == "" {
		raw = strings.TrimSpace(query.Get("preset"))
	}
	if raw == "" {
		return "", nil
	}
	key := strings.ToLower(raw)
	if chain, ok := imageTransformClasses[key]; ok {
		return chain, nil
	}
	return "", fmt.Errorf("unknown image class: %s", raw)
}

func joinChain(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, "/ ")
		if part != "" {
			out = append(out, part)
		}
	}
	return strings.Join(out, "/")
}

func hasImageTransformQuery(query url.Values) bool {
	for key := range query {
		if imageTransformQueryKeys[strings.ToLower(key)] {
			return true
		}
	}
	return false
}

func requestedImageFormat(query url.Values, accept string, negotiateDefault bool) (string, bool, error) {
	raw := strings.TrimSpace(query.Get("format"))
	if raw == "" {
		raw = strings.TrimSpace(query.Get("f"))
	}
	if raw == "" {
		raw = formatFromAutoParam(query.Get("auto"))
	}

	if raw == "" && negotiateDefault {
		if negotiated := negotiateFormat(accept); negotiated != "" {
			return negotiated, true, nil
		}
	}
	if raw == "" {
		return "webp", false, nil
	}

	format, varyAccept, err := normalizeImageFormat(raw, accept)
	if err != nil {
		return "", false, err
	}
	return format, varyAccept, nil
}

func formatFromAutoParam(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" || raw == "false" || raw == "0" || raw == "off" {
		return ""
	}
	if raw == "true" || raw == "1" || raw == "on" {
		return "auto"
	}
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' '
	}) {
		switch strings.TrimSpace(part) {
		case "avif", "webp":
			return "auto"
		}
	}
	return ""
}

func normalizeImageFormat(raw, accept string) (string, bool, error) {
	format := strings.ToLower(strings.TrimSpace(raw))
	switch format {
	case "", "source", "original":
		return "", false, nil
	case "auto":
		if negotiated := negotiateFormat(accept); negotiated != "" {
			return negotiated, true, nil
		}
		return "webp", true, nil
	case "jpg", "jpeg", "bjpg", "bjpeg", "pjpg", "pjpeg":
		return "jpeg", false, nil
	case "png", "png8":
		return "png", false, nil
	case "webp", "webply", "webpll":
		return "webp", false, nil
	case "avif", "gif", "tiff", "tif":
		return format, false, nil
	default:
		return "", false, fmt.Errorf("unsupported format: %s", raw)
	}
}

func queryQuality(query url.Values) (int, error) {
	raw := strings.TrimSpace(query.Get("quality"))
	if raw == "" {
		raw = strings.TrimSpace(query.Get("q"))
	}
	if raw == "" {
		return uploadURLWebPQuality, nil
	}
	if q, ok := autoQuality(raw); ok {
		return q, nil
	}
	q, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid quality: %s", raw)
	}
	if q < 1 || q > 100 {
		return 0, fmt.Errorf("quality must be between 1 and 100")
	}
	return q, nil
}

func autoQuality(raw string) (int, bool) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	switch raw {
	case "auto", "q_auto", "auto:good", "q_auto:good":
		return uploadURLWebPQuality, true
	case "auto:best", "q_auto:best":
		return 92, true
	case "auto:eco", "q_auto:eco":
		return 75, true
	case "auto:low", "q_auto:low":
		return 60, true
	default:
		return 0, false
	}
}

func queryPositiveInt(query url.Values, keys ...string) (int, error) {
	for _, key := range keys {
		raw := strings.TrimSpace(query.Get(key))
		if raw == "" {
			continue
		}
		n, err := strconv.Atoi(raw)
		if err != nil {
			return 0, fmt.Errorf("invalid %s: %s", key, raw)
		}
		if n < 0 {
			return 0, fmt.Errorf("%s must not be negative", key)
		}
		return n, nil
	}
	return 0, nil
}

func queryDPR(query url.Values) (float64, error) {
	raw := strings.TrimSpace(query.Get("dpr"))
	if raw == "" {
		return 0, nil
	}
	if strings.EqualFold(raw, "auto") {
		return 2, nil
	}
	dpr, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid dpr: %s", raw)
	}
	if dpr < 1 || dpr > 10 {
		return 0, fmt.Errorf("dpr must be between 1 and 10")
	}
	return dpr, nil
}

func applyAspectRatio(query url.Values, width, height int) (int, int, error) {
	raw := firstQuery(query, "aspect-ratio", "aspect", "ar")
	if raw == "" {
		return width, height, nil
	}
	if width <= 0 && height <= 0 {
		return width, height, fmt.Errorf("aspect ratio requires width or height")
	}
	w, h, err := ratioDimensions(raw, width, height)
	if err != nil {
		return width, height, err
	}
	if width <= 0 {
		width = w
	}
	if height <= 0 {
		height = h
	}
	return width, height, nil
}
