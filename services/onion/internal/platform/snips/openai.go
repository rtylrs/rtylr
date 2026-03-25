package snips

import (
	"context"
	"errors"
	"fmt"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

var (
	ErrOpenAINotConfigured = errors.New("OpenAI is not configured. Please set OPENAI_API_KEY environment variable")
	ErrOpenAINoChoices     = errors.New("no response choices received from OpenAI")
)

type OpenAIProvider struct {
	client    openai.Client
	available bool
}

func NewOpenAIProvider() (*OpenAIProvider, error) {
	if !config.IsOpenAIConfigured() {
		return nil, ErrOpenAINotConfigured
	}

	var opts []option.RequestOption

	opts = append(opts, option.WithAPIKey(config.GetOpenAIAPIKey()))

	if orgID := config.GetOpenAIOrgID(); orgID != "" {
		opts = append(opts, option.WithOrganization(orgID))
	}

	if baseURL := config.GetOpenAIBaseURL(); baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	client := openai.NewClient(opts...)

	return &OpenAIProvider{
		client:    client,
		available: true,
	}, nil
}

// Name returns the provider name.
func (c *OpenAIProvider) Name() string {
	return "openai"
}

// IsAvailable checks if the provider is available.
func (c *OpenAIProvider) IsAvailable() bool {
	return c.available
}

// CompleteText implements the Provider interface.
func (c *OpenAIProvider) CompleteText(ctx context.Context, prompt string) (string, error) {
	req := openai.ChatCompletionNewParams{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	}

	resp, err := c.client.Chat.Completions.New(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", ErrOpenAINoChoices
	}

	return resp.Choices[0].Message.Content, nil
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
	req := openai.ChatCompletionNewParams{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	}

	stream := c.client.Chat.Completions.NewStreaming(ctx, req)
	if stream == nil {
		c.sendErrorToken(tokenChan, "failed to create streaming response")

		return
	}

	defer func() {
		if err := stream.Close(); err != nil {
			logger.Global().Error("Failed to close stream", logger.Error(err))
		}
	}()

	for stream.Next() {
		chunk := stream.Current()
		if !c.processChunk(ctx, chunk, tokenChan) {
			return
		}

		if c.isStreamComplete(chunk) {
			break
		}
	}

	if err := stream.Err(); err != nil {
		c.sendErrorToken(tokenChan, err.Error())

		return
	}

	c.sendCompletionToken(tokenChan)
}

func (c *OpenAIProvider) processChunk(_ context.Context, chunk openai.ChatCompletionChunk, tokenChan chan<- StreamToken) bool {
	for i := range chunk.Choices {
		choice := &chunk.Choices[i]
		if choice.Delta.Content != "" {
			if !c.sendTextToken(tokenChan, choice.Delta.Content) {
				return false
			}
		}
	}

	return true
}

func (c *OpenAIProvider) isStreamComplete(chunk openai.ChatCompletionChunk) bool {
	return len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason == "stop"
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

func (c *OpenAIProvider) sendCompletionToken(tokenChan chan<- StreamToken) {
	select {
	case tokenChan <- StreamToken{Done: true}:
	default:
	}
}

// Close implements the Provider interface.
func (c *OpenAIProvider) Close() error {
	// OpenAI client doesn't need explicit closing
	return nil
}
