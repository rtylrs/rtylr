package snips

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
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
	healthMu        sync.Mutex
	failures        map[string]int
	cooldownUntil   map[string]time.Time
}

const providerFailureCooldown = 10 * time.Second

type completionOptionsProvider interface {
	CompleteTextWithOptions(ctx context.Context, prompt string, opts CompletionOptions) (string, error)
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
		providers:     make([]Provider, 0),
		failures:      make(map[string]int),
		cooldownUntil: make(map[string]time.Time),
	}

	// Unconfigured providers are skipped; the first registered becomes primary.
	register := func(p Provider) {
		client.providers = append(client.providers, p)
		if client.primaryProvider == nil {
			client.primaryProvider = p
		}
	}

	if config.IsGeminiConfigured() {
		if p, err := NewGeminiProvider(ctx); err == nil {
			register(p)
		}
	}
	if config.IsOpenCodeConfigured() {
		if p, err := NewOpenCodeProvider(); err == nil {
			register(p)
		}
	}
	if config.IsDeepSeekConfigured() {
		if p, err := NewDeepSeekProvider(); err == nil {
			register(p)
		}
	}
	if config.IsOpenAIConfigured() {
		if p, err := NewOpenAIProvider(); err == nil {
			register(p)
		}
	}

	for i, provider := range client.providers {
		if provider.Name() == config.GetAIPrimaryProvider() {
			copy(client.providers[1:i+1], client.providers[:i])
			client.providers[0] = provider
			client.primaryProvider = provider
			break
		}
	}

	if len(client.providers) == 0 {
		return nil, ErrNoProvidersConfigured
	}

	return client, nil
}

func (c *Client) CompleteText(ctx context.Context, prompt string) (string, error) {
	return c.CompleteTextWithOptions(ctx, prompt, CompletionOptions{})
}

func (c *Client) CompleteTextWithOptions(ctx context.Context, prompt string, opts CompletionOptions) (string, error) {
	if c == nil || len(c.providers) == 0 {
		return "", ErrNoProvidersConfigured
	}
	ctx = ensureConversation(ctx)
	var lastErr error
	// Try each healthy provider in order until one succeeds. A short circuit
	// prevents a broken upstream from adding a full timeout to every request.
	for _, provider := range c.providers {
		if !c.providerReady(provider) || !providerSupportsInput(ctx, provider) {
			continue
		}

		if withOptions, ok := provider.(completionOptionsProvider); ok {
			result, err := withOptions.CompleteTextWithOptions(ctx, prompt, opts)
			if err == nil {
				c.markProviderSuccess(provider)
				return result, nil
			}
			lastErr = err
			c.markProviderFailure(ctx, provider, err)
			continue
		}

		result, err := provider.CompleteText(ctx, prompt)
		if err == nil {
			c.markProviderSuccess(provider)
			return result, nil
		}
		lastErr = err
		c.markProviderFailure(ctx, provider, err)
	}

	if lastErr != nil {
		return "", fmt.Errorf("%w: %w", ErrAllProvidersFailed, lastErr)
	}
	return "", ErrAllProvidersFailed
}

func (c *Client) CompleteTextStream(ctx context.Context, prompt string) (<-chan StreamToken, error) {
	if c == nil || len(c.providers) == 0 {
		return nil, ErrNoProvidersConfigured
	}
	ctx = ensureConversation(ctx)
	providers := c.streamingProviderOrder()
	var lastErr error
	for i, provider := range providers {
		if !c.providerReady(provider) || !providerSupportsInput(ctx, provider) {
			continue
		}
		tokenChan, err := provider.CompleteTextStream(ctx, prompt)
		if err == nil {
			c.markProviderSuccess(provider)
			return c.superviseProviderStream(ctx, prompt, provider, tokenChan, providers[i+1:]), nil
		}
		lastErr = err
		c.markProviderFailure(ctx, provider, err)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrAllProvidersFailed, lastErr)
	}
	return nil, ErrAllProvidersFailed
}

func (c *Client) streamingProviderOrder() []Provider {
	ordered := make([]Provider, 0, len(c.providers))
	if c.primaryProvider != nil {
		ordered = append(ordered, c.primaryProvider)
	}
	for _, provider := range c.providers {
		if provider == nil || provider == c.primaryProvider {
			continue
		}
		ordered = append(ordered, provider)
	}
	return ordered
}

// superviseProviderStream can fail over only before visible text has been
// delivered. Once a user has seen part of an answer, restarting another model
// would splice two unrelated answers together; the caller receives the
// provider error and can render a truthful partial-stream failure instead.
func (c *Client) superviseProviderStream(ctx context.Context, prompt string, provider Provider, source <-chan StreamToken, remaining []Provider) <-chan StreamToken {
	out := make(chan StreamToken, 100)
	go func() {
		defer close(out)
		visible := false
		current := provider
		currentSource := source
		for {
			select {
			case <-ctx.Done():
				return
			case token, ok := <-currentSource:
				if !ok {
					c.markProviderFailure(ctx, current, errors.New("stream closed before completion"))
					if !visible && len(remaining) > 0 {
						next, nextSource, err := c.nextProviderStream(ctx, prompt, remaining)
						if err == nil {
							current, currentSource, remaining = next, nextSource, remaining[1:]
							continue
						}
						c.markProviderFailure(ctx, remaining[0], err)
					}
					sendStreamToken(ctx, out, StreamToken{Done: true, Error: "provider stream ended before completion"})
					return
				}
				if token.Error != "" {
					c.markProviderFailure(ctx, current, errors.New(token.Error))
					if !visible && len(remaining) > 0 {
						next, nextSource, err := c.nextProviderStream(ctx, prompt, remaining)
						if err == nil {
							current, currentSource, remaining = next, nextSource, remaining[1:]
							continue
						}
						c.markProviderFailure(ctx, remaining[0], err)
					}
					token.Provider = current.Name()
					sendStreamToken(ctx, out, token)
					return
				}
				token.Provider = current.Name()
				if token.Text != "" {
					visible = true
				}
				if !sendStreamToken(ctx, out, token) {
					return
				}
				if token.Done {
					c.markProviderSuccess(current)
					return
				}
			}
		}
	}()
	return out
}

func (c *Client) nextProviderStream(ctx context.Context, prompt string, providers []Provider) (Provider, <-chan StreamToken, error) {
	var lastErr error
	for _, provider := range providers {
		if !c.providerReady(provider) || !providerSupportsInput(ctx, provider) {
			continue
		}
		stream, err := provider.CompleteTextStream(ctx, prompt)
		if err == nil {
			c.markProviderSuccess(provider)
			return provider, stream, nil
		}
		lastErr = err
		c.markProviderFailure(ctx, provider, err)
	}
	if lastErr == nil {
		lastErr = ErrAllProvidersFailed
	}
	return nil, nil, lastErr
}

func sendStreamToken(ctx context.Context, out chan<- StreamToken, token StreamToken) bool {
	select {
	case out <- token:
		return true
	case <-ctx.Done():
		return false
	}
}

func (c *Client) providerReady(provider Provider) bool {
	if provider == nil || !provider.IsAvailable() {
		return false
	}
	name := provider.Name()
	c.healthMu.Lock()
	defer c.healthMu.Unlock()
	if c.cooldownUntil == nil {
		c.cooldownUntil = make(map[string]time.Time)
	}
	return time.Now().After(c.cooldownUntil[name])
}

func (c *Client) markProviderSuccess(provider Provider) {
	if provider == nil {
		return
	}
	c.healthMu.Lock()
	defer c.healthMu.Unlock()
	if c.failures == nil {
		c.failures = make(map[string]int)
	}
	if c.cooldownUntil == nil {
		c.cooldownUntil = make(map[string]time.Time)
	}
	name := provider.Name()
	delete(c.failures, name)
	delete(c.cooldownUntil, name)
}

func (c *Client) markProviderFailure(ctx context.Context, provider Provider, err error) {
	if provider == nil {
		return
	}
	c.healthMu.Lock()
	defer c.healthMu.Unlock()
	if c.failures == nil {
		c.failures = make(map[string]int)
	}
	if c.cooldownUntil == nil {
		c.cooldownUntil = make(map[string]time.Time)
	}
	name := provider.Name()
	c.failures[name]++
	if c.failures[name] >= 2 {
		c.cooldownUntil[name] = time.Now().Add(providerFailureCooldown)
	}
	logger.Ctx(ctx).Warn("AI provider attempt failed",
		logger.String("provider", name),
		logger.Any("failure_count", c.failures[name]),
		logger.Error(err),
	)
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
	if c == nil {
		return status
	}

	for _, provider := range c.providers {
		status[provider.Name()] = c.providerReady(provider)
	}

	return status
}

func (c *Client) GetPrimaryProvider() string {
	if c != nil && c.primaryProvider != nil {
		return c.primaryProvider.Name()
	}

	return ""
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	var lastErr error

	for _, provider := range c.providers {
		if err := provider.Close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}
