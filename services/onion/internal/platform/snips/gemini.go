package snips

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

var (
	ErrGeminiNotConfigured   = errors.New("gemini is not configured. Please set GEMINI_API_KEY environment variable")
	ErrGeminiAllModelsFailed = errors.New("all gemini models failed")
)

type GeminiProvider struct {
	client *genai.Client
	models []string
}

func NewGeminiProvider(ctx context.Context) (*GeminiProvider, error) {
	if !config.IsGeminiConfigured() {
		return nil, ErrGeminiNotConfigured
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(config.GetGeminiAPIKey()))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiProvider{
		client: client,
		models: []string{"gemini-3.8-flash", "gemini-3.7-flash"},
	}, nil
}

// Name returns the provider name.
func (c *GeminiProvider) Name() string {
	return "gemini"
}

// IsAvailable checks if the provider is available.
func (c *GeminiProvider) IsAvailable() bool {
	return c.client != nil
}

// CompleteText implements the Provider interface.
func (c *GeminiProvider) CompleteText(ctx context.Context, prompt string) (string, error) {
	return c.CompleteTextWithOptions(ctx, prompt, CompletionOptions{})
}

func (c *GeminiProvider) CompleteTextWithOptions(ctx context.Context, prompt string, opts CompletionOptions) (string, error) {
	var lastErr error

	for _, modelName := range c.models {
		logger.Ctx(ctx).Info("Trying Gemini model", logger.String("model", modelName))
		model := c.client.GenerativeModel(modelName)
		applyGeminiCompletionOptions(model, opts)

		resp, err := model.GenerateContent(ctx, geminiParts(ctx, prompt)...)
		if err != nil {
			logger.Ctx(ctx).Error("Gemini model failed",
				logger.String("model", modelName),
				logger.Error(err))
			lastErr = err

			continue
		}

		if len(resp.Candidates) == 0 {
			logger.Ctx(ctx).Warn("Gemini model returned no candidates", logger.String("model", modelName))

			continue
		}

		candidate := resp.Candidates[0]
		if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
			logger.Ctx(ctx).Warn("Gemini model returned empty content", logger.String("model", modelName))

			continue
		}

		part := candidate.Content.Parts[0]
		if textPart, ok := part.(genai.Text); ok {
			logger.Ctx(ctx).Info("Gemini model succeeded", logger.String("model", modelName))

			return string(textPart), nil
		}

		logger.Ctx(ctx).Warn("Gemini model returned non-text content", logger.String("model", modelName))
	}

	if lastErr != nil {
		return "", fmt.Errorf("all Gemini models failed: %w", lastErr)
	}

	return "", ErrGeminiAllModelsFailed
}

func applyGeminiCompletionOptions(model *genai.GenerativeModel, opts CompletionOptions) {
	if opts.Temperature == nil && !opts.JSONResponse {
		return
	}

	maxTokens := int32(config.AIMaxOutputTokens)
	cfg := genai.GenerationConfig{
		MaxOutputTokens: &maxTokens,
	}
	if opts.JSONResponse {
		cfg.ResponseMIMEType = "application/json"
	}
	model.GenerationConfig = cfg
}

// CompleteTextStream implements the Provider interface.
func (c *GeminiProvider) CompleteTextStream(ctx context.Context, prompt string) (<-chan StreamToken, error) {
	tokenChan := make(chan StreamToken, 100)

	go func() {
		defer close(tokenChan)
		defer func() {
			if r := recover(); r != nil {
				logger.Global().Error("panic in Gemini stream goroutine", logger.Any("panic", r))
				select {
				case tokenChan <- StreamToken{Error: "internal streaming error", Done: true}:
				case <-ctx.Done():
				}
			}
		}()

		c.processGeminiStream(ctx, prompt, tokenChan)
	}()

	return tokenChan, nil
}

func (c *GeminiProvider) processGeminiStream(ctx context.Context, prompt string, tokenChan chan<- StreamToken) {
	for _, modelName := range c.models {
		logger.Ctx(ctx).Info("Trying Gemini stream model", logger.String("model", modelName))

		if c.tryModelStream(ctx, prompt, modelName, tokenChan) {
			logger.Ctx(ctx).Info("Gemini stream model succeeded", logger.String("model", modelName))

			return // Success.
		}

		logger.Ctx(ctx).Error("Gemini stream model failed", logger.String("model", modelName))
	}

	logger.Ctx(ctx).Error("All Gemini stream models failed")

	select {
	case tokenChan <- StreamToken{Error: "all models failed", Done: true}:
	case <-ctx.Done():
	}
}

func (c *GeminiProvider) tryModelStream(ctx context.Context, prompt, modelName string, tokenChan chan<- StreamToken) bool {
	model := c.client.GenerativeModel(modelName)

	maxTokens := int32(config.AIMaxOutputTokens)
	model.GenerationConfig = genai.GenerationConfig{
		MaxOutputTokens: &maxTokens,
	}

	iter := model.GenerateContentStream(ctx, geminiParts(ctx, prompt)...)
	if iter == nil {
		logger.Ctx(ctx).Error("Gemini stream model returned nil iterator", logger.String("model", modelName))

		return false
	}

	return c.processModelIteration(ctx, iter, tokenChan)
}

func (c *GeminiProvider) processModelIteration(ctx context.Context, iter *genai.GenerateContentResponseIterator, tokenChan chan<- StreamToken) bool {
	emittedText := false
	for {
		resp, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				select {
				case tokenChan <- StreamToken{Done: true}:
					return true
				case <-ctx.Done():
					return false
				}
			}

			logger.Ctx(ctx).Error("Gemini stream iterator failed", logger.Error(err))
			if emittedText {
				select {
				case tokenChan <- StreamToken{Error: "Gemini stopped while generating a response", Done: true}:
				case <-ctx.Done():
				}
				return true
			}
			return false
		}

		if resp == nil {
			logger.Ctx(ctx).Error("Gemini stream model returned an empty response")
			return false
		}

		if !c.processResponse(ctx, resp, tokenChan) {
			return false
		}
		emittedText = true
	}
}

// processResponse emits every text part in the chunk. Gemini streams
// interleave non-text parts (thought summaries, usage metadata) and can
// deliver several parts per chunk; treating one of them as a failure
// abandoned the model mid-answer and restarted the next one.
func (c *GeminiProvider) processResponse(ctx context.Context, resp *genai.GenerateContentResponse, tokenChan chan<- StreamToken) bool {
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return true
	}

	for _, part := range resp.Candidates[0].Content.Parts {
		textPart, ok := part.(genai.Text)
		if !ok {
			continue
		}
		text := string(textPart)
		if strings.TrimSpace(text) == "" {
			continue
		}

		logger.Global().Debug("Gemini returned text chunk", logger.Int("length", len(text)))

		select {
		case tokenChan <- StreamToken{Text: text, Done: false}:
		case <-ctx.Done():
			return false
		}
	}

	return true
}

// Close implements the Provider interface.
func (c *GeminiProvider) Close() error {
	return c.client.Close()
}

func geminiParts(ctx context.Context, prompt string) []genai.Part {
	parts := []genai.Part{genai.Text(prompt)}
	for _, input := range imagesFromContext(ctx) {
		mime, data, err := DecodeImage(input)
		if err == nil {
			parts = append(parts, genai.Blob{MIMEType: mime, Data: data})
		}
	}
	return parts
}
