package snips

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	_ "golang.org/x/image/webp"
)

type ImageInput struct {
	Name    string `json:"name"`
	DataURL string `json:"data_url"`
}
type imageKey struct{}

func ValidateImages(images []ImageInput) error {
	if len(images) > 4 {
		return errors.New("at most four images are allowed")
	}
	total := 0
	for _, input := range images {
		mime, data, err := DecodeImage(input)
		if err != nil || mime == "" {
			return errors.New("invalid image")
		}
		total += len(data)
		if total > 2<<20 || len(input.Name) > 200 {
			return errors.New("image attachment limit exceeded")
		}
		cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
		expected := strings.TrimPrefix(mime, "image/")
		if expected == "jpeg" {
			expected = "jpeg"
		}
		if err != nil || format != expected || cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > 4096 || cfg.Height > 4096 || cfg.Width*cfg.Height > 16000000 {
			return errors.New("invalid image dimensions or format")
		}
	}
	return nil
}
func DecodeImage(input ImageInput) (string, []byte, error) {
	prefix, encoded, ok := strings.Cut(input.DataURL, ",")
	mime := strings.TrimSuffix(strings.TrimPrefix(prefix, "data:"), ";base64")
	if !ok || (mime != "image/png" && mime != "image/jpeg" && mime != "image/webp") || prefix != "data:"+mime+";base64" || len(encoded) > 3<<20 {
		return "", nil, errors.New("unsupported image")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	return mime, data, err
}
func WithImages(ctx context.Context, images []ImageInput) context.Context {
	return context.WithValue(ctx, imageKey{}, images)
}
func imagesFromContext(ctx context.Context) []ImageInput {
	images, _ := ctx.Value(imageKey{}).([]ImageInput)
	return images
}
func providerSupportsInput(ctx context.Context, p Provider) bool {
	if len(imagesFromContext(ctx)) == 0 || p.Name() == "gemini" {
		return true
	}
	provider, ok := p.(*OpenAIProvider)
	return ok && (provider.name == "deepseek" || provider.name == "opencode") && strings.HasPrefix(strings.ToLower(provider.model), "deepseek-v4.") && strings.Contains(strings.ToLower(provider.model), "flash")
}
