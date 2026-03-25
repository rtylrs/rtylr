package snips

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/platform/config"
)

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
}

func NewClient(ctx context.Context) (*Client, error) {
	if !config.IsAnyAIConfigured() {
		return nil, ErrNoProvidersConfigured
	}

	client := &Client{
		providers: make([]Provider, 0),
	}

	// Register providers in priority order
	if config.IsGeminiConfigured() {
		if provider, err := NewGeminiProvider(ctx); err == nil {
			client.providers = append(client.providers, provider)
			if client.primaryProvider == nil {
				client.primaryProvider = provider
			}
		}
	}

	if config.IsOpenAIConfigured() {
		if provider, err := NewOpenAIProvider(); err == nil {
			client.providers = append(client.providers, provider)
			if client.primaryProvider == nil {
				client.primaryProvider = provider
			}
		}
	}

	// Future: MindsDB provider
	// if config.IsMindsDBConfigured() {
	//     if provider, err := NewMindsDBProvider(ctx); err == nil {
	//         client.providers = append(client.providers, provider)
	//         if client.primaryProvider == nil {
	//             client.primaryProvider = provider
	//         }
	//     }
	// }

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
			// Create enhanced channel with provider name
			enhancedChan := make(chan StreamToken, 100)

			go func() {
				defer close(enhancedChan)

				for token := range tokenChan {
					token.Provider = c.primaryProvider.Name()
					enhancedChan <- token
				}
			}()

			return enhancedChan, nil
		}
	}

	// Fallback to first available provider
	for _, provider := range c.providers {
		if provider.IsAvailable() {
			tokenChan, err := provider.CompleteTextStream(ctx, prompt)
			if err == nil {
				// Create enhanced channel with provider name
				enhancedChan := make(chan StreamToken, 100)

				go func() {
					defer close(enhancedChan)

					for token := range tokenChan {
						token.Provider = provider.Name()
						enhancedChan <- token
					}
				}()

				return enhancedChan, nil
			}
		}
	}

	return nil, ErrAllProvidersFailed
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
