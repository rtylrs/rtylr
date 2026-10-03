package snips

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/shared"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

var (
	ErrOpenAINotConfigured    = errors.New("OpenAI is not configured. Please set OPENAI_API_KEY environment variable")
	ErrDeepSeekNotConfigured  = errors.New("DeepSeek is not configured. Please set DEEPSEEK_API_KEY environment variable")
	ErrNvidiaNIMNotConfigured = errors.New("NVIDIA NIM is not configured. Please set NVIDIA_NIM_API_KEY environment variable")
	ErrOpenCodeNotConfigured  = errors.New("OpenCode Zen is not configured. Please set OPENCODE_API_KEY environment variable")
	ErrOpenAINoChoices        = errors.New("no response choices received from provider")
)

// OpenAIProvider speaks the OpenAI chat-completions API, also used by DeepSeek,
// NVIDIA NIM and OpenCode Zen with provider-specific base URLs and models.
type OpenAIProvider struct {
	client    openai.Client
	name      string
	model     string
	available bool
}

func newOpenAICompatProvider(name, apiKey, baseURL, model, orgID string) *OpenAIProvider {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if orgID != "" {
		opts = append(opts, option.WithOrganization(orgID))
	}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	return &OpenAIProvider{
		client:    openai.NewClient(opts...),
		name:      name,
		model:     model,
		available: true,
	}
}

// NewOpenAIProvider builds the OpenAI fallback provider.
func NewOpenAIProvider() (*OpenAIProvider, error) {
	if !config.IsOpenAIConfigured() {
		return nil, ErrOpenAINotConfigured
	}
	return newOpenAICompatProvider(
		"openai",
		config.GetOpenAIAPIKey(),
		config.GetOpenAIBaseURL(),
		"gpt-4o-mini",
		config.GetOpenAIOrgID(),
	), nil
}

// NewDeepSeekProvider builds the primary DeepSeek provider (OpenAI-compatible).
func NewDeepSeekProvider() (*OpenAIProvider, error) {
	if !config.IsDeepSeekConfigured() {
		return nil, ErrDeepSeekNotConfigured
	}
	return newOpenAICompatProvider(
		"deepseek",
		config.GetDeepSeekAPIKey(),
		config.GetDeepSeekBaseURL(),
		config.GetDeepSeekModel(),
		"",
	), nil
}

// NewNvidiaNIMProvider builds an NVIDIA NIM provider (OpenAI-compatible).
func NewNvidiaNIMProvider() (*OpenAIProvider, error) {
	if !config.IsNvidiaNIMConfigured() {
		return nil, ErrNvidiaNIMNotConfigured
	}
	return newOpenAICompatProvider(
		"nvidia",
		config.GetNvidiaNIMAPIKey(),
		config.GetNvidiaNIMBaseURL(),
		config.GetNvidiaNIMModel(),
		"",
	), nil
}

// NewOpenCodeProvider builds the final OpenCode Zen fallback provider.
func NewOpenCodeProvider() (*OpenAIProvider, error) {
	if !config.IsOpenCodeConfigured() {
		return nil, ErrOpenCodeNotConfigured
	}
	return newOpenAICompatProvider(
		"opencode",
		config.GetOpenCodeAPIKey(),
		config.GetOpenCodeBaseURL(),
		config.GetOpenCodeModel(),
		"",
	), nil
}

func (c *OpenAIProvider) Name() string      { return c.name }
func (c *OpenAIProvider) IsAvailable() bool { return c.available }

func (c *OpenAIProvider) chatParams(prompt string, stream bool, opts CompletionOptions) openai.ChatCompletionNewParams {
	req := openai.ChatCompletionNewParams{
		Model:     c.model,
		MaxTokens: openai.Int(int64(config.AIMaxOutputTokens)),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	}
	if c.name == "nvidia" {
		// Reserve the output budget for the visible answer, not model reasoning.
		req.SetExtraFields(map[string]any{"chat_template_kwargs": map[string]any{"enable_thinking": false}})
	}
	if opts.Temperature != nil {
		req.Temperature = openai.Float(*opts.Temperature)
	}
	if opts.JSONResponse {
		jsonFormat := shared.NewResponseFormatJSONObjectParam()
		req.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &jsonFormat,
		}
	}
	if stream {
		// Ask the provider to include token usage in the final stream chunk so we
		// can meter real usage for quotas.
		req.StreamOptions = openai.ChatCompletionStreamOptionsParam{IncludeUsage: openai.Bool(true)}
	}
	return req
}

// CompleteText implements the Provider interface with a bounded timeout and
// retry-with-backoff on transient failures.
func (c *OpenAIProvider) CompleteText(ctx context.Context, prompt string) (string, error) {
	return c.CompleteTextWithOptions(ctx, prompt, CompletionOptions{})
}

func (c *OpenAIProvider) CompleteTextWithOptions(ctx context.Context, prompt string, opts CompletionOptions) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= config.AIMaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}

		resp, err := c.completeTextOnce(ctx, prompt, opts)
		if err != nil && opts.JSONResponse {
			fallbackOpts := opts
			fallbackOpts.JSONResponse = false
			resp, err = c.completeTextOnce(ctx, prompt, fallbackOpts)
		}
		if err != nil {
			lastErr = fmt.Errorf("failed to create chat completion: %w", err)
			continue
		}
		if len(resp.Choices) == 0 {
			return "", ErrOpenAINoChoices
		}
		return resp.Choices[0].Message.Content, nil
	}

	return "", lastErr
}

func (c *OpenAIProvider) completeTextOnce(ctx context.Context, prompt string, opts CompletionOptions) (*openai.ChatCompletion, error) {
	callCtx, cancel := context.WithTimeout(ctx, config.AIRequestTimeout)
	defer cancel()
	return c.client.Chat.Completions.New(callCtx, c.chatParams(prompt, false, opts))
}

// CompleteTextStream implements the Provider interface.
func (c *OpenAIProvider) CompleteTextStream(ctx context.Context, prompt string) (<-chan StreamToken, error) {
	tokenChan := make(chan StreamToken, 100)

	go func() {
		defer close(tokenChan)
		// Never let a provider/SDK panic escape this goroutine — an unrecovered
		// goroutine panic crashes the whole API process (the HTTP recover
		// middleware can't catch it). Surface it as a stream error instead.
		defer func() {
			if r := recover(); r != nil {
				logger.Global().Error("panic in AI stream goroutine", logger.Any("panic", r))
				c.sendErrorToken(ctx, tokenChan, "internal streaming error")
			}
		}()
		c.processOpenAIStream(ctx, prompt, tokenChan)
	}()

	return tokenChan, nil
}

func (c *OpenAIProvider) processOpenAIStream(ctx context.Context, prompt string, tokenChan chan<- StreamToken) {
	// Bound the whole stream so a hung provider can't pin the connection open.
	ctx, cancel := context.WithTimeout(ctx, config.AIRequestTimeout)
	defer cancel()

	stream := c.client.Chat.Completions.NewStreaming(ctx, c.chatParams(prompt, true, CompletionOptions{}))
	if stream == nil {
		c.sendErrorToken(ctx, tokenChan, "failed to create streaming response")
		return
	}

	defer func() {
		if err := stream.Close(); err != nil {
			logger.Global().Error("Failed to close stream", logger.Error(err))
		}
	}()

	var promptTokens, completionTokens int64
	for stream.Next() {
		chunk := stream.Current()

		// The usage-bearing chunk arrives last (after FinishReason=stop) and has
		// no choices, so capture usage whenever present and keep reading.
		if chunk.Usage.TotalTokens > 0 {
			promptTokens = chunk.Usage.PromptTokens
			completionTokens = chunk.Usage.CompletionTokens
		}

		for i := range chunk.Choices {
			if content := chunk.Choices[i].Delta.Content; content != "" {
				if !c.sendTextToken(ctx, tokenChan, content) {
					return
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		c.sendErrorToken(ctx, tokenChan, err.Error())
		return
	}

	c.sendCompletionToken(ctx, tokenChan, int(promptTokens), int(completionTokens))
}

func (c *OpenAIProvider) sendTextToken(ctx context.Context, tokenChan chan<- StreamToken, content string) bool {
	select {
	case tokenChan <- StreamToken{Text: content, Done: false}:
		return true
	case <-ctx.Done():
		return false
	}
}

func (c *OpenAIProvider) sendErrorToken(ctx context.Context, tokenChan chan<- StreamToken, errorMsg string) bool {
	select {
	case tokenChan <- StreamToken{Error: errorMsg, Done: true}:
		return true
	case <-ctx.Done():
		return false
	}
}

func (c *OpenAIProvider) sendCompletionToken(ctx context.Context, tokenChan chan<- StreamToken, promptTokens, completionTokens int) bool {
	select {
	case tokenChan <- StreamToken{Done: true, PromptTokens: promptTokens, CompletionTokens: completionTokens}:
		return true
	case <-ctx.Done():
		return false
	}
}

// Close implements the Provider interface.
func (c *OpenAIProvider) Close() error {
	return nil
}
