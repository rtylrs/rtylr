package snips

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

var (
	ErrGeminiNotConfigured   = errors.New("gemini is not configured. Please set GEMINI_API_KEY environment variable")
	ErrGeminiAllModelsFailed = errors.New("all gemini models failed")
)

const geminiRequestTimeout = 35 * time.Second

type GeminiProvider struct {
	client     *genai.Client
	httpClient *http.Client
	models     []string
}

func NewGeminiProvider(ctx context.Context) (*GeminiProvider, error) {
	if !config.IsGeminiConfigured() {
		return nil, ErrGeminiNotConfigured
	}

	httpClient := &http.Client{Timeout: geminiRequestTimeout}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:     config.GetGeminiAPIKey(),
		Backend:    genai.BackendGeminiAPI,
		HTTPClient: httpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiProvider{
		client:     client,
		httpClient: httpClient,
		models:     []string{"gemini-3.8-flash", "gemini-3.7-flash"},
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
	requestCtx, cancel := context.WithTimeout(ctx, geminiRequestTimeout)
	defer cancel()
	contents := geminiContents(ctx, prompt)
	var lastErr error

	for _, modelName := range c.models {
		if requestCtx.Err() != nil {
			lastErr = requestCtx.Err()
			break
		}
		logger.Ctx(requestCtx).Info("Trying Gemini model", logger.String("model", modelName))
		resp, err := c.client.Models.GenerateContent(requestCtx, modelName, contents, geminiGenerationConfig(opts))
		if err != nil {
			logger.Ctx(requestCtx).Error("Gemini model failed",
				logger.String("model", modelName),
				logger.Error(err))
			lastErr = err
			continue
		}

		if text := geminiResponseText(resp); text != "" {
			logger.Ctx(requestCtx).Info("Gemini model succeeded", logger.String("model", modelName))
			return text, nil
		}
		logger.Ctx(requestCtx).Warn("Gemini model returned no text", logger.String("model", modelName))
	}

	if lastErr != nil {
		return "", fmt.Errorf("all Gemini models failed: %w", lastErr)
	}
	return "", ErrGeminiAllModelsFailed
}

func geminiGenerationConfig(opts CompletionOptions) *genai.GenerateContentConfig {
	result := &genai.GenerateContentConfig{
		ThinkingConfig: &genai.ThinkingConfig{
			ThinkingLevel: genai.ThinkingLevelLow,
		},
	}
	if opts.Temperature != nil || opts.JSONResponse {
		result.MaxOutputTokens = int32(config.AIMaxOutputTokens)
	}
	if opts.JSONResponse {
		result.ResponseMIMEType = "application/json"
	}
	if opts.Temperature != nil {
		temperature := float32(*opts.Temperature)
		result.Temperature = &temperature
	}
	return result
}

func geminiStreamGenerationConfig() *genai.GenerateContentConfig {
	result := geminiGenerationConfig(CompletionOptions{})
	result.MaxOutputTokens = int32(config.AIMaxOutputTokens)
	return result
}

func geminiContents(ctx context.Context, prompt string) []*genai.Content {
	parts := []*genai.Part{genai.NewPartFromText(prompt)}
	for _, input := range imagesFromContext(ctx) {
		mime, data, err := DecodeImage(input)
		if err == nil {
			parts = append(parts, genai.NewPartFromBytes(data, mime))
		}
	}
	return []*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)}
}

func geminiResponseText(response *genai.GenerateContentResponse) string {
	if response == nil || len(response.Candidates) == 0 || response.Candidates[0].Content == nil {
		return ""
	}
	var text strings.Builder
	for _, part := range response.Candidates[0].Content.Parts {
		if part != nil && part.Text != "" && !part.Thought {
			text.WriteString(part.Text)
		}
	}
	return strings.TrimSpace(text.String())
}

// CompleteTextStream implements the Provider interface.
func (c *GeminiProvider) CompleteTextStream(ctx context.Context, prompt string) (<-chan StreamToken, error) {
	tokenChan := make(chan StreamToken, 100)
	streamCtx, cancel := context.WithTimeout(ctx, geminiRequestTimeout)

	go func() {
		defer close(tokenChan)
		defer cancel()
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Global().Error("panic in Gemini stream goroutine", logger.Any("panic", recovered))
				_ = sendGeminiToken(streamCtx, tokenChan, StreamToken{Error: "internal streaming error", Done: true})
			}
		}()

		c.processGeminiStream(streamCtx, prompt, tokenChan)
	}()

	return tokenChan, nil
}

func (c *GeminiProvider) processGeminiStream(ctx context.Context, prompt string, tokenChan chan<- StreamToken) {
	contents := geminiContents(ctx, prompt)
	for _, modelName := range c.models {
		if ctx.Err() != nil {
			break
		}
		logger.Ctx(ctx).Info("Trying Gemini stream model", logger.String("model", modelName))
		if c.tryModelStream(ctx, contents, modelName, tokenChan) {
			logger.Ctx(ctx).Info("Gemini stream model succeeded", logger.String("model", modelName))
			return
		}
		logger.Ctx(ctx).Error("Gemini stream model failed", logger.String("model", modelName))
	}

	logger.Ctx(ctx).Error("All Gemini stream models failed")
	if ctx.Err() != nil {
		_ = sendGeminiToken(ctx, tokenChan, StreamToken{Error: "Gemini request timed out", Done: true})
		return
	}
	_ = sendGeminiToken(ctx, tokenChan, StreamToken{Error: "all models failed", Done: true})
}

func (c *GeminiProvider) tryModelStream(ctx context.Context, contents []*genai.Content, modelName string, tokenChan chan<- StreamToken) bool {
	return c.processModelIteration(ctx, modelName, contents, tokenChan)
}

func (c *GeminiProvider) processModelIteration(ctx context.Context, modelName string, contents []*genai.Content, tokenChan chan<- StreamToken) bool {
	emittedText := false
	for response, err := range c.client.Models.GenerateContentStream(ctx, modelName, contents, geminiStreamGenerationConfig()) {
		if err != nil {
			logger.Ctx(ctx).Error("Gemini stream iterator failed", logger.Error(err))
			if emittedText {
				_ = sendGeminiToken(ctx, tokenChan, StreamToken{Error: "Gemini stopped while generating a response", Done: true})
				return true
			}
			return false
		}
		if response == nil {
			logger.Ctx(ctx).Error("Gemini stream model returned an empty response")
			return false
		}
		if geminiResponseHasText(response) {
			emittedText = true
		}
		if !c.processResponse(ctx, response, tokenChan) {
			return false
		}
	}

	if ctx.Err() != nil {
		return false
	}
	return sendGeminiToken(ctx, tokenChan, StreamToken{Done: true})
}

func geminiResponseHasText(response *genai.GenerateContentResponse) bool {
	if response == nil || len(response.Candidates) == 0 || response.Candidates[0].Content == nil {
		return false
	}
	for _, part := range response.Candidates[0].Content.Parts {
		if part != nil && !part.Thought && strings.TrimSpace(part.Text) != "" {
			return true
		}
	}
	return false
}

// processResponse emits every non-thought text part in the chunk.
func (c *GeminiProvider) processResponse(ctx context.Context, response *genai.GenerateContentResponse, tokenChan chan<- StreamToken) bool {
	if response == nil || len(response.Candidates) == 0 || response.Candidates[0].Content == nil {
		return true
	}
	for _, part := range response.Candidates[0].Content.Parts {
		if part == nil || part.Thought || strings.TrimSpace(part.Text) == "" {
			continue
		}
		logger.Global().Debug("Gemini returned text chunk", logger.Int("length", len(part.Text)))
		if !sendGeminiToken(ctx, tokenChan, StreamToken{Text: part.Text}) {
			return false
		}
	}
	return true
}

func sendGeminiToken(ctx context.Context, tokenChan chan<- StreamToken, token StreamToken) bool {
	select {
	case tokenChan <- token:
		return true
	case <-ctx.Done():
		return false
	}
}

// Close implements the Provider interface.
func (c *GeminiProvider) Close() error {
	if c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}
	return nil
}
