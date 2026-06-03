package cdn

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
)

// allowedUploadMIMETypes is the whitelist for the upload endpoint. Anything
// outside this set is rejected so the bucket can't double as free file hosting.
var allowedUploadMIMETypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"image/gif":       true,
	"image/avif":      true,
	"image/tiff":      true,
	"video/mp4":       true,
	"video/webm":      true,
	"video/quicktime": true,
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
	return fmt.Sprintf("%x", h.Sum(nil))[:32]
}

// Params holds the parameters for a single image transformation request.
type Params struct {
	Width   int
	Height  int
	Fit     string
	Format  string
	Quality int
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

// Validate returns an error if parameters are out of acceptable range.
func (p *Params) Validate() error {
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
	allowed := map[string]bool{"webp": true, "jpeg": true, "jpg": true, "png": true, "avif": true}
	if !allowed[strings.ToLower(p.Format)] {
		return fmt.Errorf("unsupported format: %s", p.Format)
	}
	allowedFit := map[string]bool{"cover": true, "contain": true, "fill": true, "inside": true, "outside": true}
	if !allowedFit[strings.ToLower(p.Fit)] {
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
	)
}

// StorageKey returns the S3 object key for the transformed asset.
func (p *Params) StorageKey(fileUUID, key string) string {
	return fmt.Sprintf("transformed/%s/%s%s", fileUUID, key, extensionForFormat(p.Format))
}

// CLIArgs returns the command-line arguments for invoking the media-engine.
func (p *Params) CLIArgs(inputPath, outputPath string) []string {
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
