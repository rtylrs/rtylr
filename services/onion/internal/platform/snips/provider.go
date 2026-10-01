package snips

import "context"

// CompletionOptions configures non-streaming provider calls used for internal
// planning steps. Streaming user-facing chat keeps the default provider style.
type CompletionOptions struct {
	Temperature  *float64
	JSONResponse bool
}

func DeterministicJSONCompletionOptions() CompletionOptions {
	temperature := 0.0
	return CompletionOptions{
		Temperature:  &temperature,
		JSONResponse: true,
	}
}

// Provider represents any AI provider (LLM, database, custom models, etc.)
type Provider interface {
	Name() string
	IsAvailable() bool
	CompleteText(ctx context.Context, prompt string) (string, error)
	CompleteTextStream(ctx context.Context, prompt string) (<-chan StreamToken, error)
	Close() error
}

// ProviderCapabilities defines what a provider can do.
type ProviderCapabilities struct {
	SupportsStreaming  bool // For real-time streaming responses
	SupportsCompletion bool // For text completion
	SupportsEmbeddings bool // For vector embeddings
	SupportsDatabase   bool // For MindsDB-style database queries
	SupportsFineTuning bool // For custom model training
	IsCustomModel      bool // For endpoints outside the built-in provider catalog
}
