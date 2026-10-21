package cdn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/platform/logger"
)

type imageCapabilityResponse struct {
	Formats      []string          `json:"formats"`
	Inputs       []string          `json:"inputs"`
	Fits         []string          `json:"fits"`
	Classes      map[string]string `json:"classes"`
	Effects      []string          `json:"effects"`
	Quality      []string          `json:"quality"`
	Geometry     []string          `json:"geometry"`
	QueryAliases []string          `json:"query_aliases"`
	Endpoints    []string          `json:"endpoints"`
	Purge        []string          `json:"purge"`
	Analysis     []string          `json:"analysis"`
	Security     []string          `json:"security"`
}

type responsiveVariant struct {
	Width int    `json:"width"`
	URL   string `json:"url"`
}

type responsiveResponse struct {
	ID       string              `json:"id"`
	Widths   []int               `json:"widths"`
	Srcset   string              `json:"srcset"`
	Variants []responsiveVariant `json:"variants"`
}

func (c *CDN) capabilities(w http.ResponseWriter, r *http.Request) {
	resp := imageCapabilityResponse{
		Formats: []string{"auto", "avif", "webp", "jpeg", "png", "gif", "tiff"},
		Inputs:  []string{"avif", "gif", "heic", "heif", "jpeg", "jxl", "png", "tiff", "webp"},
		Fits:    []string{"bounds", "cover", "crop", "contain", "fill", "limit", "pad", "scale"},
		Classes: imageTransformClasses,
		Effects: []string{
			"blur", "sharpen", "grayscale", "brightness", "contrast", "saturation",
			"normalize", "negate", "invert", "sepia", "hue", "gamma", "colorize",
			"tint", "duotone", "pixelate", "posterize", "threshold", "edge",
			"emboss", "vignette", "median", "oil", "trim", "text", "watermark",
			"radius", "opacity", "flip", "flop",
		},
		Quality:  []string{"1-100", "auto", "auto:best", "auto:good", "auto:eco", "auto:low"},
		Geometry: []string{"width", "height", "dpr", "aspect-ratio", "crop", "precrop", "x/y", "gravity", "canvas", "pad", "border"},
		QueryAliases: []string{
			"w/width", "h/height", "q/quality", "f/format", "g/gravity",
			"b/background/bg-color/background-color", "ar/aspect/aspect-ratio",
			"class/preset", "text/txt/caption", "watermark/wm/wm-text", "t",
		},
		Endpoints: []string{
			"GET /upload?url=...",
			"GET /upload?id=...",
			"GET /upload/plan",
			"GET /upload/responsive",
			"GET /upload/variants",
			"GET /images/{id}",
			"GET /images/{id}/plan",
			"GET /images/{id}/responsive",
			"GET /images/{id}/variants",
			"GET /images/{id}/info",
			"GET /images/{id}/analyze",
			"POST /sign",
		},
		Purge:    []string{"PURGE /images/{id}", "PURGE /files/{id}", "PURGE /upload?url=...", "POST /purge"},
		Analysis: []string{"info", "blurhash", "average-color", "phash"},
		Security: []string{"optional expires/sig signed delivery", "private-IP remote fetch blocking", "bounded transform dimensions", "bounded media-engine concurrency"},
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", uploadURLCacheControl)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to encode capabilities", logger.Error(err))
	}
}

func (c *CDN) imageInfo(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	query.Set("metric", "info")
	r.URL.RawQuery = query.Encode()
	c.imageAnalyze(w, r)
}

func (c *CDN) imageAnalyze(w http.ResponseWriter, r *http.Request) {
	scope, ok := c.validateSignedDelivery(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	metric := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("metric")))
	if metric == "" {
		metric = "info"
	}
	switch metric {
	case "info", "blurhash", "average-color", "phash":
	default:
		http.Error(w, "unsupported analysis metric", http.StatusBadRequest)
		return
	}

	meta, err := c.fileForRead(r.Context(), scope, id)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	if !isUploadURLTransformableMIME(meta.MIMEType) {
		http.Error(w, "file is not an image", http.StatusUnsupportedMediaType)
		return
	}

	out, err := c.runImageAnalysis(r, meta.StorageKey, metric)
	if err != nil {
		logger.Ctx(r.Context()).Error("cdn: image analysis failed", logger.String("file_uuid", id), logger.Error(err))
		http.Error(w, "image analysis failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", immutableCacheControl)
	w.Header().Set("ETag", quoteETag(cacheKey("image-analysis", id, metric)))
	w.Header().Set("Surrogate-Key", "cdn-file-"+id)
	if metric == "info" {
		w.Header().Set("Content-Type", "application/json")
		var payload json.RawMessage
		if err := json.Unmarshal(out, &payload); err != nil {
			http.Error(w, "invalid image analysis response", http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			logger.Ctx(r.Context()).Warn("cdn: failed to encode image analysis response", logger.Error(err))
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"id":     id,
		"metric": metric,
		"value":  strings.TrimSpace(string(out)),
	}); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to encode image analysis response", logger.Error(err))
	}
}

func (c *CDN) runImageAnalysis(r *http.Request, storageKey, metric string) ([]byte, error) {
	if c.mediaEngine == "" {
		return nil, fmt.Errorf("media-engine path is not configured")
	}
	tmpIn, _, cleanup, err := c.stageOriginal(r.Context(), storageKey)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	args := []string{"analyze", "--input", tmpIn, "--metric", metric}
	if metric == "blurhash" {
		if x := strings.TrimSpace(r.URL.Query().Get("x")); x != "" {
			args = append(args, "--x", x)
		}
		if y := strings.TrimSpace(r.URL.Query().Get("y")); y != "" {
			args = append(args, "--y", y)
		}
	}

	if c.engineSem != nil {
		select {
		case c.engineSem <- struct{}{}:
			defer func() { <-c.engineSem }()
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
	}
	execCtx, cancel := context.WithTimeout(r.Context(), engineTimeout)
	defer cancel()

	if c.engine != nil {
		if conn, release, acquireErr := c.engine.acquire(execCtx); acquireErr == nil {
			out, runErr := conn.run(execCtx, args)
			release()
			if runErr == nil {
				return bytes.TrimSpace([]byte(out)), nil
			}
		} else if execCtx.Err() != nil {
			return nil, execCtx.Err()
		}
	}

	// #nosec G204 G702 -- mediaEngine is a configured binary path; args are fixed metric names and temp paths.
	// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
	cmd := exec.CommandContext(execCtx, c.mediaEngine, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("media-engine analyze failed: %w\noutput: %s", err, out)
	}
	return bytes.TrimSpace(out), nil
}

func (c *CDN) imageResponsive(w http.ResponseWriter, r *http.Request) {
	scope, ok := c.validateSignedDelivery(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if _, err := c.fileForRead(r.Context(), scope, id); err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	query := r.URL.Query()
	widths := responsiveWidthsFromQuery(query)

	variants := make([]responsiveVariant, 0, len(widths))
	srcset := make([]string, 0, len(widths))
	for _, width := range widths {
		variantQuery := imageTransformQuery(query)
		variantQuery.Set("width", strconv.Itoa(width))
		if variantQuery.Get("format") == "" && variantQuery.Get("f") == "" && variantQuery.Get("auto") == "" {
			variantQuery.Set("format", "auto")
		}
		u := fmt.Sprintf("%s/images/%s?%s", c.baseURL, id, variantQuery.Encode())
		variants = append(variants, responsiveVariant{Width: width, URL: u})
		srcset = append(srcset, fmt.Sprintf("%s %dw", u, width))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", immutableCacheControl)
	if err := json.NewEncoder(w).Encode(responsiveResponse{
		ID:       id,
		Widths:   widths,
		Srcset:   strings.Join(srcset, ", "),
		Variants: variants,
	}); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to encode responsive response", logger.Error(err))
	}
}

func responsiveWidths(minW, maxW, count int) []int {
	if count <= 0 || maxW < minW {
		return nil
	}
	if count == 1 {
		return []int{maxW}
	}
	widths := make([]int, 0, count)
	seen := map[int]bool{}
	step := float64(maxW-minW) / float64(count-1)
	for i := 0; i < count; i++ {
		width := int(float64(minW) + step*float64(i) + 0.5)
		if !seen[width] {
			widths = append(widths, width)
			seen[width] = true
		}
	}
	return widths
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
