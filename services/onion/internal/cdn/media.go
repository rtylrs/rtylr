package cdn

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// allowedUploadMIMETypes is the whitelist for the upload endpoint. Anything
// outside this set is rejected so the bucket can't double as free file hosting.
var allowedUploadMIMETypes = map[string]bool{
	"application/msword":       true,
	"application/pdf":          true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"image/avif":      true,
	"image/gif":       true,
	"image/jpeg":      true,
	"image/png":       true,
	"image/tiff":      true,
	"image/webp":      true,
	"text/csv":        true,
	"text/plain":      true,
	"video/mp4":       true,
	"video/quicktime": true,
	"video/webm":      true,
}

func isAllowedMIMEType(mime string) bool {
	return allowedUploadMIMETypes[strings.ToLower(mime)]
}

// extensionForFormat returns the file extension for an image/video format name.
func extensionForFormat(format string) string {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return ".jpg"
	case "png":
		return ".png"
	case "webp":
		return ".webp"
	case "avif":
		return ".avif"
	case "gif":
		return ".gif"
	default:
		return "." + format
	}
}

// mimeForFormat returns the MIME type for an image format name.
func mimeForFormat(format string) string {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	case "avif":
		return "image/avif"
	case "gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

// negotiateFormat picks an output format from a request Accept header,
// preferring the most modern codec the client advertises (AVIF, then WebP).
// It returns "" when the header offers nothing better, leaving the caller's
// default in place. An explicit ?format= query param should take precedence
// over this.
func negotiateFormat(accept string) string {
	a := strings.ToLower(accept)
	switch {
	case strings.Contains(a, "image/avif"):
		return "avif"
	case strings.Contains(a, "image/webp"):
		return "webp"
	default:
		return ""
	}
}

// mimeTypeToFormat maps an upload MIME type to the short format name used when
// building the original object's storage key.
func mimeTypeToFormat(mime string) string {
	switch mime {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	case "image/avif":
		return "avif"
	case "image/tiff":
		return "tiff"
	case "video/mp4":
		return "mp4"
	case "video/webm":
		return "webm"
	case "video/quicktime":
		return "mov"
	case "application/pdf":
		return "pdf"
	case "application/msword":
		return "doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "docx"
	case "application/vnd.ms-excel":
		return "xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return "xlsx"
	case "text/csv":
		return "csv"
	case "text/plain":
		return "txt"
	default:
		return "bin"
	}
}

// cacheKey generates a deterministic, collision-resistant key from its parts.
// A null separator between parts prevents ambiguous concatenation (e.g.
// "1"+"23" vs "12"+"3").
func cacheKey(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum[:16])
}

// Params holds the parameters for a single image transformation request.
type Params struct {
	Width   int
	Height  int
	Fit     string
	Format  string
	Quality int
	// Chain, when set, is a Cloudinary-style transformation chain (e.g.
	// "w_400,c_fill,g_auto/e_blur:300") applied via the media-engine's transform
	// subcommand instead of a plain resize. Width/Height/Fit are ignored.
	Chain string
}

// parseChainParams extracts the output format (f_) and quality (q_) directives
// from a transformation chain so the CDN can set the right Content-Type and
// storage extension. The last occurrence wins.
func parseChainParams(chain string) (format string, quality int) {
	for _, step := range strings.Split(chain, "/") {
		for _, tok := range strings.Split(step, ",") {
			switch {
			case strings.HasPrefix(tok, "f_"):
				format = tok[2:]
			case strings.HasPrefix(tok, "q_"):
				if q, err := strconv.Atoi(tok[2:]); err == nil {
					quality = q
				}
			}
		}
	}
	return format, quality
}

// Defaults fills zero-value fields with sensible production defaults.
func (p *Params) Defaults() {
	if p.Quality == 0 {
		p.Quality = 85
	}
	if p.Format == "" {
		p.Format = "webp"
	}
	if p.Fit == "" {
		p.Fit = "cover"
	}
}

// maxPixelArea bounds the output width*height to limit per-request memory.
const maxPixelArea int64 = 24_000_000

// allowedOutputFormat reports whether fmt is an encoder the CDN serves.
func allowedOutputFormat(format string) bool {
	switch strings.ToLower(format) {
	case "webp", "jpeg", "jpg", "png", "avif":
		return true
	default:
		return false
	}
}

func allowedFitMode(fit string) bool {
	switch strings.ToLower(fit) {
	case "cover", "contain", "fill", "inside", "outside":
		return true
	default:
		return false
	}
}

// Validate returns an error if parameters are out of acceptable range.
func (p *Params) Validate() error {
	// Chain mode: geometry/effects live in the chain string; only the output
	// format and quality are validated here (the engine validates the rest).
	if p.Chain != "" {
		if p.Quality < 1 || p.Quality > 100 {
			return fmt.Errorf("quality must be between 1 and 100")
		}
		if !allowedOutputFormat(p.Format) {
			return fmt.Errorf("unsupported format: %s", p.Format)
		}
		return nil
	}
	if p.Width < 0 || p.Width > 8000 {
		return fmt.Errorf("width must be between 0 and 8000")
	}
	if p.Height < 0 || p.Height > 8000 {
		return fmt.Errorf("height must be between 0 and 8000")
	}
	if p.Width == 0 && p.Height == 0 {
		return fmt.Errorf("at least one of width or height must be specified")
	}
	// Cap the output area independently of the per-dimension limits: 8000x8000
	// is within bounds but would allocate ~256 MB of pixels. 24 MP comfortably
	// covers any real display while bounding memory.
	if p.Width > 0 && p.Height > 0 && int64(p.Width)*int64(p.Height) > maxPixelArea {
		return fmt.Errorf("requested dimensions exceed the %d pixel limit", maxPixelArea)
	}
	if p.Quality < 1 || p.Quality > 100 {
		return fmt.Errorf("quality must be between 1 and 100")
	}
	if !allowedOutputFormat(p.Format) {
		return fmt.Errorf("unsupported format: %s", p.Format)
	}
	if !allowedFitMode(p.Fit) {
		return fmt.Errorf("unsupported fit mode: %s", p.Fit)
	}
	return nil
}

// CacheKey returns a deterministic hash key for this transformation of a file.
func (p *Params) CacheKey(fileUUID string) string {
	return cacheKey(
		fileUUID,
		strconv.Itoa(p.Width),
		strconv.Itoa(p.Height),
		p.Fit,
		strings.ToLower(p.Format),
		strconv.Itoa(p.Quality),
		p.Chain,
	)
}

// StorageKey returns the S3 object key for the transformed asset.
func (p *Params) StorageKey(fileUUID, key string) string {
	return fmt.Sprintf("transformed/%s/%s%s", fileUUID, key, extensionForFormat(p.Format))
}

// CLIArgs returns the command-line arguments for invoking the media-engine.
func (p *Params) CLIArgs(inputPath, outputPath string) []string {
	if p.Chain != "" {
		// transform subcommand applies the full chain; --format/--quality pin
		// the output so it matches the Content-Type/extension the CDN serves.
		return []string{
			"transform",
			"--input", inputPath,
			"--output", outputPath,
			"--chain", p.Chain,
			"--format", strings.ToLower(p.Format),
			"--quality", strconv.Itoa(p.Quality),
		}
	}
	args := []string{
		"resize",
		"--input", inputPath,
		"--output", outputPath,
		"--format", strings.ToLower(p.Format),
		"--quality", strconv.Itoa(p.Quality),
		"--fit", strings.ToLower(p.Fit),
	}
	if p.Width > 0 {
		args = append(args, "--width", strconv.Itoa(p.Width))
	}
	if p.Height > 0 {
		args = append(args, "--height", strconv.Itoa(p.Height))
	}
	return args
}
