package snips

import (
	"context"
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/config"
)

// backoff returns the delay before retry attempt n (1-based): 200ms, 400ms, …,
// capped at 2s.
func backoff(attempt int) time.Duration {
	d := 200 * time.Millisecond << (attempt - 1)
	if d > 2*time.Second {
		return 2 * time.Second
	}
	return d
}

var (
	ErrNoProvidersConfigured = errors.New("no AI providers configured")
	ErrAllProvidersFailed    = errors.New("all AI providers failed")
)

type Client struct {
	providers       []Provider
	primaryProvider Provider
}

type StreamToken struct {
	Text     string `json:"text"`
	Done     bool   `json:"done"`
	Provider string `json:"provider,omitempty"`
	Error    string `json:"error,omitempty"`
	// Token usage, populated on the final (Done) token when the provider reports
	// it. Used by the quota system to meter real usage rather than estimates.
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
}

func NewClient(ctx context.Context) (*Client, error) {
	if !config.IsAnyAIConfigured() {
		return nil, ErrNoProvidersConfigured
	}

	client := &Client{
		providers: make([]Provider, 0),
	}

	// Default priority: DeepSeek, then NVIDIA NIM (both OpenAI-compatible), then
	// Gemini and OpenAI as fallbacks. AI_PRIMARY_PROVIDER can promote deepseek or
	// nvidia to the front when both are configured. Any unconfigured provider is
	// skipped; the first registered becomes primary (used for streaming).
	register := func(p Provider) {
		client.providers = append(client.providers, p)
		if client.primaryProvider == nil {
			client.primaryProvider = p
		}
	}

	deepseekFirst := config.GetAIPrimaryProvider() != "nvidia"
	addDeepSeek := func() {
		if config.IsDeepSeekConfigured() {
			if p, err := NewDeepSeekProvider(); err == nil {
				register(p)
			}
		}
	}
	addNvidia := func() {
		if config.IsNvidiaNIMConfigured() {
			if p, err := NewNvidiaNIMProvider(); err == nil {
				register(p)
			}
		}
	}

	if deepseekFirst {
		addDeepSeek()
		addNvidia()
	} else {
		addNvidia()
		addDeepSeek()
	}

	if config.IsGeminiConfigured() {
		if p, err := NewGeminiProvider(ctx); err == nil {
			register(p)
		}
	}

	if config.IsOpenAIConfigured() {
		if p, err := NewOpenAIProvider(); err == nil {
			register(p)
		}
	}

	if len(client.providers) == 0 {
		return nil, ErrNoProvidersConfigured
	}

	return client, nil
}

func (c *Client) CompleteText(ctx context.Context, prompt string) (string, error) {
	// Try each provider in order until one succeeds
	for _, provider := range c.providers {
		if !provider.IsAvailable() {
			continue
		}

		result, err := provider.CompleteText(ctx, prompt)
		if err == nil {
			return result, nil
		}
	}

	return "", ErrAllProvidersFailed
}

func (c *Client) CompleteTextStream(ctx context.Context, prompt string) (<-chan StreamToken, error) {
	// Use primary provider for streaming
	if c.primaryProvider != nil && c.primaryProvider.IsAvailable() {
		tokenChan, err := c.primaryProvider.CompleteTextStream(ctx, prompt)
		if err == nil {
			return enhanceStreamWithProvider(ctx, c.primaryProvider.Name(), tokenChan), nil
		}
	}

	// Fallback to first available provider
	for _, provider := range c.providers {
		if provider.IsAvailable() {
			tokenChan, err := provider.CompleteTextStream(ctx, prompt)
			if err == nil {
				return enhanceStreamWithProvider(ctx, provider.Name(), tokenChan), nil
			}
		}
	}

	return nil, ErrAllProvidersFailed
}

func enhanceStreamWithProvider(ctx context.Context, providerName string, tokenChan <-chan StreamToken) <-chan StreamToken {
	enhancedChan := make(chan StreamToken, 100)

	go func() {
		defer close(enhancedChan)

		for {
			select {
			case <-ctx.Done():
				return
			case token, ok := <-tokenChan:
				if !ok {
					return
				}

				token.Provider = providerName
				select {
				case enhancedChan <- token:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return enhancedChan
}

func (c *Client) GetProviderStatus() map[string]bool {
	status := make(map[string]bool)

	for _, provider := range c.providers {
		status[provider.Name()] = provider.IsAvailable()
	}

	return status
}

func (c *Client) GetPrimaryProvider() string {
	if c.primaryProvider != nil {
		return c.primaryProvider.Name()
	}

	return ""
}

func (c *Client) Close() error {
	var lastErr error

	for _, provider := range c.providers {
		if err := provider.Close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}
