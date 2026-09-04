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
	"image/heic":      true,
	"image/heif":      true,
	"image/jpeg":      true,
	"image/jxl":       true,
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
	case "heic":
		return ".heic"
	case "heif":
		return ".heif"
	case "jxl", "jpegxl":
		return ".jxl"
	case "tiff", "tif":
		return ".tiff"
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
	case "heic":
		return "image/heic"
	case "heif":
		return "image/heif"
	case "jxl", "jpegxl":
		return "image/jxl"
	case "tiff", "tif":
		return "image/tiff"
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
	case "image/heic":
		return "heic"
	case "image/heif":
		return "heif"
	case "image/jxl":
		return "jxl"
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
				if q, ok := autoQuality(tok[2:]); ok {
					quality = q
				} else if q, err := strconv.Atoi(tok[2:]); err == nil {
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

const (
	maxTransformDimension = 8000
	maxTransformScale     = 4.0
	// maxTransformBorder caps border/pad growth per step so bo_ tokens can
	// never inflate the canvas past the pixel-area budget.
	maxTransformBorder = 2000
	maxBlurStrength    = 2000.0
	maxMedianWindow    = 31
)

// allowedOutputFormat reports whether fmt is an encoder the CDN serves.
func allowedOutputFormat(format string) bool {
	switch strings.ToLower(format) {
	case "webp", "jpeg", "jpg", "png", "avif", "gif", "tiff", "tif":
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
	if p.Width < 0 || p.Width > 8000 {
		return fmt.Errorf("width must be between 0 and 8000")
	}
	if p.Height < 0 || p.Height > 8000 {
		return fmt.Errorf("height must be between 0 and 8000")
	}
	if p.Width > 0 && p.Height > 0 && int64(p.Width)*int64(p.Height) > maxPixelArea {
		return fmt.Errorf("requested dimensions exceed the %d pixel limit", maxPixelArea)
	}

	// Chain mode: geometry/effects live in the chain string; only the output
	// format and quality are validated here (the engine validates the rest).
	if p.Chain != "" {
		if len(p.Chain) > 2048 {
			return fmt.Errorf("transformation chain is too long")
		}
		if strings.ContainsAny(p.Chain, "\x00\r\n") {
			return fmt.Errorf("transformation chain contains invalid characters")
		}
		if p.Quality < 1 || p.Quality > 100 {
			return fmt.Errorf("quality must be between 1 and 100")
		}
		if err := validateTransformationChain(p.Chain); err != nil {
			return err
		}
		if !allowedOutputFormat(p.Format) {
			return fmt.Errorf("unsupported format: %s", p.Format)
		}
		return nil
	}
	if p.Width == 0 && p.Height == 0 {
		return fmt.Errorf("at least one of width or height must be specified")
	}
	// Cap the output area independently of the per-dimension limits: 8000x8000
	// is within bounds but would allocate ~256 MB of pixels. 24 MP comfortably
	// covers any real display while bounding memory.
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

func validateTransformationChain(chain string) error {
	maxWidth, maxHeight := 0, 0
	borderTotal := 0
	zScale := 1.0
	for _, step := range strings.Split(chain, "/") {
		for _, token := range strings.Split(step, ",") {
			parts := strings.SplitN(token, "_", 2)
			if len(parts) != 2 {
				continue
			}
			key, value := strings.ToLower(parts[0]), parts[1]
			switch key {
			case "w", "h":
				n, err := strconv.Atoi(value)
				if err != nil || n < 0 || n > maxTransformDimension {
					return fmt.Errorf("chain %s must be between 0 and %d", key, maxTransformDimension)
				}
				if key == "w" && n > maxWidth {
					maxWidth = n
				}
				if key == "h" && n > maxHeight {
					maxHeight = n
				}
			case "z":
				n, err := strconv.ParseFloat(value, 64)
				if err != nil || n <= 0 || n > maxTransformScale {
					return fmt.Errorf("chain scale must be greater than 0 and at most %.0fx", maxTransformScale)
				}
				if n > zScale {
					zScale = n
				}
			case "bo":
				// Border/pad widens the canvas by 2n per axis; without a cap
				// a single bo_40000 token turns any image into a ~gigapixel
				// allocation.
				nStr := strings.SplitN(value, ":", 2)[0]
				n, err := strconv.Atoi(nStr)
				if err != nil || n <= 0 || n > maxTransformBorder {
					return fmt.Errorf("chain border must be between 1 and %d", maxTransformBorder)
				}
				borderTotal += n
			case "e":
				effectParts := strings.SplitN(value, ":", 2)
				if len(effectParts) != 2 {
					continue
				}
				effect := strings.ToLower(effectParts[0])
				switch effect {
				case "blur":
					n, err := strconv.ParseFloat(effectParts[1], 64)
					if err != nil {
						return fmt.Errorf("invalid %s effect value", effect)
					}
					if n < 0 || n > maxBlurStrength {
						return fmt.Errorf("blur strength must be between 0 and %.0f", maxBlurStrength)
					}
				case "median", "oil_paint":
					n, err := strconv.ParseFloat(effectParts[1], 64)
					if err != nil {
						return fmt.Errorf("invalid %s effect value", effect)
					}
					if n < 0 || n > maxMedianWindow {
						return fmt.Errorf("%s window must be between 0 and %d", effect, maxMedianWindow)
					}
				}
			}
		}
	}
	// Account for border growth and scale amplification so the declared
	// w/h alone can't hide a much larger effective canvas.
	if maxWidth > 0 && maxHeight > 0 {
		effW := maxWidth + 2*borderTotal
		effH := maxHeight + 2*borderTotal
		area := float64(int64(effW)*int64(effH)) * zScale * zScale
		if area > float64(maxPixelArea) {
			return fmt.Errorf("chain dimensions exceed the %d pixel limit", maxPixelArea)
		}
	} else if borderTotal > 0 && zScale > 1 {
		return fmt.Errorf("chain dimensions exceed the %d pixel limit", maxPixelArea)
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
