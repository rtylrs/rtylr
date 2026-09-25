package cdn

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/gif" // register GIF decoder
	"image/jpeg"
	_ "image/png" // register PNG decoder
	"io"
	"os"
	"strings"

	nude "github.com/koyachi/go-nude"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	_ "golang.org/x/image/webp" // register WebP decoder

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

// ErrImageBlocked is returned by uploadFile when the moderation provider
// flags an image. The upload handler maps it to 422 so clients can tell a
// policy rejection apart from a malformed request.
var ErrImageBlocked = errors.New("image rejected by content moderation")

// moderationMaxBytes bounds how large an image we send to a moderation
// provider. Larger images skip the check (fail-open) rather than shipping
// tens of megabytes to an external API per upload.
const (
	moderationMaxBytes        = 8 << 20 // 8 MiB
	moderationMaxPixels int64 = 24_000_000
)

// ImageModerator decides whether an uploaded image is allowed. Implementations
// must be safe for concurrent use.
type ImageModerator interface {
	// Check returns blocked=true when the image violates content policy.
	// reason is a short label for logs (never shown verbatim to end users).
	Check(ctx context.Context, mimeType string, data []byte) (blocked bool, reason string, err error)
}

// NewModeratorFromConfig builds the configured moderator, or nil when
// moderation is disabled. The second return value is the fail-closed flag.
func NewModeratorFromConfig(cfg *config.AppConfig) (ImageModerator, bool) {
	if cfg == nil || !cfg.CDNModerationEnabled {
		return nil, false
	}

	provider := cfg.CDNModerationProvider
	if provider == "openai" && !config.IsOpenAIConfigured() {
		logger.Global().Warn("cdn: moderation provider openai selected but OPENAI_API_KEY is unset; falling back to local heuristic")
		provider = "local"
	}

	switch provider {
	case "openai":
		return newOpenAIModerator(config.GetOpenAIAPIKey(), config.GetOpenAIBaseURL()), cfg.CDNModerationFailClosed
	case "local":
		return &localNudeModerator{}, cfg.CDNModerationFailClosed
	default:
		logger.Global().Warn("cdn: unknown moderation provider; moderation disabled",
			logger.String("provider", provider))
		return nil, false
	}
}

// openAIModerator sends the image to OpenAI's moderation endpoint
// (omni-moderation-latest), which is free to use. Only the sexual-content
// categories block an upload — the endpoint also scores violence etc., but
// this filter is scoped to nudity.
type openAIModerator struct {
	client openai.Client
}

func newOpenAIModerator(apiKey, baseURL string) *openAIModerator {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	return &openAIModerator{client: openai.NewClient(opts...)}
}

func (m *openAIModerator) Check(ctx context.Context, mimeType string, data []byte) (bool, string, error) {
	dataURL := "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data)

	res, err := m.client.Moderations.New(ctx, openai.ModerationNewParams{
		Model: openai.ModerationModelOmniModerationLatest,
		Input: openai.ModerationNewParamsInputUnion{
			OfModerationMultiModalArray: []openai.ModerationMultiModalInputUnionParam{
				openai.ModerationMultiModalInputParamOfImageURL(openai.ModerationImageURLInputImageURLParam{
					URL: dataURL,
				}),
			},
		},
	})
	if err != nil {
		return false, "", fmt.Errorf("openai moderation: %w", err)
	}
	if len(res.Results) == 0 {
		return false, "", errors.New("openai moderation: empty result")
	}

	r := res.Results[0]
	switch {
	case r.Categories.SexualMinors:
		return true, "sexual/minors", nil
	case r.Categories.Sexual:
		return true, "sexual", nil
	default:
		return false, "", nil
	}
}

// localNudeModerator is a zero-cost in-process fallback based on skin-tone
// region detection (koyachi/go-nude). It is a heuristic: expect false
// positives on portraits and false negatives on real nudity. Prefer the
// openai provider when an API key is available.
type localNudeModerator struct{}

func (m *localNudeModerator) Check(_ context.Context, _ string, data []byte) (bool, string, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return false, "", fmt.Errorf("reading image metadata for moderation: %w", err)
	}
	if config.Width <= 0 || config.Height <= 0 || int64(config.Width)*int64(config.Height) > moderationMaxPixels {
		return false, "", fmt.Errorf("image dimensions exceed moderation limit")
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return false, "", fmt.Errorf("decoding image for moderation: %w", err)
	}

	// go-nude only reads from a file path and picks its decoder by extension,
	// so re-encode whatever we decoded (incl. webp) to a temp JPEG.
	tmp, err := os.CreateTemp("", "cdn-moderation-*.jpg")
	if err != nil {
		return false, "", fmt.Errorf("creating moderation temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		if removeErr := os.Remove(tmpPath); removeErr != nil {
			logger.Global().Warn("cdn: failed to remove moderation temp file", logger.Error(removeErr))
		}
	}()

	encodeErr := jpeg.Encode(tmp, img, &jpeg.Options{Quality: 85})
	closeErr := tmp.Close()
	if encodeErr != nil {
		return false, "", fmt.Errorf("encoding image for moderation: %w", encodeErr)
	}
	if closeErr != nil {
		return false, "", fmt.Errorf("closing moderation temp file: %w", closeErr)
	}

	isNude, err := nude.IsNude(tmpPath)
	if err != nil {
		return false, "", fmt.Errorf("nudity heuristic: %w", err)
	}
	if isNude {
		return true, "nudity-heuristic", nil
	}
	return false, "", nil
}

// moderateImage runs the configured moderator over an image upload. It returns
// the (possibly re-buffered) body reader, since checking requires consuming the
// stream. A nil moderator or non-image MIME passes through untouched.
func (c *CDN) moderateImage(ctx context.Context, mimeType string, size int64, body io.Reader) (io.Reader, error) {
	if c.moderator == nil || !strings.HasPrefix(mimeType, "image/") {
		return body, nil
	}
	if size > moderationMaxBytes {
		logger.Ctx(ctx).Warn("cdn: image exceeds moderation size cap; skipping check",
			logger.String("mime_type", mimeType))
		return body, nil
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return body, fmt.Errorf("buffering image for moderation: %w", err)
	}

	blocked, reason, err := c.moderator.Check(ctx, mimeType, data)
	if err != nil {
		if c.moderationFailClosed {
			logger.Ctx(ctx).Error("cdn: moderation check failed; rejecting upload (fail-closed)", logger.Error(err))
			return bytes.NewReader(data), ErrImageBlocked
		}
		logger.Ctx(ctx).Warn("cdn: moderation check failed; allowing upload (fail-open)", logger.Error(err))
		return bytes.NewReader(data), nil
	}
	if blocked {
		logger.Ctx(ctx).Warn("cdn: image blocked by content moderation",
			logger.String("reason", reason),
			logger.String("mime_type", mimeType))
		return bytes.NewReader(data), ErrImageBlocked
	}

	return bytes.NewReader(data), nil
}
