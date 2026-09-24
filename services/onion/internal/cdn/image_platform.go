package cdn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/voxire/onion/internal/platform/logger"
)

var defaultImageVariantClasses = []string{
	"thumb",
	"card",
	"gallery",
	"hero",
	"og",
	"product",
	"product-thumb",
	"pos",
	"pos-thumb",
	"lqip",
}

type imagePlanResponse struct {
	ID                string   `json:"id,omitempty"`
	SourceURL         string   `json:"source_url,omitempty"`
	URL               string   `json:"url"`
	Chain             string   `json:"chain,omitempty"`
	Format            string   `json:"format"`
	Quality           int      `json:"quality,omitempty"`
	ContentType       string   `json:"content_type"`
	CacheControl      string   `json:"cache_control"`
	Vary              []string `json:"vary,omitempty"`
	SurrogateKeys     []string `json:"surrogate_keys,omitempty"`
	TransformExplicit bool     `json:"transform_explicit"`
	Original          bool     `json:"original"`
}

type imageVariantPlan struct {
	Class       string `json:"class"`
	URL         string `json:"url"`
	Chain       string `json:"chain"`
	Format      string `json:"format"`
	Quality     int    `json:"quality"`
	ContentType string `json:"content_type"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

type imageVariantsResponse struct {
	ID        string                      `json:"id,omitempty"`
	SourceURL string                      `json:"source_url,omitempty"`
	Variants  map[string]imageVariantPlan `json:"variants"`
}

func (c *CDN) imagePlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "invalid image id", http.StatusBadRequest)
		return
	}
	meta, err := c.fileByUUID(r.Context(), id)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	resp, err := c.planUploadedImage(id, meta.MIMEType, r.URL.Query(), r.Header.Get("Accept"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, r, resp, immutableCacheControl, len(resp.Vary) > 0)
}

func (c *CDN) uploadPlan(w http.ResponseWriter, r *http.Request) {
	if id := strings.TrimSpace(r.URL.Query().Get("id")); id != "" {
		if _, err := uuid.Parse(id); err != nil {
			http.Error(w, "invalid file id", http.StatusBadRequest)
			return
		}
		meta, err := c.fileByUUID(r.Context(), id)
		if err != nil {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		resp, err := c.planUploadedImage(id, meta.MIMEType, r.URL.Query(), r.Header.Get("Accept"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, r, resp, immutableCacheControl, len(resp.Vary) > 0)
		return
	}

	raw := strings.TrimSpace(r.URL.Query().Get("url"))
	source, err := c.resolveUploadURLSource(raw, r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := c.planUploadSource(raw, source, r.URL.Query(), r.Header.Get("Accept"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, r, resp, uploadURLCacheControl, len(resp.Vary) > 0)
}

func (c *CDN) planUploadedImage(id, mimeType string, query url.Values, accept string) (imagePlanResponse, error) {
	opts := imageParamOptions{
		defaultFullImage: true,
		negotiateDefault: true,
	}
	parsed, err := parseImageParams(query, accept, opts)
	if err != nil {
		return imagePlanResponse{}, err
	}
	if parsed.original {
		return imagePlanResponse{
			ID:           id,
			URL:          c.fileURL(id),
			Format:       mimeTypeToFormat(mimeType),
			ContentType:  mimeType,
			CacheControl: immutableCacheControl,
			Original:     true,
		}, nil
	}

	// uploadImageReadURL historically defaults to WebP when no format is
	// present. A plan has already negotiated its default from Accept, so pin
	// that result into the URL and describe the exact URL we return.
	readQuery := imageTransformQuery(query)
	if readQuery.Get("format") == "" && readQuery.Get("f") == "" && readQuery.Get("auto") == "" {
		readQuery.Set("format", parsed.params.Format)
	}
	served, err := parseImageParams(readQuery, accept, opts)
	if err != nil {
		return imagePlanResponse{}, err
	}
	p := served.params
	return imagePlanResponse{
		ID:                id,
		URL:               c.uploadImageReadURL(id, readQuery),
		Chain:             p.Chain,
		Format:            p.Format,
		Quality:           p.Quality,
		ContentType:       mimeForFormat(p.Format),
		CacheControl:      immutableCacheControl,
		Vary:              varyHeaders(parsed.varyAccept),
		SurrogateKeys:     []string{"cdn-file-" + id, served.surrogateKey},
		TransformExplicit: parsed.explicit,
	}, nil
}

func (c *CDN) planUploadSource(raw string, source uploadURLSource, query url.Values, accept string) (imagePlanResponse, error) {
	parsed, err := parseImageParams(query, accept, imageParamOptions{
		defaultFullImage: true,
	})
	if err != nil {
		return imagePlanResponse{}, err
	}
	if parsed.original {
		return imagePlanResponse{
			SourceURL:    raw,
			URL:          c.uploadSourceReadURL(raw, query),
			Format:       "source",
			ContentType:  "application/octet-stream",
			CacheControl: uploadURLCacheControl,
			Original:     true,
		}, nil
	}
	p := parsed.params
	return imagePlanResponse{
		SourceURL:         raw,
		URL:               c.uploadSourceReadURL(raw, query),
		Chain:             p.Chain,
		Format:            p.Format,
		Quality:           p.Quality,
		ContentType:       mimeForFormat(p.Format),
		CacheControl:      uploadURLCacheControl,
		Vary:              varyHeaders(parsed.varyAccept),
		SurrogateKeys:     []string{uploadURLSurrogateKey(uploadSourceCacheSeed(raw, source)), parsed.surrogateKey},
		TransformExplicit: parsed.explicit,
	}, nil
}

func (c *CDN) imageVariants(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "invalid image id", http.StatusBadRequest)
		return
	}
	if _, err := c.fileByUUID(r.Context(), id); err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	variants, err := c.variantPlansForID(id, requestedVariantClasses(r.URL.Query()), r.URL.Query(), r.Header.Get("Accept"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, r, imageVariantsResponse{ID: id, Variants: variants}, immutableCacheControl, true)
}

func (c *CDN) uploadVariants(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("url"))
	if _, err := c.resolveUploadURLSource(raw, r.Host); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	variants, err := c.variantPlansForUpload(raw, requestedVariantClasses(r.URL.Query()), r.URL.Query(), r.Header.Get("Accept"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, r, imageVariantsResponse{SourceURL: raw, Variants: variants}, uploadURLCacheControl, true)
}

func (c *CDN) variantPlansForID(id string, classes []string, baseQuery url.Values, accept string) (map[string]imageVariantPlan, error) {
	out := make(map[string]imageVariantPlan, len(classes))
	for _, class := range classes {
		query := imageTransformQuery(baseQuery)
		query.Set("class", class)
		parsed, err := parseImageParams(query, accept, imageParamOptions{defaultFullImage: true})
		if err != nil {
			return nil, err
		}
		p := parsed.params
		out[class] = imageVariantPlan{
			Class:       class,
			URL:         c.uploadImageReadURL(id, query),
			Chain:       p.Chain,
			Format:      p.Format,
			Quality:     p.Quality,
			ContentType: mimeForFormat(p.Format),
			Width:       p.Width,
			Height:      p.Height,
		}
	}
	return out, nil
}

func (c *CDN) variantPlansForUpload(raw string, classes []string, baseQuery url.Values, accept string) (map[string]imageVariantPlan, error) {
	out := make(map[string]imageVariantPlan, len(classes))
	for _, class := range classes {
		query := imageTransformQuery(baseQuery)
		query.Set("class", class)
		parsed, err := parseImageParams(query, accept, imageParamOptions{defaultFullImage: true})
		if err != nil {
			return nil, err
		}
		p := parsed.params
		out[class] = imageVariantPlan{
			Class:       class,
			URL:         c.uploadSourceReadURL(raw, query),
			Chain:       p.Chain,
			Format:      p.Format,
			Quality:     p.Quality,
			ContentType: mimeForFormat(p.Format),
			Width:       p.Width,
			Height:      p.Height,
		}
	}
	return out, nil
}

func (c *CDN) uploadResponsive(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("url"))
	if _, err := c.resolveUploadURLSource(raw, r.Host); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		u := c.uploadSourceReadURL(raw, variantQuery)
		variants = append(variants, responsiveVariant{Width: width, URL: u})
		srcset = append(srcset, fmt.Sprintf("%s %dw", u, width))
	}

	writeJSON(w, r, responsiveResponse{
		Widths:   widths,
		Srcset:   strings.Join(srcset, ", "),
		Variants: variants,
	}, uploadURLCacheControl)
}

func responsiveWidthsFromQuery(query url.Values) []int {
	minW := clampInt(queryInt(query, "min"), 16, 8192)
	if minW == 16 && strings.TrimSpace(query.Get("min")) == "" {
		minW = 320
	}
	maxW := clampInt(queryInt(query, "max"), minW, 8192)
	if maxW == minW && strings.TrimSpace(query.Get("max")) == "" {
		maxW = 1600
		if maxW < minW {
			maxW = minW
		}
	}
	count := clampInt(queryInt(query, "count"), 1, 12)
	if count == 1 && strings.TrimSpace(query.Get("count")) == "" {
		count = 5
	}
	return responsiveWidths(minW, maxW, count)
}

func requestedVariantClasses(query url.Values) []string {
	raw := strings.TrimSpace(query.Get("classes"))
	if raw == "" {
		return append([]string(nil), defaultImageVariantClasses...)
	}
	seen := map[string]bool{}
	classes := []string{}
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ' ' || r == ';' }) {
		class := strings.ToLower(strings.TrimSpace(part))
		if class == "" || seen[class] {
			continue
		}
		seen[class] = true
		classes = append(classes, class)
	}
	sort.Strings(classes)
	return classes
}

func (c *CDN) uploadSourceReadURL(raw string, query url.Values) string {
	out := imageTransformQuery(query)
	if !hasImageTransformQuery(out) {
		out.Set("t", "f_webp")
	} else if out.Get("format") == "" && out.Get("f") == "" && out.Get("t") == "" && out.Get("auto") == "" {
		out.Set("format", "webp")
	}
	out.Set("url", raw)
	return fmt.Sprintf("%s/upload?%s", c.baseURL, out.Encode())
}

func uploadSourceCacheSeed(raw string, source uploadURLSource) string {
	switch {
	case source.remoteURL != nil:
		return "remote:" + source.remoteURL.String()
	case source.storageKey != "":
		return "storage:" + source.storageKey
	default:
		return raw
	}
}

func varyHeaders(varyAccept bool) []string {
	if varyAccept {
		return []string{"Accept"}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, r *http.Request, payload any, cacheControl string, varyAccept ...bool) {
	if len(varyAccept) > 0 && varyAccept[0] {
		w.Header().Add("Vary", "Accept")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", cacheControl)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Ctx(r.Context()).Warn("cdn: failed to encode json response", logger.Error(err))
	}
}
