package echo

import (
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/ai/classifier"
	"github.com/voxire/onion/internal/modules/caret/internal/application"
)

// NewFormMappingToolExecutor builds the Caret document-to-form mapping executor
// from a checkpoint path. Empty or unloadable paths fall back to the lexical
// matcher; the returned executor is always usable and the error describes
// the fallback.
func NewFormMappingToolExecutor(checkpointPath string) (application.ToolExecutor, error) {
	path := strings.TrimSpace(checkpointPath)
	if path == "" {
		return application.NewDocumentFormMappingToolExecutor(nil, map[string]any{
			"kind": "lexical",
		}), nil
	}
	scorer, err := classifier.Load(path)
	if err != nil {
		return application.NewDocumentFormMappingToolExecutor(nil, map[string]any{
			"kind":  "lexical",
			"error": classifier.CheckpointErrorKind(err),
		}), fmt.Errorf("form mapping checkpoint: %w", err)
	}
	cfg := scorer.Config()
	return application.NewDocumentFormMappingToolExecutor(scorer, map[string]any{
		"kind":          "s1-forms",
		"encoder":       cfg.Encoder,
		"width":         cfg.Width,
		"params":        scorer.Params(),
		"weights_bytes": scorer.WeightsBytes(),
	}), nil
}
