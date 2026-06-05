package snips

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

var (
	ErrOpenAINotConfigured   = errors.New("OpenAI is not configured. Please set OPENAI_API_KEY environment variable")
	ErrDeepSeekNotConfigured = errors.New("DeepSeek is not configured. Please set DEEPSEEK_API_KEY environment variable")
	ErrOpenAINoChoices       = errors.New("no response choices received from provider")
)

// OpenAIProvider speaks the OpenAI chat-completions API. DeepSeek is wire
// compatible, so the same provider drives it by pointing at a different base URL
// and model — the only differences are name/model/base URL, all injected here.
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

func (c *OpenAIProvider) Name() string      { return c.name }
func (c *OpenAIProvider) IsAvailable() bool { return c.available }

func (c *OpenAIProvider) chatParams(prompt string, stream bool) openai.ChatCompletionNewParams {
	req := openai.ChatCompletionNewParams{
		Model:     c.model,
		MaxTokens: openai.Int(int64(config.AIMaxOutputTokens)),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
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
	var lastErr error
	for attempt := 0; attempt <= config.AIMaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}

		callCtx, cancel := context.WithTimeout(ctx, config.AIRequestTimeout)
		resp, err := c.client.Chat.Completions.New(callCtx, c.chatParams(prompt, false))
		cancel()
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

// CompleteTextStream implements the Provider interface.
func (c *OpenAIProvider) CompleteTextStream(ctx context.Context, prompt string) (<-chan StreamToken, error) {
	tokenChan := make(chan StreamToken, 100)

	go func() {
		defer close(tokenChan)
		c.processOpenAIStream(ctx, prompt, tokenChan)
	}()

	return tokenChan, nil
}

func (c *OpenAIProvider) processOpenAIStream(ctx context.Context, prompt string, tokenChan chan<- StreamToken) {
	// Bound the whole stream so a hung provider can't pin the connection open.
	ctx, cancel := context.WithTimeout(ctx, config.AIRequestTimeout)
	defer cancel()

	stream := c.client.Chat.Completions.NewStreaming(ctx, c.chatParams(prompt, true))
	if stream == nil {
		c.sendErrorToken(tokenChan, "failed to create streaming response")
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
				if !c.sendTextToken(tokenChan, content) {
					return
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		c.sendErrorToken(tokenChan, err.Error())
		return
	}

	c.sendCompletionToken(tokenChan, int(promptTokens), int(completionTokens))
}

func (c *OpenAIProvider) sendTextToken(tokenChan chan<- StreamToken, content string) bool {
	select {
	case tokenChan <- StreamToken{Text: content, Done: false}:
		return true
	default:
		return false
	}
}

func (c *OpenAIProvider) sendErrorToken(tokenChan chan<- StreamToken, errorMsg string) {
	select {
	case tokenChan <- StreamToken{Error: errorMsg, Done: true}:
	default:
	}
}

func (c *OpenAIProvider) sendCompletionToken(tokenChan chan<- StreamToken, promptTokens, completionTokens int) {
	select {
	case tokenChan <- StreamToken{Done: true, PromptTokens: promptTokens, CompletionTokens: completionTokens}:
	default:
	}
}

// Close implements the Provider interface.
func (c *OpenAIProvider) Close() error {
	return nil
}
