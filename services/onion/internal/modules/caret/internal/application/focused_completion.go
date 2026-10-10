package application

import (
	"context"

	"github.com/voxire/onion/internal/platform/snips"
)

type optionsCompleter interface {
	CompleteTextWithOptions(ctx context.Context, prompt string, opts snips.CompletionOptions) (string, error)
}

// completeFocused runs a completion at a low temperature when the AI
// supports it: research and learning passes want consistency, not variety.
func completeFocused(ctx context.Context, ai TextCompleter, prompt string, temperature float64) (string, error) {
	if withOptions, ok := ai.(optionsCompleter); ok {
		return withOptions.CompleteTextWithOptions(ctx, prompt, snips.CompletionOptions{Temperature: &temperature})
	}
	return ai.CompleteText(ctx, prompt)
}
